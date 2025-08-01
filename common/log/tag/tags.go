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

package tag

import (
	"fmt"
	"time"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/types"
)

// All logging tags are defined in this file.
// To help finding available tags, we recommend that all tags to be categorized and placed in the corresponding section.
// We currently have those categories:
//   0. Common tags that can't be categorized(or belong to more than one)
//   1. Workflow: these tags are information that are useful to our customer, like workflow-id/run-id/task-list/...
//   2. System : these tags are internal information which usually cannot be understood by our customers,

// /////////////////  Common tags defined here ///////////////////

// Error returns tag for Error
func Error(err error) Tag {
	return newErrorTag("error", err)
}

// ClusterName returns tag for ClusterName
func ClusterName(clusterName string) Tag {
	return newStringTag("cluster-name", clusterName)
}

// Region returns tag for Region
func Region(region string) Tag {
	return newStringTag("region", region)
}

// ActiveClusterName returns tag for ActiveClusterName.
func ActiveClusterName(activeClusterName string) Tag {
	return newStringTag("active-cluster-name", activeClusterName)
}

// ActiveClusterChangeType returns tag for ActiveClusterChangeType.
func ActiveClusterChangeType(changeType string) Tag {
	return newStringTag("active-active-change-type", changeType)
}

// Timestamp returns tag for Timestamp
func Timestamp(timestamp time.Time) Tag {
	return newTimeTag("timestamp", timestamp)
}

func EarliestTime(time int64) Tag {
	return newInt64("earliest-time", time)
}

func LatestTime(time int64) Tag {
	return newInt64("latest-time", time)
}

// /////////////////  Workflow tags defined here: ( wf is short for workflow) ///////////////////

// WorkflowAction returns tag for WorkflowAction
func workflowAction(action string) Tag {
	return newPredefinedStringTag("wf-action", action)
}

// WorkflowListFilterType returns tag for WorkflowListFilterType
func workflowListFilterType(listFilterType string) Tag {
	return newPredefinedStringTag("wf-list-filter-type", listFilterType)
}

// general

// ClientImpl returns tag for ClientImpl
func ClientImpl(clientImpl string) Tag {
	return newStringTag("client-impl", clientImpl)
}

// ClientFeatureVersion returns tag for ClientFeatureVersion
func ClientFeatureVersion(clientFeatureVersion string) Tag {
	return newStringTag("client-feature-version", clientFeatureVersion)
}

// WorkflowError returns tag for WorkflowError
func WorkflowError(error error) Tag {
	return newErrorTag("wf-error", error)
}

// WorkflowTimeoutType returns tag for WorkflowTimeoutType
func WorkflowTimeoutType(timeoutType int64) Tag {
	return newInt64("wf-timeout-type", timeoutType)
}

// ActivityTimeoutType returns tag for ActivityTimeoutType
func ActivityTimeoutType(timerType shared.TimeoutType) Tag {
	return newStringTag("activity-timer-type", timerType.String())
}

// WorkflowPollContextTimeout returns tag for WorkflowPollContextTimeout
func WorkflowPollContextTimeout(pollContextTimeout time.Duration) Tag {
	return newDurationTag("wf-poll-context-timeout", pollContextTimeout)
}

// WorkflowHandlerName returns tag for WorkflowHandlerName
func WorkflowHandlerName(handlerName string) Tag {
	return newStringTag("wf-handler-name", handlerName)
}

// WorkflowID returns tag for WorkflowID
func WorkflowID(workflowID string) Tag {
	return newStringTag("wf-id", workflowID)
}

// TargetWorkflowID returns tag for target WorkflowID
func TargetWorkflowID(workflowID string) Tag {
	return newStringTag("target-wf-id", workflowID)
}

// WorkflowType returns tag for WorkflowType
func WorkflowType(wfType string) Tag {
	return newStringTag("wf-type", wfType)
}

// WorkflowSignalName returns tag for WorkflowSignalName
func WorkflowSignalName(signalName string) Tag {
	return newStringTag("wf-signal-name", signalName)
}

// WorkflowRequestID returns tag for WorkflowRequestID
func WorkflowRequestID(requestID string) Tag {
	return newStringTag("wf-request-id", requestID)
}

// WorkflowState returns tag for WorkflowState
func WorkflowState(s int) Tag {
	return newInt("wf-state", s)
}

// WorkflowRunID returns tag for WorkflowRunID
func WorkflowRunID(runID string) Tag {
	return newStringTag("wf-run-id", runID)
}

// TargetWorkflowRunID returns tag for target  WorkflowRunID
func TargetWorkflowRunID(runID string) Tag {
	return newStringTag("target-wf-run-id", runID)
}

// WorkflowResetBaseRunID returns tag for WorkflowResetBaseRunID
func WorkflowResetBaseRunID(runID string) Tag {
	return newStringTag("wf-reset-base-run-id", runID)
}

