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

package queue

import (
	"context"
	"encoding/json"
	"errors"
	"runtime/debug"
	"sync"

	"github.com/uber/cadence/common/activecluster"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/shard"
	htask "github.com/uber/cadence/service/history/task"
)

type (
	// TaskAllocator verifies if a task should be processed or not
	TaskAllocator interface {
		VerifyActiveTask(domainID, wfID, rID string, task interface{}) (bool, error)
		VerifyFailoverActiveTask(targetDomainIDs map[string]struct{}, domainID, wfID, rID string, task interface{}) (bool, error)
		VerifyStandbyTask(standbyCluster string, domainID, wfID, rID string, task interface{}) (bool, error)
		Lock()
		Unlock()
	}

	taskAllocatorImpl struct {
		currentClusterName string
		shard              shard.Context
		domainCache        cache.DomainCache
		activeClusterMgr   activecluster.Manager
		logger             log.Logger

		locker sync.RWMutex
	}
)

// NewTaskAllocator create a new task allocator
func NewTaskAllocator(shard shard.Context) TaskAllocator {
	return &taskAllocatorImpl{
		currentClusterName: shard.GetService().GetClusterMetadata().GetCurrentClusterName(),
		shard:              shard,
		domainCache:        shard.GetDomainCache(),
		activeClusterMgr:   shard.GetActiveClusterManager(),
		logger:             shard.GetLogger(),
	}
}

// VerifyActiveTask, will return true if task activeness check is successful
func (t *taskAllocatorImpl) VerifyActiveTask(domainID, wfID, rID string, task interface{}) (bool, error) {
	return t.verifyTaskActiveness(t.currentClusterName, domainID, wfID, rID, task)
}

// VerifyFailoverActiveTask, will return true if task activeness check is successful
func (t *taskAllocatorImpl) VerifyFailoverActiveTask(targetDomainIDs map[string]struct{}, domainID, wfID, rID string, task interface{}) (bool, error) {
	_, ok := targetDomainIDs[domainID]
	if !ok {
		return false, nil
	}

	return t.verifyTaskActiveness("", domainID, wfID, rID, task)
}

// VerifyStandbyTask, will return true if task standbyness check is successful
func (t *taskAllocatorImpl) VerifyStandbyTask(standbyCluster string, domainID, wfID, rID string, task interface{}) (bool, error) {
	return t.verifyTaskActiveness(standbyCluster, domainID, wfID, rID, task)
}

func (t *taskAllocatorImpl) verifyTaskActiveness(cluster string, domainID, wfID, rID string, task interface{}) (b bool, e error) {
	if t.logger.DebugOn() {
		defer func() {
			data, err := json.Marshal(task)
			if err != nil {
				t.logger.Error("Failed to marshal task.", tag.Error(err))
			}
			t.logger.Debugf("verifyTaskActiveness returning (%v, %v) for cluster %s, domainID %s, wfID %s, rID %s, task %s, stacktrace %s",
				b,
				e,
				cluster,
				domainID,
				wfID,
				rID,
				string(data),
				string(debug.Stack()),
			)
		}()
	}
	t.locker.RLock()
	defer t.locker.RUnlock()

	domainEntry, err := t.domainCache.GetDomainByID(domainID)
	if err != nil {
		// it is possible that the domain is deleted
		// we should treat that domain as not active
		if _, ok := err.(*types.EntityNotExistsError); !ok {
			t.logger.Warn("Cannot find domain", tag.WorkflowDomainID(domainID))
			return false, err
		}
		t.logger.Warn("Cannot find domain, default to not process task.", tag.WorkflowDomainID(domainID), tag.Value(task))
		return false, nil
	}

	// return error for pending active domain so the task can be retried
	if err := t.checkDomainPendingActive(
		domainEntry,
		domainID,
		task,
	); err != nil {
		return false, err
	}

	if cluster == "" { // failover queue task. Revisit this logic. It's copied from previous implementation
		return true, nil
	}

	if domainEntry.GetReplicationConfig().IsActiveActive() {
		resp, err := t.activeClusterMgr.LookupWorkflow(context.Background(), domainID, wfID, rID)
		if err != nil {
			t.logger.Debugf("Error getting active cluster for domain %s, wfID %s, rID %s: %v", domainID, wfID, rID, err)
			return false, err
		}
		if resp.ClusterName != cluster {
			t.logger.Debugf("Skip task because workflow %s of domain %s is not active on cluster: %s", wfID, domainID, cluster)
			return false, nil
		}

		t.logger.Debugf("Active cluster for domain %s, wfID %s, rID %s: %s", domainID, wfID, rID, resp.ClusterName)
		return true, nil
	}

	isActiveInGivenCluster, _ := domainEntry.IsActiveIn(cluster)
	if !isActiveInGivenCluster {
		t.logger.Debugf("Skip task because domain %s is not active on cluster: %s", domainID, cluster)
		return false, nil
	}

	return true, nil
}

func (t *taskAllocatorImpl) checkDomainPendingActive(domainEntry *cache.DomainCacheEntry, taskDomainID string, task interface{}) error {
	if domainEntry.IsGlobalDomain() && domainEntry.GetFailoverEndTime() != nil {
		// the domain is pending active, pause on processing this task
		t.logger.Debug("Domain is not in pending active, skip task.", tag.WorkflowDomainID(taskDomainID), tag.Value(task))
		return htask.ErrTaskPendingActive
	}
	return nil
}

// Lock block all task allocation
func (t *taskAllocatorImpl) Lock() {
	t.locker.Lock()
}

// Unlock resume the task allocator
func (t *taskAllocatorImpl) Unlock() {
	t.locker.Unlock()
}

// isDomainNotRegistered checks either if domain does not exist or is in deprecated or deleted status
func isDomainNotRegistered(shard shard.Context, domainID string) (bool, error) {
	domainEntry, err := shard.GetDomainCache().GetDomainByID(domainID)
	if err != nil {
		// error in finding a domain
		return false, err
	}
	info := domainEntry.GetInfo()
	if info == nil {
		return false, errors.New("domain info is nil in cache")
	}
	return info.Status == persistence.DomainStatusDeprecated || info.Status == persistence.DomainStatusDeleted, nil
}
