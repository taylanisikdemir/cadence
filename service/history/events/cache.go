// Copyright (c) 2020 Uber Technologies, Inc.
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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination cache_mock.go

package events

import (
	"context"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig/dynamicproperties"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
)

type (
	// Cache caches workflow history event
	Cache interface {
		GetEvent(
			ctx context.Context,
			shardID int,
			domainID string,
			workflowID string,
			runID string,
			firstEventID int64,
			eventID int64,
			branchToken []byte,
		) (*types.HistoryEvent, error)
		PutEvent(
			domainID string,
			workflowID string,
			runID string,
			eventID int64,
			event *types.HistoryEvent,
		)
	}

	cacheImpl struct {
		cache.Cache
		domainCache    cache.DomainCache
		historyManager persistence.HistoryManager
		disabled       bool
		logger         log.Logger
		metricsClient  metrics.Client
		shardID        *int
	}

	eventKey struct {
		domainID   string
		workflowID string
		runID      string
		eventID    int64
	}
)

var (
	errEventNotFoundInBatch = &types.EntityNotExistsError{Message: "History event not found within expected batch"}
)

var _ Cache = (*cacheImpl)(nil)

// NewGlobalCache creates a new global events cache
func NewGlobalCache(
	initialCount int,
	maxCount int,
	ttl time.Duration,
	historyManager persistence.HistoryManager,
	logger log.Logger,
	metricsClient metrics.Client,
	enableSizeBasedCache dynamicproperties.BoolPropertyFn,
	maxSize dynamicproperties.IntPropertyFn,
	domainCache cache.DomainCache,
) Cache {
	return newCacheWithOption(
		nil,
		initialCount,
		maxCount,
		ttl,
		historyManager,
		false,
		logger,
		metricsClient,
		enableSizeBasedCache,
		maxSize,
		domainCache,
	)
}

// NewCache creates a new events cache
func NewCache(
	shardID int,
	historyManager persistence.HistoryManager,
	config *config.Config,
	logger log.Logger,
	metricsClient metrics.Client,
	domainCache cache.DomainCache,
) Cache {
	return newCacheWithOption(
		&shardID,
		config.EventsCacheInitialCount(),
		config.EventsCacheMaxCount(),
		config.EventsCacheTTL(),
		historyManager,
		false,
		logger,
		metricsClient,
		config.EnableSizeBasedHistoryEventCache,
		config.EventsCacheMaxSize,
		domainCache,
	)
}

func newCacheWithOption(
	shardID *int,
	initialCount int,
	maxCount int,
	ttl time.Duration,
	historyManager persistence.HistoryManager,
	disabled bool,
	logger log.Logger,
	metricsClient metrics.Client,
	enableSizeBasedCache dynamicproperties.BoolPropertyFn,
	maxSize dynamicproperties.IntPropertyFn,
	domainCache cache.DomainCache,
) *cacheImpl {
	opts := &cache.Options{}
	opts.InitialCapacity = initialCount
	opts.TTL = ttl
	opts.MaxCount = maxCount
	opts.MetricsScope = metricsClient.Scope(metrics.EventsCacheGetEventScope)
	if shardID != nil {
		opts.MetricsScope = opts.MetricsScope.Tagged(metrics.ShardIDTag(*shardID))
	}

	opts.MaxSize = maxSize
	opts.Logger = logger.WithTags(tag.ComponentEventsCache)
	opts.IsSizeBased = enableSizeBasedCache

	return &cacheImpl{
		Cache:          cache.New(opts),
		domainCache:    domainCache,
		historyManager: historyManager,
		disabled:       disabled,
		logger:         logger.WithTags(tag.ComponentEventsCache),
		metricsClient:  metricsClient,
		shardID:        shardID,
	}
}

func newEventKey(
	domainID,
	workflowID,
	runID string,
	eventID int64,
) eventKey {
	return eventKey{
		domainID:   domainID,
		workflowID: workflowID,
		runID:      runID,
		eventID:    eventID,
	}
}

func (e *cacheImpl) GetEvent(
	ctx context.Context,
	shardID int,
	domainID string,
	workflowID string,
	runID string,
	firstEventID int64,
	eventID int64,
	branchToken []byte,
) (*types.HistoryEvent, error) {
	e.metricsClient.IncCounter(metrics.EventsCacheGetEventScope, metrics.CacheRequests)
	sw := e.metricsClient.StartTimer(metrics.EventsCacheGetEventScope, metrics.CacheLatency)
	defer sw.Stop()

	key := newEventKey(domainID, workflowID, runID, eventID)
	// Test hook for disabling cache
	if !e.disabled {
		event, cacheHit := e.Cache.Get(key).(*types.HistoryEvent)
		if cacheHit {
			return event, nil
		}
	}

	e.metricsClient.IncCounter(metrics.EventsCacheGetEventScope, metrics.CacheMissCounter)
	event, err := e.getHistoryEventFromStore(ctx, firstEventID, eventID, branchToken, shardID, domainID)

	if err != nil {
		e.metricsClient.IncCounter(metrics.EventsCacheGetEventScope, metrics.CacheFailures)
		e.logger.Error("EventsCache unable to retrieve event from store",
			tag.Error(err),
			tag.WorkflowID(workflowID),
			tag.WorkflowRunID(runID),
			tag.WorkflowDomainID(domainID),
			tag.WorkflowEventID(eventID))
		return nil, err
	}

	e.Put(key, event)
	return event, nil
}

func (e *cacheImpl) PutEvent(
	domainID, workflowID,
	runID string,
	eventID int64,
	event *types.HistoryEvent,
) {
	e.metricsClient.IncCounter(metrics.EventsCachePutEventScope, metrics.CacheRequests)
	sw := e.metricsClient.StartTimer(metrics.EventsCachePutEventScope, metrics.CacheLatency)
	defer sw.Stop()

	key := newEventKey(domainID, workflowID, runID, eventID)
	e.Put(key, event)
}

func (e *cacheImpl) getHistoryEventFromStore(
	ctx context.Context,
	firstEventID int64,
	eventID int64,
	branchToken []byte,
	shardID int,
	domainID string,
) (*types.HistoryEvent, error) {
	e.metricsClient.IncCounter(metrics.EventsCacheGetFromStoreScope, metrics.CacheRequests)
	sw := e.metricsClient.StartTimer(metrics.EventsCacheGetFromStoreScope, metrics.CacheLatency)
	defer sw.Stop()

	var historyEvents []*types.HistoryEvent
	domainName, err := e.domainCache.GetDomainName(domainID)
	if err != nil {
		return nil, err
	}
	response, err := e.historyManager.ReadHistoryBranch(ctx, &persistence.ReadHistoryBranchRequest{
		BranchToken:   branchToken,
		MinEventID:    firstEventID,
		MaxEventID:    eventID + 1,
		PageSize:      1,
		NextPageToken: nil,
		ShardID:       common.IntPtr(shardID),
		DomainName:    domainName,
	})

	if err != nil {
		e.metricsClient.IncCounter(metrics.EventsCacheGetFromStoreScope, metrics.CacheFailures)
		return nil, err
	}

	historyEvents = response.HistoryEvents

	// find history event from batch and return back single event to caller
	for _, e := range historyEvents {
		if e.ID == eventID {
			return e, nil
		}
	}

	return nil, errEventNotFoundInBatch
}