// WorkflowResetNewRunID returns tag for WorkflowResetNewRunID
func WorkflowResetNewRunID(runID string) Tag {
	return newStringTag("wf-reset-new-run-id", runID)
}

// WorkflowBinaryChecksum returns tag for WorkflowBinaryChecksum
func WorkflowBinaryChecksum(cs string) Tag {
	return newStringTag("wf-binary-checksum", cs)
}

// WorkflowActivityID returns tag for WorkflowActivityID
func WorkflowActivityID(id string) Tag {
	return newStringTag("wf-activity-id", id)
}

// WorkflowTimerID returns tag for WorkflowTimerID
func WorkflowTimerID(id string) Tag {
	return newStringTag("wf-timer-id", id)
}

// WorkflowBeginningRunID returns tag for WorkflowBeginningRunID
func WorkflowBeginningRunID(beginningRunID string) Tag {
	return newStringTag("wf-beginning-run-id", beginningRunID)
}

// WorkflowEndingRunID returns tag for WorkflowEndingRunID
func WorkflowEndingRunID(endingRunID string) Tag {
	return newStringTag("wf-ending-run-id", endingRunID)
}

// WorkflowDecisionTimeoutSeconds returns tag for WorkflowDecisionTimeoutSeconds
func WorkflowDecisionTimeoutSeconds(s int32) Tag {
	return newInt32("wf-decision-timeout", s)
}

// QueryID returns tag for QueryID
func QueryID(queryID string) Tag {
	return newStringTag("query-id", queryID)
}

// BlobSizeViolationOperation returns tag for BlobSizeViolationOperation
func BlobSizeViolationOperation(operation string) Tag {
	return newStringTag("blob-size-violation-operation", operation)
}

// WorkflowCronSchedule returns a tag to report a workflow's cron schedule
func WorkflowCronSchedule(schedule string) Tag {
	return newStringTag("wf-cron-schedule", schedule)
}

// WorkflowCloseStatus returns a tag to report a workflow's close status
func WorkflowCloseStatus(status int) Tag {
	return newInt("close-status", status)
}

// IsWorkflowOpen returns a tag to report a workflow is open or not
func IsWorkflowOpen(isOpen bool) Tag {
	return newBoolTag("is-workflow-open", isOpen)
}

// WorkflowTerminationReason returns a tag to report a workflow's termination reason
func WorkflowTerminationReason(reason string) Tag {
	return newStringTag("wf-termination-reason", reason)
}

func Duration(duration time.Duration) Tag {
	return newDurationTag("duration", duration)
}

// domain related

// WorkflowDomainID returns tag for WorkflowDomainID
func WorkflowDomainID(domainID string) Tag {
	return newStringTag("wf-domain-id", domainID)
}

// TargetWorkflowDomainID returns tag for target WorkflowDomainID
func TargetWorkflowDomainID(domainID string) Tag {
	return newStringTag("target-wf-domain-id", domainID)
}

// WorkflowDomainName returns tag for WorkflowDomainName
func WorkflowDomainName(domainName string) Tag {
	return newStringTag("wf-domain-name", domainName)
}

// WorkflowDomainIDs returns tag for WorkflowDomainIDs
func WorkflowDomainIDs(domainIDs interface{}) Tag {
	return newObjectTag("wf-domain-ids", domainIDs)
}

// OperationName returns tag for OperationName
func OperationName(operationName string) Tag {
	return newStringTag("operation-name", operationName)
}

// history event ID related

// WorkflowEventID returns tag for WorkflowEventID
func WorkflowEventID(eventID int64) Tag {
	return newInt64("wf-history-event-id", eventID)
}

// WorkflowScheduleID returns tag for WorkflowScheduleID
func WorkflowScheduleID(scheduleID int64) Tag {
	return newInt64("wf-schedule-id", scheduleID)
}

// WorkflowStartedID returns tag for WorkflowStartedID
func WorkflowStartedID(id int64) Tag {
	return newInt64("wf-started-id", id)
}

// WorkflowInitiatedID returns tag for WorkflowInitiatedID
func WorkflowInitiatedID(id int64) Tag {
	return newInt64("wf-initiated-id", id)
}

// WorkflowFirstEventID returns tag for WorkflowFirstEventID
func WorkflowFirstEventID(firstEventID int64) Tag {
	return newInt64("wf-first-event-id", firstEventID)
}

// WorkflowNextEventID returns tag for WorkflowNextEventID
func WorkflowNextEventID(nextEventID int64) Tag {
	return newInt64("wf-next-event-id", nextEventID)
}

// WorkflowResetNextEventID returns tag for WorkflowResetNextEventID
func WorkflowResetNextEventID(resetNextEventID int64) Tag {
	return newInt64("wf-reset-next-event-id", resetNextEventID)
}

func WorkflowExternalEntityType(externalEntityType string) Tag {
	return newStringTag("wf-external-entity-type", externalEntityType)
}

