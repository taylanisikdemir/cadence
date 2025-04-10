// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cache

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig/dynamicproperties"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
)

var (
	// ErrCacheFull is returned if Put fails due to cache being filled with pinned elements
	ErrCacheFull = errors.New("Cache capacity is fully occupied with pinned elements")
)

// upper limit to prevent infinite growing
const cacheCountLimit = 1 << 25

// default size limit for size based cache if not defined (1 GB)
const cacheDefaultSizeLimit = 1 << 30

// lru is a concurrent fixed size cache that evicts elements in lru order
type (
	lru struct {
		mut           sync.Mutex
		byAccess      *list.List
		byKey         map[interface{}]*list.Element
		maxCount      int
		ttl           time.Duration
		pin           bool
		rmFunc        RemovedFunc
		sizeFunc      GetCacheItemSizeFunc
		maxSize       dynamicproperties.IntPropertyFn
		currSize      uint64
		sizeByKey     map[interface{}]uint64
		isSizeBased   bool
		activelyEvict bool
		// We use this instead of time.Now() in order to make testing easier
		timeSource   clock.TimeSource
		logger       log.Logger
		metricsScope metrics.Scope
	}

	iteratorImpl struct {
		lru        *lru
		createTime time.Time
		nextItem   *list.Element
	}

	entryImpl struct {
		key        interface{}
		createTime time.Time
		value      interface{}
		refCount   int
	}
)

// Close closes the iterator
func (it *iteratorImpl) Close() {
	it.lru.mut.Unlock()
}

// HasNext return true if there is more items to be returned
func (it *iteratorImpl) HasNext() bool {
	return it.nextItem != nil
}

// Next return the next item
func (it *iteratorImpl) Next() Entry {
	if it.nextItem == nil {
		panic("LRU cache iterator Next called when there is no next item")
	}

	entry := it.nextItem.Value.(*entryImpl)
	it.nextItem = it.nextItem.Next()
	// make a copy of the entry so there will be no concurrent access to this entry
	entry = &entryImpl{
		key:        entry.key,
		value:      entry.value,
		createTime: entry.createTime,
	}
	it.prepareNext()
	return entry
}

func (it *iteratorImpl) prepareNext() {
	for it.nextItem != nil {
		entry := it.nextItem.Value.(*entryImpl)
		if it.lru.isEntryExpired(entry, it.createTime) {
			nextItem := it.nextItem.Next()
			it.lru.deleteInternal(it.nextItem)
			it.nextItem = nextItem
		} else {
			return
		}
	}
}

// Iterator returns an iterator to the map. This map
// does not use re-entrant locks, so access or modification
// to the map during iteration can cause a dead lock.
func (c *lru) Iterator() Iterator {
	c.mut.Lock()
	iterator := &iteratorImpl{
		lru:        c,
		createTime: c.timeSource.Now(),
		nextItem:   c.byAccess.Front(),
	}
	iterator.prepareNext()
	return iterator
}

func (entry *entryImpl) Key() interface{} {
	return entry.key
}

func (entry *entryImpl) Value() interface{} {
	return entry.value
}

func (entry *entryImpl) CreateTime() time.Time {
	return entry.createTime
}

