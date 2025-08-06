// Copyright (c) 2018 Uber Technologies, Inc.
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

/*
To run locally:

1. Pick a scenario from the existing config files simulation/replication/testdata/replication_simulation_${scenario}.yaml or add a new one

2. Run the scenario
`./simulation/replication/run.sh default`

Full test logs can be found at test.log file. Event json logs can be found at replication-simulator-output folder.
See the run.sh script for more details about how to parse events.
*/
package replication

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
	simTypes "github.com/uber/cadence/simulation/replication/types"
)

func TestReplicationSimulation(t *testing.T) {
	flag.Parse()

	simTypes.Logf(t, "Starting Replication Simulation")

	simTypes.Logf(t, "Sleeping for 30 seconds to allow services to start/warmup")
	time.Sleep(30 * time.Second)

	// load config
	simCfg, err := simTypes.LoadConfig()
	require.NoError(t, err, "failed to load config")

	// initialize cadence clients
	for clusterName := range simCfg.Clusters {
		simCfg.MustInitClientsFor(t, clusterName)
	}

	simTypes.Logf(t, "Registering domains")
	for domainName, domainCfg := range simCfg.Domains {
		simTypes.Logf(t, "Domain: %s", domainName)
		simCfg.MustRegisterDomain(t, domainName, domainCfg)
	}

	// wait for domain data to be replicated and workers to start.
	waitUntilWorkersReady(t)

	sort.Slice(simCfg.Operations, func(i, j int) bool {
		return simCfg.Operations[i].At < simCfg.Operations[j].At
	})

	startTime := time.Now().UTC()
	simTypes.Logf(t, "Simulation start time: %v", startTime)
	for i, op := range simCfg.Operations {
		op := op
		waitForOpTime(t, op, startTime)
		var err error
		switch op.Type {
		case simTypes.ReplicationSimulationOperationStartWorkflow:
			err = startWorkflow(t, op, simCfg)
		case simTypes.ReplicationSimulationOperationResetWorkflow:
			err = resetWorkflow(t, op, simCfg)
		case simTypes.ReplicationSimulationOperationChangeActiveClusters:
			err = changeActiveClusters(t, op, simCfg)
		case simTypes.ReplicationSimulationOperationValidate:
			err = validate(t, op, simCfg)
		case simTypes.ReplicationSimulationOperationQueryWorkflow:
			err = queryWorkflow(t, op, simCfg)
		case simTypes.ReplicationSimulationOperationSignalWithStartWorkflow:
			err = signalWithStartWorkflow(t, op, simCfg)
		case simTypes.ReplicationSimulationOperationMigrateDomainToActiveActive:
			err = migrateDomainToActiveActive(t, op, simCfg)
		default:
			require.Failf(t, "unknown operation type", "operation type: %s", op.Type)
		}

		if err != nil {
			t.Fatalf("Operation %d failed: %v", i, err)
		}
	}

	// Print the test summary.
	// Don't change the start/end line format as it is used by scripts to parse the summary info
	executionTime := time.Since(startTime)
	var testSummary []string
	testSummary = append(testSummary, "Simulation Summary:")
	testSummary = append(testSummary, fmt.Sprintf("Simulation Duration: %v", executionTime))
	testSummary = append(testSummary, "End of Simulation Summary")
	fmt.Println(strings.Join(testSummary, "\n"))
}