func WorkflowExternalEntityKey(externalEntityKey string) Tag {
	return newStringTag("wf-external-entity-key", externalEntityKey)
}

// history tree

// WorkflowTreeID returns tag for WorkflowTreeID
func WorkflowTreeID(treeID string) Tag {
	return newStringTag("wf-tree-id", treeID)
}

// WorkflowBranchID returns tag for WorkflowBranchID
func WorkflowBranchID(branchID string) Tag {
	return newStringTag("wf-branch-id", branchID)
}

// workflow task

// WorkflowDecisionType returns tag for WorkflowDecisionType
func WorkflowDecisionType(decisionType int64) Tag {
	return newInt64("wf-decision-type", decisionType)
}

// WorkflowQueryType returns tag for WorkflowQueryType
func WorkflowQueryType(qt string) Tag {
	return newStringTag("wf-query-type", qt)
}

// WorkflowDecisionFailCause returns tag for WorkflowDecisionFailCause
func WorkflowDecisionFailCause(decisionFailCause int64) Tag {
	return newInt64("wf-decision-fail-cause", decisionFailCause)
}

// WorkflowTaskListType returns tag for WorkflowTaskListType
func WorkflowTaskListType(taskListType int) Tag {
	return newInt("wf-task-list-type", taskListType)
}

// WorkflowTaskListKind returns tag for WorkflowTaskListKind
func WorkflowTaskListKind(taskListKind int32) Tag {
	return newInt32("wf-task-list-kind", taskListKind)
}

// WorkflowTaskListName returns tag for WorkflowTaskListName
func WorkflowTaskListName(taskListName string) Tag {
	return newStringTag("wf-task-list-name", taskListName)
}

// size limit

// WorkflowSize returns tag for WorkflowSize
func WorkflowSize(workflowSize int64) Tag {
	return newInt64("wf-size", workflowSize)
}

// WorkflowSignalCount returns tag for SignalCount
func WorkflowSignalCount(signalCount int32) Tag {
	return newInt32("wf-signal-count", signalCount)
}

// WorkflowHistorySize returns tag for HistorySize
func WorkflowHistorySize(historySize int) Tag {
	return newInt("wf-history-size", historySize)
}

// WorkflowHistorySizeBytes returns tag for HistorySizeBytes
func WorkflowHistorySizeBytes(historySizeBytes int) Tag {
	return newInt("wf-history-size-bytes", historySizeBytes)
}

// WorkflowEventCount returns tag for EventCount
func WorkflowEventCount(eventCount int) Tag {
	return newInt("wf-event-count", eventCount)
}

func WorkflowEventType(eventType string) Tag {
	return newStringTag("wf-event-type", eventType)
}

// /////////////////  System tags defined here:  ///////////////////
// Tags with pre-define values

// component returns tag for component
func component(component string) Tag {
	return newPredefinedStringTag("component", component)
}

// lifecycle returns tag for lifecycle
func lifecycle(lifecycle string) Tag {
	return newPredefinedStringTag("lifecycle", lifecycle)
}

// storeOperation returns tag for storeOperation
func storeOperation(storeOperation string) Tag {
	return newPredefinedStringTag("store-operation", storeOperation)
}

// clientOperation returns tag for clientOperation
func clientOperation(clientOperation string) Tag {
	return newPredefinedStringTag("client-operation", clientOperation)
}

// operationResult returns tag for operationResult
func operationResult(operationResult string) Tag {
	return newPredefinedStringTag("operation-result", operationResult)
}

// errorType returns tag for errorType
func errorType(errorType string) Tag {
	return newPredefinedStringTag("error-type", errorType)
}

// shardupdate returns tag for shardupdate
func shardupdate(shardupdate string) Tag {
	return newPredefinedStringTag("shard-update", shardupdate)
}

// idType returns tag for idType
func idType(idType string) Tag {
	return newPredefinedStringTag("id-type", idType)
}

// queueType returns tag for queueType
func queueType(queueType string) Tag {
	return newPredefinedStringTag("queue-type", queueType)
}

// general

// Service returns tag for Service
func Service(sv string) Tag {
	return newStringTag("service", sv)
}

// DestService returns tag for destination service
func DestService(sv string) Tag {
	return newStringTag("dest-service", sv)
}

// Addresses returns tag for Addresses
func Addresses(ads []string) Tag {
	return newObjectTag("addresses", ads)
}

// Subscriber returns tag for Subscriber
func Subscriber(subscriber string) Tag {
	return newStringTag("subscriber", subscriber)
}

// Address return tag for Address
func Address(ad string) Tag {
	return newStringTag("address", ad)
}

// Env return tag for runtime environment
func Env(env string) Tag {
	return newStringTag("env", env)
}

// Key returns tag for Key
func Key(k string) Tag {
	return newStringTag("key", k)
}