// New creates a new cache with the given options
func New(opts *Options) Cache {
	if opts == nil || (opts.MaxCount <= 0 && (opts.MaxSize() <= 0 || opts.GetCacheItemSizeFunc == nil)) {
		panic("Either MaxCount (count based) or " +
			"MaxSize and GetCacheItemSizeFunc (size based) options must be provided for the LRU cache")
	}

	timeSource := opts.TimeSource
	if timeSource == nil {
		timeSource = clock.NewRealTimeSource()
	}

	cache := &lru{
		byAccess:      list.New(),
		byKey:         make(map[interface{}]*list.Element, opts.InitialCapacity),
		ttl:           opts.TTL,
		pin:           opts.Pin,
		rmFunc:        opts.RemovedFunc,
		activelyEvict: opts.ActivelyEvict,
		timeSource:    timeSource,
		logger:        opts.Logger,
		isSizeBased:   opts.IsSizeBased,
		metricsScope:  opts.MetricsScope,
	}

	if cache.logger == nil {
		cache.logger = log.NewNoop()
	}

	if cache.metricsScope == nil {
		cache.metricsScope = metrics.NoopScope(1)
	}

	if cache.isSizeBased {
		cache.sizeFunc = opts.GetCacheItemSizeFunc
		cache.maxSize = opts.MaxSize
		if cache.maxSize == nil {
			// If maxSize is not defined for size-based cache, set default to cacheCountLimit
			cache.maxSize = dynamicproperties.GetIntPropertyFn(cacheDefaultSizeLimit)
		}
		cache.sizeByKey = make(map[interface{}]uint64, opts.InitialCapacity)
	} else {
		// cache is count based if max size and sizeFunc are not provided
		cache.maxCount = opts.MaxCount
		cache.maxSize = dynamicproperties.GetIntPropertyFn(0)
		cache.sizeFunc = func(interface{}) uint64 {
			return 0
		}
	}

	cache.logger.Info("LRU cache initialized",
		tag.Value(map[string]interface{}{
			"isSizeBased": cache.isSizeBased,
			"maxCount":    cache.maxCount,
			"maxSize":     cache.maxSize(),
		}),
	)

	return cache
}

// Get retrieves the value stored under the given key
func (c *lru) Get(key interface{}) interface{} {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.evictExpiredItems()

	element := c.byKey[key]
	if element == nil {
		c.metricsScope.IncCounter(metrics.BaseCacheMiss)
		return nil
	}

	entry := element.Value.(*entryImpl)

	if c.isEntryExpired(entry, c.timeSource.Now()) {
		// Entry has expired
		c.deleteInternal(element)
		c.metricsScope.IncCounter(metrics.BaseCacheMiss)
		return nil
	}

	if c.pin {
		entry.refCount++
	}
	c.byAccess.MoveToFront(element)
	c.metricsScope.IncCounter(metrics.BaseCacheHit)
	return entry.value
}

// Put puts a new value associated with a given key, returning the existing value (if present)
func (c *lru) Put(key interface{}, value interface{}) interface{} {
	if c.pin {
		panic("Cannot use Put API in Pin mode. Use Delete and PutIfNotExist if necessary")
	}
	val, _ := c.putInternal(key, value, true)
	return val
}

// PutIfNotExist puts a value associated with a given key if it does not exist
func (c *lru) PutIfNotExist(key interface{}, value interface{}) (interface{}, error) {
	existing, err := c.putInternal(key, value, false)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		// This is a new value
		return value, err
	}

	return existing, err
}

// Delete deletes a key, value pair associated with a key
func (c *lru) Delete(key interface{}) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.evictExpiredItems()

	element := c.byKey[key]
	if element != nil {
		c.deleteInternal(element)
	}
}

// Release decrements the ref count of a pinned element.
func (c *lru) Release(key interface{}) {
	c.mut.Lock()
	defer c.mut.Unlock()

	elt, ok := c.byKey[key]
	if !ok {
		return
	}
	entry := elt.Value.(*entryImpl)
	entry.refCount--
}

// Size returns the number of entries currently in the lru, useful if cache is not full
func (c *lru) Size() int {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.evictExpiredItems()

	return len(c.byKey)
}

// evictExpiredItems evicts all items in the cache which are expired
func (c *lru) evictExpiredItems() {
	if !c.activelyEvict {
		return // do nothing if activelyEvict is not set
	}

	now := c.timeSource.Now()
	for elt := c.byAccess.Back(); elt != nil; elt = c.byAccess.Back() {
		if !c.isEntryExpired(elt.Value.(*entryImpl), now) {
			// List is sorted by item age, so we can stop as soon as we found first non expired item.
			break
		}
		c.deleteInternal(elt)
	}
}