func startWorkflow(
	t *testing.T,
	op *simTypes.Operation,
	simCfg *simTypes.ReplicationSimulationConfig,
) error {
	t.Helper()

	simTypes.Logf(t, "Starting workflow: %s on domain %s on cluster: %s", op.WorkflowID, op.Domain, op.Cluster)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if op.WorkflowExecutionStartToCloseTimeout == 0 || op.WorkflowExecutionStartToCloseTimeout < op.WorkflowDuration {
		return fmt.Errorf("workflow execution start to close timeout must be specified and should be greater than workflow duration")
	}

	resp, err := simCfg.MustGetFrontendClient(t, op.Cluster).StartWorkflowExecution(ctx,
		&types.StartWorkflowExecutionRequest{
			RequestID:                           uuid.New(),
			Domain:                              op.Domain,
			WorkflowID:                          op.WorkflowID,
			WorkflowType:                        &types.WorkflowType{Name: op.WorkflowType},
			TaskList:                            &types.TaskList{Name: simTypes.TasklistName},
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(int32((op.WorkflowExecutionStartToCloseTimeout).Seconds())),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(5),
			Input:                               mustJSON(t, &simTypes.WorkflowInput{Duration: op.WorkflowDuration, ActivityCount: op.ActivityCount}),
			WorkflowIDReusePolicy:               types.WorkflowIDReusePolicyAllowDuplicate.Ptr(),
			DelayStartSeconds:                   common.Int32Ptr(op.DelayStartSeconds),
		})

	if err != nil {
		return err
	}

	simTypes.Logf(t, "Started workflow: %s on domain: %s on cluster: %s. RunID: %s", op.WorkflowID, op.Domain, op.Cluster, resp.GetRunID())

	return nil
}

func resetWorkflow(
	t *testing.T,
	op *simTypes.Operation,
	simCfg *simTypes.ReplicationSimulationConfig,
) error {
	t.Helper()

	simTypes.Logf(t, "Resetting workflow: %s on domain %s on cluster: %s", op.WorkflowID, op.Domain, op.Cluster)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := simCfg.MustGetFrontendClient(t, op.Cluster).DescribeWorkflowExecution(ctx,
		&types.DescribeWorkflowExecutionRequest{
			Domain: op.Domain,
			Execution: &types.WorkflowExecution{
				WorkflowID: op.WorkflowID,
			},
		})

	if err != nil {
		return fmt.Errorf("failed to describe workflow %s: %w", op.WorkflowID, err)
	}

	simTypes.Logf(t, "DescribeWorkflowExecution workflowID: %s on domain: %s on cluster: %s. RunID: %s", op.WorkflowID, op.Domain, op.Cluster, resp.GetWorkflowExecutionInfo().Execution.RunID)

	resetResp, err := simCfg.MustGetFrontendClient(t, op.Cluster).ResetWorkflowExecution(ctx,
		&types.ResetWorkflowExecutionRequest{
			Domain: op.Domain,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: op.WorkflowID,
				RunID:      resp.GetWorkflowExecutionInfo().Execution.RunID,
			},
			Reason:                "resetting workflow",
			DecisionFinishEventID: op.EventID,
		})

	if err != nil {
		if op.Want.Error != "" {
			if strings.Contains(err.Error(), op.Want.Error) {
				simTypes.Logf(t, "Expected error: %s", op.Want.Error)
				return nil
			}
			return fmt.Errorf("expected error: %s, but got: %s", op.Want.Error, err.Error())
		}
		simTypes.Logf(t, err.Error())
		return err
	}

	if op.Want.Error != "" {
		return fmt.Errorf("expected error: %s, but got nil", op.Want.Error)
	}

	simTypes.Logf(t, "Reset workflowID: %s on domain: %s on cluster: %s. RunID: %s", op.WorkflowID, op.Domain, op.Cluster, resetResp.GetRunID())

	return nil
}