// Name returns tag for Name
func Name(k string) Tag {
	return newStringTag("name", k)
}

// Mode returns tag for Mode
func Mode(mode string) Tag {
	return newStringTag("mode", mode)
}

// Value returns tag for Value
func Value(v interface{}) Tag {
	return newObjectTag("value", v)
}

// Reason returns tag for Reason
func Reason(reason string) Tag {
	return newStringTag("reason", reason)
}

// ValueType returns tag for ValueType
func ValueType(v interface{}) Tag {
	return newStringTag("value-type", fmt.Sprintf("%T", v))
}

// DefaultValue returns tag for DefaultValue
func DefaultValue(v interface{}) Tag {
	return newObjectTag("default-value", v)
}

// Port returns tag for Port
func Port(p int) Tag {
	return newInt("port", p)
}

// CursorTimestamp returns tag for CursorTimestamp
func CursorTimestamp(timestamp time.Time) Tag {
	return newTimeTag("cursor-timestamp", timestamp)
}

// MetricScope returns tag for MetricScope
func MetricScope(metricScope int) Tag {
	return newInt("metric-scope", metricScope)
}

// StoreType returns tag for StoreType
func StoreType(storeType string) Tag {
	return newPredefinedStringTag("store-type", storeType)
}

// StoreError returns tag for StoreError
func StoreError(storeErr error) Tag {
	return newErrorTag("store-error", storeErr)
}

// StoreShard returns tag for StoreShard
func StoreShard(storeShard string) Tag {
	return newPredefinedStringTag("store-shard", storeShard)
}

// ClientError returns tag for ClientError
func ClientError(clientErr error) Tag {
	return newErrorTag("client-error", clientErr)
}

// DetailInfo returns tag for DetailInfo
func DetailInfo(i string) Tag {
	return newStringTag("detail-info", i)
}

// Counter returns tag for Counter
func Counter(c int) Tag {
	return newInt("counter", c)
}

// Number returns tag for Number
func Number(n int64) Tag {
	return newInt64("number", n)
}

// NextNumber returns tag for NextNumber
func NextNumber(n int64) Tag {
	return newInt64("next-number", n)
}

// Bool returns tag for Bool
func Bool(b bool) Tag {
	return newBoolTag("bool", b)
}

/* Tags for logging manual access */

// RequestCaller returns tag for caller (the name of the service making this request)
func RequestCaller(callerName string) Tag {
	return newStringTag("request-caller", callerName)
}

// ActorID returns tag for the actor ID
func ActorID(actorID string) Tag {
	return newStringTag("actor-id", actorID)
}

// ActorEmail returns tag for the actor's email address
func ActorEmail(actorEmail string) Tag {
	return newStringTag("actor-email", actorEmail)
}

// HandlerCall returns tag for the API name of a request
func HandlerCall(handlerCall string) Tag {
	return newStringTag("handler-call", handlerCall)
}

// RequestBody returns the tag for the API request body
func RequestBody(requestBody string) Tag {
	return newStringTag("request-body", requestBody)
}

// RequestType return tag for the type of request (internal, external)
func RequestType(requestType string) Tag {
	return newStringTag("request-type", requestType)
}

// history engine shard

// ShardID returns tag for ShardID
func ShardID(shardID int) Tag {
	return newInt("shard-id", shardID)
}

// ShardTime returns tag for ShardTime
func ShardTime(shardTime interface{}) Tag {
	return newObjectTag("shard-time", shardTime)
}

// ShardReplicationAck returns tag for ShardReplicationAck
func ShardReplicationAck(shardReplicationAck int64) Tag {
	return newInt64("shard-replication-ack", shardReplicationAck)
}

// ShardReplicationToken returns information about a particular replication request
func ShardReplicationToken(token interface{}) Tag {
	return newObjectTag("shard-replication-token", token)
}

// PreviousShardRangeID returns tag for PreviousShardRangeID
func PreviousShardRangeID(id int64) Tag {
	return newInt64("previous-shard-range-id", id)
}

// ShardRangeID returns tag for ShardRangeID
func ShardRangeID(id int64) Tag {
	return newInt64("shard-range-id", id)
}

// ReadLevel returns tag for ReadLevel
func ReadLevel(lv int64) Tag {
	return newInt64("read-level", lv)
}

// MinLevel returns tag for MinLevel
func MinLevel(lv int64) Tag {
	return newInt64("min-level", lv)
}

// MaxLevel returns tag for MaxLevel
func MaxLevel(lv int64) Tag {
	return newInt64("max-level", lv)
}

// ShardTransferAcks returns tag for ShardTransferAcks
func ShardTransferAcks(shardTransferAcks interface{}) Tag {
	return newObjectTag("shard-transfer-acks", shardTransferAcks)
}

// ShardTimerAcks returns tag for ShardTimerAcks
func ShardTimerAcks(shardTimerAcks interface{}) Tag {
	return newObjectTag("shard-timer-acks", shardTimerAcks)
}