// Put puts a new value associated with a given key, returning the existing value (if present)
// allowUpdate flag is used to control overwrite behavior if the value exists
func (c *lru) putInternal(key interface{}, value interface{}, allowUpdate bool) (interface{}, error) {
	valueSize := uint64(0)
	if c.isSizeBased {
		sizeableValue, ok := value.(Sizeable)
		if !ok {
			return nil, fmt.Errorf("value %T does not implement sizable. Key: %+v", value, key)
		}
		valueSize = sizeableValue.ByteSize()
	}
	c.mut.Lock()
	defer c.mut.Unlock()

	c.evictExpiredItems()

	element := c.byKey[key]
	if element != nil {
		entry := element.Value.(*entryImpl)
		if c.isEntryExpired(entry, c.timeSource.Now()) {
			// Entry has expired
			c.deleteInternal(element)
		} else {
			existing := entry.value
			if allowUpdate {
				if c.isSizeBased {
					for c.isCacheFull() {
						oldest := c.byAccess.Back().Value.(*entryImpl)
						if oldest.refCount > 0 {
							// Cache is full with pinned elements
							// we don't update
							return existing, ErrCacheFull
						}
						c.deleteInternal(c.byAccess.Back())
					}
					c.updateSizeOnDelete(key)
					c.updateSizeOnAdd(key, valueSize)
				}
				entry.value = value
				if c.ttl != 0 {
					entry.createTime = c.timeSource.Now()
				}
			}

			c.byAccess.MoveToFront(element)
			if c.pin {
				entry.refCount++
			}
			return existing, nil
		}
	}

	entry := &entryImpl{
		key:   key,
		value: value,
	}

	if c.pin {
		entry.refCount++
	}

	if c.ttl != 0 {
		entry.createTime = c.timeSource.Now()
	}

	c.byKey[key] = c.byAccess.PushFront(entry)
	c.updateSizeOnAdd(key, valueSize)
	for c.isCacheFull() {
		oldest := c.byAccess.Back().Value.(*entryImpl)

		if oldest.refCount > 0 {
			// Cache is full with pinned elements
			// revert the insert and return
			c.deleteInternal(c.byAccess.Front())
			return nil, ErrCacheFull
		}

		c.deleteInternal(c.byAccess.Back())
	}
	return nil, nil
}

func (c *lru) deleteInternal(element *list.Element) {
	entry := c.byAccess.Remove(element).(*entryImpl)
	if c.rmFunc != nil {
		go c.rmFunc(entry.value)
	}
	delete(c.byKey, entry.key)
	c.updateSizeOnDelete(entry.key)
}

func (c *lru) isEntryExpired(entry *entryImpl, currentTime time.Time) bool {
	return entry.refCount == 0 && !entry.createTime.IsZero() && currentTime.After(entry.createTime.Add(c.ttl))
}

func (c *lru) isCacheFull() bool {
	count := len(c.byKey)
	// if the value size is greater than maxSize(should never happen) then the item won't be cached
	return (!c.isSizeBased && count == c.maxCount) || c.currSize > uint64(c.maxSize()) || count > cacheCountLimit
}

func (c *lru) updateSizeOnAdd(key interface{}, valueSize uint64) {
	if c.isSizeBased {
		c.sizeByKey[key] = valueSize
		// the int overflow should not happen here
		c.currSize += uint64(valueSize)
		c.emitSizeOnUpdate()
	}
}

func (c *lru) updateSizeOnDelete(key interface{}) {
	if c.isSizeBased {
		c.currSize -= uint64(c.sizeByKey[key])
		c.emitSizeOnUpdate()
		delete(c.sizeByKey, key)
	}
}

func (c *lru) emitSizeOnUpdate() {
	if c.isSizeBased {
		c.metricsScope.UpdateGauge(metrics.BaseCacheByteSize, float64(c.currSize))
		c.metricsScope.UpdateGauge(metrics.BaseCacheByteSizeLimitGauge, float64(c.maxSize()))
	}
}