func changeActiveClusters(
	t *testing.T,
	op *simTypes.Operation,
	simCfg *simTypes.ReplicationSimulationConfig,
) error {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	descResp, err := simCfg.MustGetFrontendClient(t, simCfg.PrimaryCluster).DescribeDomain(ctx, &types.DescribeDomainRequest{Name: common.StringPtr(op.Domain)})
	if err != nil {
		return fmt.Errorf("failed to describe domain %s: %w", op.Domain, err)
	}

	if !simCfg.IsActiveActiveDomain(op.Domain) {
		fromCluster := descResp.ReplicationConfiguration.ActiveClusterName
		toCluster := op.NewActiveCluster
		simTypes.Logf(t, "Changing active clusters for domain %s from %s to %s", op.Domain, fromCluster, toCluster)

		ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err = simCfg.MustGetFrontendClient(t, simCfg.PrimaryCluster).UpdateDomain(ctx,
			&types.UpdateDomainRequest{
				Name:                     op.Domain,
				ActiveClusterName:        &toCluster,
				FailoverTimeoutInSeconds: op.FailoverTimeout,
			})
		if err != nil {
			return fmt.Errorf("failed to update ActiveClusterName, err: %w", err)
		}

		simTypes.Logf(t, "Failed over from %s to %s", fromCluster, toCluster)
	} else {

		ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		activeClustersByRegion := make(map[string]types.ActiveClusterInfo)
		for region, cluster := range op.NewActiveClustersByRegion {
			activeClustersByRegion[region] = types.ActiveClusterInfo{
				ActiveClusterName: cluster,
				FailoverVersion:   -1, // doesn't matter. API handler will override it
			}
		}
		ac := &types.ActiveClusters{
			ActiveClustersByRegion: activeClustersByRegion,
		}
		simTypes.Logf(t, "Changing active clusters by region for domain %s from %+v to %+v", op.Domain, descResp.ReplicationConfiguration.ActiveClusters, ac)
		_, err = simCfg.MustGetFrontendClient(t, simCfg.PrimaryCluster).UpdateDomain(ctx,
			&types.UpdateDomainRequest{
				Name:           op.Domain,
				ActiveClusters: ac,
			})
		if err != nil {
			return fmt.Errorf("failed to update ActiveClusters, err: %w", err)
		}

		simTypes.Logf(t, "Failed over from %+v to %+v", descResp.ReplicationConfiguration.ActiveClusters, ac)
	}
	return nil
}

func migrateDomainToActiveActive(t *testing.T, op *simTypes.Operation, simCfg *simTypes.ReplicationSimulationConfig) error {
	t.Helper()

	simTypes.Logf(t, "Migrating domain %s to active-active", op.Domain)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	frontendCl := simCfg.MustGetFrontendClient(t, simCfg.PrimaryCluster)
	ac := &types.ActiveClusters{
		ActiveClustersByRegion: make(map[string]types.ActiveClusterInfo),
	}
	for region, cluster := range op.NewActiveClustersByRegion {
		ac.ActiveClustersByRegion[region] = types.ActiveClusterInfo{
			ActiveClusterName: cluster,
			FailoverVersion:   -1, // doesn't matter. API handler will override it
		}
	}
	_, err := frontendCl.UpdateDomain(ctx, &types.UpdateDomainRequest{
		Name:           op.Domain,
		ActiveClusters: ac,
	})
	if err != nil {
		return fmt.Errorf("failed to update domain %s to active-active: %w", op.Domain, err)
	}

	simTypes.Logf(t, "Migrated domain %s to active-active", op.Domain)
	return nil
}

func queryWorkflow(t *testing.T, op *simTypes.Operation, simCfg *simTypes.ReplicationSimulationConfig) error {
	t.Helper()

	simTypes.Logf(t, "Querying workflow: %s on domain %s on cluster: %s", op.WorkflowID, op.Domain, op.Cluster)

	frontendCl := simCfg.MustGetFrontendClient(t, op.Cluster)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	consistencyLevel := types.QueryConsistencyLevelEventual.Ptr()
	if op.ConsistencyLevel == "strong" {
		consistencyLevel = types.QueryConsistencyLevelStrong.Ptr()
	}

	queryResp, err := frontendCl.QueryWorkflow(ctx, &types.QueryWorkflowRequest{
		Domain: op.Domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: op.WorkflowID,
		},
		Query: &types.WorkflowQuery{
			QueryType: op.Query,
		},
		QueryConsistencyLevel: consistencyLevel,
	})
	if err != nil {
		return err
	}

	got := mustParseJSON(t, queryResp.GetQueryResult())
	want := op.Want.QueryResult
	if !reflect.DeepEqual(want, got) {
		return fmt.Errorf("query result mismatch. want: %v (type: %T), got: %v (type: %T)", want, want, got, got)
	}

	simTypes.Logf(t, "Query workflow: %s on domain: %s on cluster: %s. Query result: %v", op.WorkflowID, op.Domain, op.Cluster, got)

	return nil
}