// ShardCrossClusterAcks returns tag for ShardCrossClusterAcks
func ShardCrossClusterAcks(shardCrossClusterAcks interface{}) Tag {
	return newObjectTag("shard-cross-cluster-acks", shardCrossClusterAcks)
}

// task queue processor

// QueueLevel returns tag for QueueLevel
func QueueLevel(level int) Tag {
	return newInt("queue-level", level)
}

// PreviousQueueLevel returns tag for PreviousQueueLevel
func PreviousQueueLevel(level int) Tag {
	return newInt("previous-queue-level", level)
}

// QueueSplitPolicyType returns tag for QueueSplitPolicyType
func QueueSplitPolicyType(policyType int) Tag {
	return newInt("split-policy-type", policyType)
}

// TaskID returns tag for TaskID
func TaskID(taskID int64) Tag {
	return newInt64("queue-task-id", taskID)
}

// TaskType returns tag for TaskType for queue processor
func TaskType(taskType int) Tag {
	return newInt("queue-task-type", taskType)
}

// TaskVisibilityTimestamp returns tag for task visibilityTimestamp
func TaskVisibilityTimestamp(timestamp int64) Tag {
	return newInt64("queue-task-visibility-timestamp", timestamp)
}

// NumberProcessed returns tag for NumberProcessed
func NumberProcessed(n int) Tag {
	return newInt("number-processed", n)
}

// NumberDeleted returns tag for NumberDeleted
func NumberDeleted(n int) Tag {
	return newInt("number-deleted", n)
}

// TimerTaskStatus returns tag for TimerTaskStatus
func TimerTaskStatus(timerTaskStatus int32) Tag {
	return newInt32("timer-task-status", timerTaskStatus)
}

// retry

// Attempt returns tag for Attempt
func Attempt(attempt int32) Tag {
	return newInt32("attempt", attempt)
}

// AttemptCount returns tag for AttemptCount
func AttemptCount(attemptCount int) Tag {
	return newInt("attempt-count", attemptCount)
}

// AttemptStart returns tag for AttemptStart
func AttemptStart(attemptStart time.Time) Tag {
	return newTimeTag("attempt-start", attemptStart)
}

// AttemptEnd returns tag for AttemptEnd
func AttemptEnd(attemptEnd time.Time) Tag {
	return newTimeTag("attempt-end", attemptEnd)
}

// ScheduleAttempt returns tag for ScheduleAttempt
func ScheduleAttempt(scheduleAttempt int64) Tag {
	return newInt64("schedule-attempt", scheduleAttempt)
}

// ElasticSearch

// ESRequest returns tag for ESRequest
func ESRequest(ESRequest string) Tag {
	return newStringTag("es-request", ESRequest)
}

// ESResponseStatus returns tag for ESResponse status
func ESResponseStatus(status int) Tag {
	return newInt("es-response-status", status)
}

// ESResponseError returns tag for ESResponse error
func ESResponseError(msg string) Tag {
	return newStringTag("es-response-error", msg)
}

// ESKey returns tag for ESKey
func ESKey(ESKey string) Tag {
	return newStringTag("es-mapping-key", ESKey)
}

// ESValue returns tag for ESValue
func ESValue(ESValue []byte) Tag {
	// convert value to string type so that the value logged is human readable
	return newStringTag("es-mapping-value", string(ESValue))
}

// ESConfig returns tag for ESConfig
func ESConfig(c interface{}) Tag {
	return newObjectTag("es-config", c)
}

// ESField returns tag for ESField
func ESField(ESField string) Tag {
	return newStringTag("es-field", ESField)
}

// ESDocID returns tag for ESDocID
func ESDocID(id string) Tag {
	return newStringTag("es-doc-id", id)
}

// ESAggregationID returns tag for ESDocID
func ESAggregationID(id string) Tag {
	return newStringTag("es-agg-id", id)
}

// LoggingCallAtKey is reserved tag
const LoggingCallAtKey = "logging-call-at"

// SysStackTrace returns tag for SysStackTrace
func SysStackTrace(stackTrace string) Tag {
	return newStringTag("sys-stack-trace", stackTrace)
}

// Kafka related

// KafkaTopicName returns tag for TopicName
func KafkaTopicName(topicName string) Tag {
	return newStringTag("kafka-topic-name", topicName)
}

// KafkaConsumerName returns tag for ConsumerName
func KafkaConsumerName(consumerName string) Tag {
	return newStringTag("kafka-consumer-name", consumerName)
}

// KafkaPartition returns tag for Partition
func KafkaPartition(partition int32) Tag {
	return newInt32("kafka-partition", partition)
}

// KafkaPartitionKey returns tag for PartitionKey
func KafkaPartitionKey(partitionKey interface{}) Tag {
	return newObjectTag("kafka-partition-key", partitionKey)
}

