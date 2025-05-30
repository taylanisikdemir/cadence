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

package scanner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
	"go.uber.org/mock/gomock"

	"github.com/uber/cadence/common/dynamicconfig/dynamicproperties"
	"github.com/uber/cadence/common/metrics"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/service/worker/scanner/tasklist"
)

type scannerWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestScannerWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(scannerWorkflowTestSuite))
}

func (s *scannerWorkflowTestSuite) TestWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(taskListScavengerActivityName, mock.Anything).Return(nil)
	env.ExecuteWorkflow(tlScannerWFTypeName)
	s.True(env.IsWorkflowCompleted())
}

func (s *scannerWorkflowTestSuite) TestScavengerActivity() {
	env := s.NewTestActivityEnvironment()
	controller := gomock.NewController(s.T())
	mockResource := resource.NewTest(s.T(), controller, metrics.Worker)
	defer mockResource.Finish(s.T())

	mockResource.TaskMgr.On("ListTaskList", mock.Anything, mock.Anything).Return(&p.ListTaskListResponse{}, nil)
	mockResource.TaskMgr.On("GetOrphanTasks", mock.Anything, mock.Anything).Return(&p.GetOrphanTasksResponse{}, nil)
	ctx := scannerContext{
		resource: mockResource,
		cfg: Config{
			TaskListScannerOptions: tasklist.Options{
				GetOrphanTasksPageSizeFn: dynamicproperties.GetIntPropertyFn(dynamicproperties.ScannerGetOrphanTasksPageSize.DefaultInt()),
				EnableCleaning:           dynamicproperties.GetBoolPropertyFn(true),
				ExecutorPollInterval:     time.Millisecond * 50,
			},
		},
	}
	env.SetTestTimeout(time.Second * 5)
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: NewScannerContext(context.Background(), "default-test-workflow-type-name", ctx),
	})
	tlScavengerHBInterval = time.Millisecond * 10
	_, err := env.ExecuteActivity(taskListScavengerActivityName)
	s.NoError(err)
}