func signalWithStartWorkflow(
	t *testing.T,
	op *simTypes.Operation,
	simCfg *simTypes.ReplicationSimulationConfig,
) error {
	t.Helper()
	simTypes.Logf(t, "SignalWithStart workflow: %s on domain %s on cluster: %s", op.WorkflowID, op.Domain, op.Cluster)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	frontendCl := simCfg.MustGetFrontendClient(t, op.Cluster)

	signalResp, err := frontendCl.SignalWithStartWorkflowExecution(ctx, &types.SignalWithStartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              op.Domain,
		WorkflowID:                          op.WorkflowID,
		WorkflowIDReusePolicy:               types.WorkflowIDReusePolicyAllowDuplicate.Ptr(),
		WorkflowType:                        &types.WorkflowType{Name: op.WorkflowType},
		SignalName:                          op.SignalName,
		SignalInput:                         mustJSON(t, op.SignalInput),
		TaskList:                            &types.TaskList{Name: simTypes.TasklistName},
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(int32((op.WorkflowExecutionStartToCloseTimeout).Seconds())),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(5),
		Input:                               mustJSON(t, &simTypes.WorkflowInput{Duration: op.WorkflowDuration, ActivityCount: op.ActivityCount}),
	})
	if err != nil {
		return err
	}

	simTypes.Logf(t, "SignalWithStart workflow: %s on domain %s on cluster: %s. RunID: %s", op.WorkflowID, op.Domain, op.Cluster, signalResp.GetRunID())
	return nil
}