// KafkaOffset returns tag for Offset
func KafkaOffset(offset int64) Tag {
	return newInt64("kafka-offset", offset)
}

// TokenLastEventID returns tag for TokenLastEventID
func TokenLastEventID(id int64) Tag {
	return newInt64("token-last-event-id", id)
}

// /////////////////  XDC tags defined here: xdc- ///////////////////

// SourceCluster returns tag for SourceCluster
func SourceCluster(sourceCluster string) Tag {
	return newStringTag("xdc-source-cluster", sourceCluster)
}

func RemoteCluster(remoteCluster string) Tag {
	return newStringTag("xdc-remote-cluster", remoteCluster)
}

// PrevActiveCluster returns tag for PrevActiveCluster
func PrevActiveCluster(prevActiveCluster string) Tag {
	return newStringTag("xdc-prev-active-cluster", prevActiveCluster)
}

// FailoverMsg returns tag for FailoverMsg
func FailoverMsg(failoverMsg string) Tag {
	return newStringTag("xdc-failover-msg", failoverMsg)
}

// FailoverVersion returns tag for Version
func FailoverVersion(version int64) Tag {
	return newInt64("xdc-failover-version", version)
}

// CurrentVersion returns tag for CurrentVersion
func CurrentVersion(currentVersion int64) Tag {
	return newInt64("xdc-current-version", currentVersion)
}

// IncomingVersion returns tag for IncomingVersion
func IncomingVersion(incomingVersion int64) Tag {
	return newInt64("xdc-incoming-version", incomingVersion)
}

// ReplicationInfo returns tag for ReplicationInfo
func ReplicationInfo(replicationInfo interface{}) Tag {
	return newObjectTag("xdc-replication-info", replicationInfo)
}

// FirstEventVersion returns tag for FirstEventVersion
func FirstEventVersion(version int64) Tag {
	return newInt64("xdc-first-event-version", version)
}

// LastEventVersion returns tag for LastEventVersion
func LastEventVersion(version int64) Tag {
	return newInt64("xdc-last-event-version", version)
}

// TokenLastEventVersion returns tag for TokenLastEventVersion
func TokenLastEventVersion(version int64) Tag {
	return newInt64("xdc-token-last-event-version", version)
}

// ResponseSize returns tag for ResponseSize
func ResponseSize(size int) Tag {
	return newInt("response-size", size)
}

// ResponseTotalSize returns tag for ResponseTotalSize
func ResponseTotalSize(size int) Tag {
	return newInt("response-total-size", size)
}

// ResponseMaxSize returns tag for ResponseMaxSize
func ResponseMaxSize(size int) Tag {
	return newInt("response-max-size", size)
}

// ReplicationMessagesTotalSize returns tag for ReplicationMessagesTotalSize
// Should be used to indicate the final size of types.ReplicationMessages
func ReplicationMessagesTotalSize(size int) Tag {
	return newInt("replication-messages-total-size", size)
}

// ReplicationMessagesMaxSize returns tag for ReplicationMessagesMaxSize
// Should be used to indicate maximum allowed size of types.ReplicationMessages
func ReplicationMessagesMaxSize(size int) Tag {
	return newInt("replication-messages-max-size", size)
}

// ReplicationTaskID returns tag for ReplicationTaskID
// Should be used to indicate id of a types.ReplicationTask
func ReplicationTaskID(id int64) Tag {
	return newInt64("replication-task-id", id)
}

// ReplicationTaskCreationTime returns tag for ReplicationTaskCreationTime
// Should be used to indicate CreationTime of a types.ReplicationTask
func ReplicationTaskCreationTime(creationTime *int64) Tag {
	if creationTime == nil {
		return newStringTag("replication-task-creation-time", "nil")
	}
	return newInt64("replication-task-creation-time", *creationTime)
}

// ReplicationTaskBatchSize returns tag for task batch size
// Should be used to indicate used batch size for replication task processing
func ReplicationTaskBatchSize(batchSize int) Tag {
	return newInt("replication-task-batch-size", batchSize)
}

// /////////////////  Archival tags defined here: archival- ///////////////////
// archival request tags

// ArchivalCallerServiceName returns tag for the service name calling archival client
func ArchivalCallerServiceName(callerServiceName string) Tag {
	return newStringTag("archival-caller-service-name", callerServiceName)
}

// ArchivalArchiveAttemptedInline returns tag for whether archival is attempted inline before signal is sent.
func ArchivalArchiveAttemptedInline(archiveInline bool) Tag {
	return newBoolTag("archival-archive-attempted-inline", archiveInline)
}

// ArchivalRequestDomainID returns tag for RequestDomainID
func ArchivalRequestDomainID(requestDomainID string) Tag {
	return newStringTag("archival-request-domain-id", requestDomainID)
}

// ArchivalRequestDomainName returns tag for RequestDomainName
func ArchivalRequestDomainName(requestDomainName string) Tag {
	return newStringTag("archival-request-domain-name", requestDomainName)
}

// ArchivalRequestWorkflowID returns tag for RequestWorkflowID
func ArchivalRequestWorkflowID(requestWorkflowID string) Tag {
	return newStringTag("archival-request-workflow-id", requestWorkflowID)
}

// ArchvialRequestWorkflowType returns tag for RequestWorkflowType
func ArchvialRequestWorkflowType(requestWorkflowType string) Tag {
	return newStringTag("archival-request-workflow-type", requestWorkflowType)
}

// ArchivalRequestRunID returns tag for RequestRunID
func ArchivalRequestRunID(requestRunID string) Tag {
	return newStringTag("archival-request-run-id", requestRunID)
}

// ArchivalRequestBranchToken returns tag for RequestBranchToken
func ArchivalRequestBranchToken(requestBranchToken []byte) Tag {
	return newObjectTag("archival-request-branch-token", requestBranchToken)
}

// ArchivalRequestNextEventID returns tag for RequestNextEventID
func ArchivalRequestNextEventID(requestNextEventID int64) Tag {
	return newInt64("archival-request-next-event-id", requestNextEventID)
}

// ArchivalRequestCloseFailoverVersion returns tag for RequestCloseFailoverVersion
func ArchivalRequestCloseFailoverVersion(requestCloseFailoverVersion int64) Tag {
	return newInt64("archival-request-close-failover-version", requestCloseFailoverVersion)
}

// ArchivalRequestCloseTimestamp returns tag for RequestCloseTimestamp
func ArchivalRequestCloseTimestamp(requestCloseTimeStamp int64) Tag {
	return newInt64("archival-request-close-timestamp", requestCloseTimeStamp)
}

// ArchivalRequestCloseStatus returns tag for RequestCloseStatus
func ArchivalRequestCloseStatus(requestCloseStatus string) Tag {
	return newStringTag("archival-request-close-status", requestCloseStatus)
}

// ArchivalURI returns tag for Archival URI
func ArchivalURI(URI string) Tag {
	return newStringTag("archival-URI", URI)
}

// ArchivalArchiveFailReason returns tag for ArchivalArchiveFailReason
func ArchivalArchiveFailReason(archiveFailReason string) Tag {
	return newStringTag("archival-archive-fail-reason", archiveFailReason)
}

// ArchivalDeleteHistoryFailReason returns tag for ArchivalDeleteHistoryFailReason
func ArchivalDeleteHistoryFailReason(deleteHistoryFailReason string) Tag {
	return newStringTag("archival-delete-history-fail-reason", deleteHistoryFailReason)
}

// ArchivalVisibilityQuery returns tag for the query for getting archived visibility record
func ArchivalVisibilityQuery(query string) Tag {
	return newStringTag("archival-visibility-query", query)
}

// The following logger tags are only used by internal archiver implemention.
// TODO: move them to internal repo once cadence plugin model is in place.

// ArchivalBlobKey returns tag for BlobKey
func ArchivalBlobKey(blobKey string) Tag {
	return newStringTag("archival-blob-key", blobKey)
}

// ArchivalDeterministicConstructionCheckFailReason returns tag for ArchivalDeterministicConstructionCheckFailReason
func ArchivalDeterministicConstructionCheckFailReason(deterministicConstructionCheckFailReason string) Tag {
	return newStringTag("archival-deterministic-construction-check-fail-reason", deterministicConstructionCheckFailReason)
}

// ArchivalNonDeterministicBlobKey returns tag for randomly generated NonDeterministicBlobKey
func ArchivalNonDeterministicBlobKey(nondeterministicBlobKey string) Tag {
	return newStringTag("archival-non-deterministic-blob-key", nondeterministicBlobKey)
}

// ArchivalBlobIntegrityCheckFailReason returns tag for ArchivalBlobIntegrityCheckFailReason
func ArchivalBlobIntegrityCheckFailReason(blobIntegrityCheckFailReason string) Tag {
	return newStringTag("archival-blob-integrity-check-fail-reason", blobIntegrityCheckFailReason)
}

// ArchivalBlobstoreContextTimeout returns tag for ArchivalBlobstoreContextTimeout
func ArchivalBlobstoreContextTimeout(blobstoreContextTimeout time.Duration) Tag {
	return newDurationTag("archival-blobstore-context-timeout", blobstoreContextTimeout)
}

// VisibilityQuery returns tag for the query for getting visibility records
func VisibilityQuery(query string) Tag {
	return newStringTag("visibility-query", query)
}

// MembershipChangeEvent is a predefined tag for when logging hashring change events,
// expected to be of type membership.ChangeEvent
func MembershipChangeEvent(event interface{}) Tag {
	return newPredefinedDynamicTag("membership-change-event", event)
}

// Dynamic Uses reflection based logging for arbitrary values
// for not very performant logging
func Dynamic(key string, v interface{}) Tag {
	return newPredefinedDynamicTag(key, v)
}