// validate performs validation based on given operation config.
// validate function does not fail the test via t.Fail (or require.X).
// It runs in separate goroutine. It should return an error.
func validate(
	t *testing.T,
	op *simTypes.Operation,
	simCfg *simTypes.ReplicationSimulationConfig,
) error {
	t.Helper()

	simTypes.Logf(t, "Validating workflow: %s on cluster: %s", op.WorkflowID, op.Cluster)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resp, err := simCfg.MustGetFrontendClient(t, op.Cluster).DescribeWorkflowExecution(ctx,
		&types.DescribeWorkflowExecutionRequest{
			Domain: op.Domain,
			Execution: &types.WorkflowExecution{
				WorkflowID: op.WorkflowID,
			},
		})
	if err != nil {
		return err
	}

	switch op.Want.Status {
	case "completed":
		// Validate workflow completed
		if resp.GetWorkflowExecutionInfo().GetCloseStatus() != types.WorkflowExecutionCloseStatusCompleted || resp.GetWorkflowExecutionInfo().GetCloseTime() == 0 {
			return fmt.Errorf("workflow %s not completed. status: %s, close time: %v", op.WorkflowID, resp.GetWorkflowExecutionInfo().GetCloseStatus(), time.Unix(0, resp.GetWorkflowExecutionInfo().GetCloseTime()))
		}
	case "failed":
		// Validate workflow failed
		if resp.GetWorkflowExecutionInfo().GetCloseStatus() != types.WorkflowExecutionCloseStatusFailed || resp.GetWorkflowExecutionInfo().GetCloseTime() == 0 {
			return fmt.Errorf("workflow %s not failed. status: %s, close time: %v", op.WorkflowID, resp.GetWorkflowExecutionInfo().GetCloseStatus(), time.Unix(0, resp.GetWorkflowExecutionInfo().GetCloseTime()))
		}
	case "canceled":
		// Validate workflow canceled
		if resp.GetWorkflowExecutionInfo().GetCloseStatus() != types.WorkflowExecutionCloseStatusCanceled || resp.GetWorkflowExecutionInfo().GetCloseTime() == 0 {
			return fmt.Errorf("workflow %s not canceled. status: %s, close time: %v", op.WorkflowID, resp.GetWorkflowExecutionInfo().GetCloseStatus(), time.Unix(0, resp.GetWorkflowExecutionInfo().GetCloseTime()))
		}
	case "terminated":
		// Validate workflow terminated
		if resp.GetWorkflowExecutionInfo().GetCloseStatus() != types.WorkflowExecutionCloseStatusTerminated || resp.GetWorkflowExecutionInfo().GetCloseTime() == 0 {
			return fmt.Errorf("workflow %s not terminated. status: %s, close time: %v", op.WorkflowID, resp.GetWorkflowExecutionInfo().GetCloseStatus(), time.Unix(0, resp.GetWorkflowExecutionInfo().GetCloseTime()))
		}
	case "continued-as-new":
		// Validate workflow continued as new
		if resp.GetWorkflowExecutionInfo().GetCloseStatus() != types.WorkflowExecutionCloseStatusContinuedAsNew || resp.GetWorkflowExecutionInfo().GetCloseTime() == 0 {
			return fmt.Errorf("workflow %s not continued as new. status: %s, close time: %v", op.WorkflowID, resp.GetWorkflowExecutionInfo().GetCloseStatus(), time.Unix(0, resp.GetWorkflowExecutionInfo().GetCloseTime()))
		}
	case "timed-out":
		// Validate workflow timed out
		if resp.GetWorkflowExecutionInfo().GetCloseStatus() != types.WorkflowExecutionCloseStatusTimedOut || resp.GetWorkflowExecutionInfo().GetCloseTime() == 0 {
			return fmt.Errorf("workflow %s not timed out. status: %s, close time: %v", op.WorkflowID, resp.GetWorkflowExecutionInfo().GetCloseStatus(), time.Unix(0, resp.GetWorkflowExecutionInfo().GetCloseTime()))
		}
	default:
		// Validate workflow is running
		if resp.GetWorkflowExecutionInfo().GetCloseTime() != 0 {
			return fmt.Errorf("workflow %s not running. status: %s, close time: %v", op.WorkflowID, resp.GetWorkflowExecutionInfo().GetCloseStatus(), time.Unix(0, resp.GetWorkflowExecutionInfo().GetCloseTime()))
		}
	}

	simTypes.Logf(t, "Validated workflow: %s on cluster: %s. Status: %s, CloseTime: %v", op.WorkflowID, op.Cluster, resp.GetWorkflowExecutionInfo().GetCloseStatus(), time.Unix(0, resp.GetWorkflowExecutionInfo().GetCloseTime()))

	// Get history to validate the worker identity that started and completed the workflow
	// Some workflows start in cluster0 and complete in cluster1. This is to validate that
	history, err := getAllHistory(t, simCfg, op.Cluster, op.Domain, op.WorkflowID)
	if err != nil {
		return err
	}

	if len(history) == 0 {
		return fmt.Errorf("no history events found for workflow %s", op.WorkflowID)
	}

	startedWorker, err := firstDecisionTaskWorker(history)
	if err != nil {
		return err
	}
	if op.Want.StartedByWorkersInCluster != "" && startedWorker != simTypes.WorkerIdentityFor(op.Want.StartedByWorkersInCluster, op.Domain) {
		return fmt.Errorf("workflow %s started by worker %s, expected %s", op.WorkflowID, startedWorker, simTypes.WorkerIdentityFor(op.Want.StartedByWorkersInCluster, op.Domain))
	}

	completedWorker, err := lastDecisionTaskWorker(history)
	if err != nil {
		return err
	}

	if op.Want.CompletedByWorkersInCluster != "" && completedWorker != simTypes.WorkerIdentityFor(op.Want.CompletedByWorkersInCluster, op.Domain) {
		return fmt.Errorf("workflow %s completed by worker %s, expected %s", op.WorkflowID, completedWorker, simTypes.WorkerIdentityFor(op.Want.CompletedByWorkersInCluster, op.Domain))
	}

	return nil
}