func IsolationGroup(group string) Tag {
	return newStringTag("isolation-group", group)
}

func TaskLatency(duration time.Duration) Tag {
	return newDurationTag("task-latency", duration)
}

func IsolationDuration(duration time.Duration) Tag {
	return newDurationTag("isolation-duration", duration)
}

func PartitionConfig(p map[string]string) Tag {
	return newObjectTag("partition-config", p)
}

func PollerGroups(pollers []string) Tag {
	return newObjectTag("poller-isolation-groups", pollers)
}

func FallbackIsolationGroup(group string) Tag {
	return newStringTag("fallback-isolation-group", group)
}

func PollerGroupsConfiguration(pollers types.IsolationGroupConfiguration) Tag {
	return newObjectTag("poller-isolation-groups", pollers.ToPartitionList())
}

func WorkflowIDCacheSize(size int) Tag {
	return newInt("workflow-id-cache-size", size)
}

func AsyncWFQueueID(queueID string) Tag {
	return newStringTag("async-wf-queue-id", queueID)
}

func AsyncWFRequestType(requestType string) Tag {
	return newStringTag("async-wf-request-type", requestType)
}

func GlobalRatelimiterKey(globalKey string) Tag {
	return newStringTag("global-ratelimit-key", globalKey)
}
func GlobalRatelimiterKeyMode(mode string) Tag {
	return newStringTag("global-ratelimit-key-mode", mode)
}
func GlobalRatelimiterIdleCount(count int) Tag {
	return newInt("global-ratelimit-key-idle-count", count)
}
func GlobalRatelimiterCollectionName(name string) Tag {
	return newStringTag("global-ratelimit-collection", name)
}
func GlobalRatelimiterPeer(peer string) Tag {
	return newStringTag("global-ratelimit-peer", peer)
}

func CurrentQPS(qps float64) Tag {
	return newFloat64Tag("current-qps", qps)
}

func NumReadPartitions(n int) Tag {
	return newInt("num-read-partitions", n)
}

func NumWritePartitions(n int) Tag {
	return newInt("num-write-partitions", n)
}

func ReadChanged(b bool) Tag {
	return newBoolTag("read-changed", b)
}

func WriteChanged(b bool) Tag {
	return newBoolTag("write-changed", b)
}

func IsolationChanged(b bool) Tag {
	return newBoolTag("isolation-changed", b)
}

func CurrentNumReadPartitions(n int) Tag {
	return newInt("current-num-read-partitions", n)
}

func CurrentNumWritePartitions(n int) Tag {
	return newInt("current-num-write-partitions", n)
}

func PartitionUpscaleThreshold(qps float64) Tag {
	return newFloat64Tag("partition-upscale-threshold", qps)
}

func PartitionDownscaleThreshold(qps float64) Tag {
	return newFloat64Tag("partition-downscale-threshold", qps)
}

func PartitionDownscaleFactor(qps float64) Tag {
	return newFloat64Tag("partition-downscale-factor", qps)
}

func MatchingTaskID(id int64) Tag {
	return newInt64("matching-task-id", id)
}

func MatchingTaskScheduleID(id int64) Tag {
	return newInt64("matching-task-schedule-id", id)
}

func DecisionTaskState(state int32) Tag {
	return newInt32("decision-task-state", state)
}

func ActivityTaskState(state int32) Tag {
	return newInt32("activity-task-state", state)
}

func ShardNamespace(name string) Tag {
	return newStringTag("shard-namespace", name)
}

func ShardExecutor(ID string) Tag {
	return newStringTag("shard-executor", ID)
}

func ShardExecutors(executorIDs []string) Tag {
	return newStringsTag("shard-executors", executorIDs)
}

func ElectionDelay(t time.Duration) Tag {
	return newDurationTag("election-delay", t)
}

func WorkflowContextLockLatency(duration time.Duration) Tag {
	return newDurationTag("workflow-context-lock-latency", duration)
}

// DynamicConfigLinearIteratorSpec is a predefined tag to log dynamic config linear iterator spec
func DynamicConfigLinearIteratorSpec(spec interface{}) Tag {
	return newObjectTag("dynamic-config-linear-iterator-spec", spec)
}

func HashRingResult(addr string) Tag {
	return newStringTag("hashring-result", addr)
}

func ShardDistributorResult(addr string) Tag {
	return newStringTag("shard-distributor-result", addr)
}

// PeerHostname returns a tag for peer hostname
func PeerHostname(hostname string) Tag {
	return newStringTag("peer-hostname", hostname)
}

// PendingTaskCount returns a tag for pending task count
func PendingTaskCount(count int) Tag {
	return newInt("pending-task-count", count)
}

// VirtualQueueID returns a tag for virtual queue id
func VirtualQueueID(id int64) Tag {
	return newInt64("virtual-queue-id", id)
}