func firstDecisionTaskWorker(history []types.HistoryEvent) (string, error) {
	for _, event := range history {
		if event.GetEventType() == types.EventTypeDecisionTaskCompleted {
			return event.GetDecisionTaskCompletedEventAttributes().Identity, nil
		}
	}
	return "", fmt.Errorf("failed to find first decision task worker because there's no DecisionTaskCompleted event found in history")
}

func lastDecisionTaskWorker(history []types.HistoryEvent) (string, error) {
	for i := len(history) - 1; i >= 0; i-- {
		event := history[i]
		if event.GetEventType() == types.EventTypeDecisionTaskCompleted {
			return event.GetDecisionTaskCompletedEventAttributes().Identity, nil
		}
	}
	return "", fmt.Errorf("failed to find lastDecisionTaskWorker because there's no DecisionTaskCompleted event found in history")
}

func waitForOpTime(t *testing.T, op *simTypes.Operation, startTime time.Time) {
	t.Helper()
	d := startTime.Add(op.At).Sub(time.Now().UTC())
	if d > 0 {
		simTypes.Logf(t, "Waiting for next operation time (t + %ds). Will sleep for %ds", int(op.At.Seconds()), int(d.Seconds()))
		<-time.After(d)
	}

	simTypes.Logf(t, "Operation time (t + %ds) reached: %v", int(op.At.Seconds()), startTime.Add(op.At))
}

func getAllHistory(t *testing.T, simCfg *simTypes.ReplicationSimulationConfig, clusterName, domainName, wfID string) ([]types.HistoryEvent, error) {
	frontendCl := simCfg.MustGetFrontendClient(t, clusterName)
	var nextPageToken []byte
	var history []types.HistoryEvent
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		response, err := frontendCl.GetWorkflowExecutionHistory(ctx, &types.GetWorkflowExecutionHistoryRequest{
			Domain: domainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: wfID,
			},
			MaximumPageSize:        1000,
			NextPageToken:          nextPageToken,
			WaitForNewEvent:        false,
			HistoryEventFilterType: types.HistoryEventFilterTypeAllEvent.Ptr(),
			SkipArchival:           true,
		})
		cancel()
		if err != nil {
			return nil, fmt.Errorf("failed to get history: %w", err)
		}

		for _, event := range response.GetHistory().GetEvents() {
			if event != nil {
				history = append(history, *event)
			}
		}

		if response.NextPageToken == nil {
			return history, nil
		}

		nextPageToken = response.NextPageToken
		time.Sleep(10 * time.Millisecond) // sleep to avoid throttling
	}
}

func mustParseJSON(t *testing.T, v []byte) any {
	var result interface{}
	err := json.Unmarshal(v, &result)
	require.NoError(t, err, "failed to unmarshal from json")
	return result
}

func mustJSON(t *testing.T, v interface{}) []byte {
	data, err := json.Marshal(v)
	require.NoError(t, err, "failed to marshal to json")
	return data
}

func waitUntilWorkersReady(t *testing.T) {
	// workers expose :6060/health endpoint. Poll on them to check if they are healthy
	simTypes.Logf(t, "Waiting for workers to start and report healthy")
	workerEndpoints := []string{
		"http://cadence-worker0:6060/health",
		"http://cadence-worker1:6060/health",
	}

	for {
		allHealthy := true
		for _, endpoint := range workerEndpoints {
			resp, err := http.Get(endpoint)
			if err != nil || resp.StatusCode != http.StatusOK {
				allHealthy = false
				break
			}
		}

		if allHealthy {
			break
		}

		simTypes.Logf(t, "Workers are not reporting healthy yet. Sleep for 2s and try again")
		time.Sleep(2 * time.Second)
	}

	simTypes.Logf(t, "All workers are healthy")
}
