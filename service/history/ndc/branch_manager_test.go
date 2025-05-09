// Copyright (c) 2019 Uber Technologies, Inc.
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

package ndc

import (
	ctx "context"
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
)

type (
	branchManagerSuite struct {
		suite.Suite
		*require.Assertions

		controller       *gomock.Controller
		mockShard        *shard.TestContext
		mockContext      *execution.MockContext
		mockMutableState *execution.MockMutableState

		mockHistoryV2Manager *mocks.HistoryV2Manager

		logger log.Logger

		branchIndex int
		domainID    string
		workflowID  string
		runID       string

		branchManager *branchManagerImpl
	}
)

func TestBranchManagerSuite(t *testing.T) {
	s := new(branchManagerSuite)
	suite.Run(t, s)
}

func (s *branchManagerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockContext = execution.NewMockContext(s.controller)
	s.mockMutableState = execution.NewMockMutableState(s.controller)

	s.mockShard = shard.NewTestContext(
		s.T(),
		s.controller,
		&persistence.ShardInfo{
			ShardID:          10,
			RangeID:          1,
			TransferAckLevel: 0,
		},
		config.NewForTest(),
	)

	s.mockHistoryV2Manager = s.mockShard.Resource.HistoryMgr

	s.logger = s.mockShard.GetLogger()

	s.domainID = uuid.New()
	s.workflowID = "some random workflow ID"
	s.runID = uuid.New()
	s.branchIndex = 0
	s.branchManager = newBranchManager(
		s.mockShard, s.mockContext, s.mockMutableState, s.logger,
	).(*branchManagerImpl)
}

func (s *branchManagerSuite) TearDownTest() {
	s.controller.Finish()
	s.mockShard.Finish(s.T())
}

func (s *branchManagerSuite) TestCreateNewBranch() {
	baseBranchToken := []byte("some random base branch token")
	baseBranchLCAEventVersion := int64(200)
	baseBranchLCAEventID := int64(1394)
	baseBranchLastEventVersion := int64(400)
	baseBranchLastEventID := int64(2333)
	domainName := "test-domain-name"
	versionHistory := persistence.NewVersionHistory(baseBranchToken, []*persistence.VersionHistoryItem{
		persistence.NewVersionHistoryItem(10, 0),
		persistence.NewVersionHistoryItem(50, 100),
		persistence.NewVersionHistoryItem(baseBranchLCAEventID, baseBranchLCAEventVersion),
		persistence.NewVersionHistoryItem(baseBranchLastEventID, baseBranchLastEventVersion),
	})
	versionHistories := persistence.NewVersionHistories(versionHistory)

	newBranchToken := []byte("some random new branch token")
	newVersionHistory, err := versionHistory.DuplicateUntilLCAItem(
		persistence.NewVersionHistoryItem(baseBranchLCAEventID, baseBranchLCAEventVersion),
	)
	s.NoError(err)

	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistence.WorkflowExecutionInfo{
		DomainID:   s.domainID,
		WorkflowID: s.workflowID,
		RunID:      s.runID,
	}).AnyTimes()
	s.mockShard.Resource.DomainCache.EXPECT().GetDomainName(gomock.Any()).Return(domainName, nil).AnyTimes()
	s.mockHistoryV2Manager.On("ForkHistoryBranch", mock.Anything, mock.MatchedBy(func(input *persistence.ForkHistoryBranchRequest) bool {
		input.Info = ""
		s.Equal(&persistence.ForkHistoryBranchRequest{
			ForkBranchToken: baseBranchToken,
			ForkNodeID:      baseBranchLCAEventID + 1,
			Info:            "",
			ShardID:         common.IntPtr(s.mockShard.GetShardID()),
			DomainName:      domainName,
		}, input)
		return true
	})).Return(&persistence.ForkHistoryBranchResponse{
		NewBranchToken: newBranchToken,
	}, nil).Once()

	newIndex, err := s.branchManager.createNewBranch(ctx.Background(), baseBranchToken, baseBranchLCAEventID, newVersionHistory)
	s.Nil(err)
	s.Equal(1, newIndex)

	compareVersionHistory, err := versionHistory.DuplicateUntilLCAItem(
		persistence.NewVersionHistoryItem(baseBranchLCAEventID, baseBranchLCAEventVersion),
	)
	s.NoError(err)
	s.NoError(compareVersionHistory.SetBranchToken(newBranchToken))
	newVersionHistory, err = versionHistories.GetVersionHistory(newIndex)
	s.NoError(err)
	s.True(compareVersionHistory.Equals(newVersionHistory))
}

func (s *branchManagerSuite) TestFlushBufferedEvents() {

	lastWriteVersion := int64(300)
	versionHistory := persistence.NewVersionHistory([]byte("some random base branch token"), []*persistence.VersionHistoryItem{
		persistence.NewVersionHistoryItem(10, 0),
		persistence.NewVersionHistoryItem(50, 100),
		persistence.NewVersionHistoryItem(100, 200),
		persistence.NewVersionHistoryItem(150, 300),
	})
	versionHistories := persistence.NewVersionHistories(versionHistory)

	incomingVersionHistory := versionHistory.Duplicate()
	err := incomingVersionHistory.AddOrUpdateItem(
		persistence.NewVersionHistoryItem(200, 300),
	)
	s.NoError(err)

	s.mockMutableState.EXPECT().GetLastWriteVersion().Return(lastWriteVersion, nil).AnyTimes()
	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockMutableState.EXPECT().HasBufferedEvents().Return(true).AnyTimes()
	s.mockMutableState.EXPECT().IsWorkflowExecutionRunning().Return(true).AnyTimes()
	s.mockMutableState.EXPECT().UpdateCurrentVersion(lastWriteVersion, true).Return(nil).Times(1)
	decisionInfo := &execution.DecisionInfo{
		ScheduleID: 1234,
		StartedID:  2345,
	}
	s.mockMutableState.EXPECT().GetInFlightDecision().Return(decisionInfo, true).Times(1)
	// GetExecutionInfo's return value is not used by this test
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistence.WorkflowExecutionInfo{}).AnyTimes()
	s.mockMutableState.EXPECT().AddDecisionTaskFailedEvent(
		decisionInfo.ScheduleID,
		decisionInfo.StartedID,
		types.DecisionTaskFailedCauseFailoverCloseDecision,
		[]byte(nil),
		execution.IdentityHistoryService,
		"",
		"",
		"",
		"",
		int64(0),
		"",
	).Return(&types.HistoryEvent{}, nil).Times(1)
	s.mockMutableState.EXPECT().FlushBufferedEvents().Return(nil).Times(1)
	s.mockMutableState.EXPECT().HasPendingDecision().Return(false).Times(1)
	s.mockMutableState.EXPECT().AddDecisionTaskScheduledEvent(false).Return(&execution.DecisionInfo{}, nil).Times(1)

	s.mockContext.EXPECT().UpdateWorkflowExecutionAsActive(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	ctx := ctx.Background()

	_, _, err = s.branchManager.flushBufferedEvents(ctx, incomingVersionHistory)
	s.NoError(err)
}

func (s *branchManagerSuite) TestPrepareVersionHistory_BranchAppendable_NoMissingEventInBetween() {

	versionHistory := persistence.NewVersionHistory([]byte("some random base branch token"), []*persistence.VersionHistoryItem{
		persistence.NewVersionHistoryItem(10, 0),
		persistence.NewVersionHistoryItem(50, 100),
		persistence.NewVersionHistoryItem(100, 200),
		persistence.NewVersionHistoryItem(150, 300),
	})
	versionHistories := persistence.NewVersionHistories(versionHistory)

	incomingVersionHistory := versionHistory.Duplicate()
	err := incomingVersionHistory.AddOrUpdateItem(
		persistence.NewVersionHistoryItem(200, 300),
	)
	s.NoError(err)

	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockMutableState.EXPECT().HasBufferedEvents().Return(false).AnyTimes()

	doContinue, index, err := s.branchManager.prepareVersionHistory(
		ctx.Background(),
		incomingVersionHistory,
		150+1,
		300)
	s.NoError(err)
	s.True(doContinue)
	s.Equal(0, index)
}

func (s *branchManagerSuite) TestPrepareVersionHistory_BranchAppendable_MissingEventInBetween() {

	versionHistory := persistence.NewVersionHistory([]byte("some random base branch token"), []*persistence.VersionHistoryItem{
		persistence.NewVersionHistoryItem(10, 0),
		persistence.NewVersionHistoryItem(50, 100),
		persistence.NewVersionHistoryItem(100, 200),
		persistence.NewVersionHistoryItem(150, 300),
	})
	versionHistories := persistence.NewVersionHistories(versionHistory)

	incomingVersionHistory := versionHistory.Duplicate()
	incomingFirstEventVersionHistoryItem := persistence.NewVersionHistoryItem(200, 300)
	err := incomingVersionHistory.AddOrUpdateItem(
		incomingFirstEventVersionHistoryItem,
	)
	s.NoError(err)

	execution := &persistence.WorkflowExecutionInfo{
		DomainID:   s.domainID,
		WorkflowID: s.workflowID,
		RunID:      s.runID,
	}
	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockMutableState.EXPECT().HasBufferedEvents().Return(false).AnyTimes()
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(execution).AnyTimes()

	_, _, err = s.branchManager.prepareVersionHistory(
		ctx.Background(),
		incomingVersionHistory,
		150+2,
		300)
	s.IsType(&types.RetryTaskV2Error{}, err)
}

func (s *branchManagerSuite) TestPrepareVersionHistory_BranchNotAppendable_NoMissingEventInBetween() {
	baseBranchToken := []byte("some random base branch token")
	baseBranchLCAEventID := int64(85)
	baseBranchLCAEventVersion := int64(200)
	domainName := "test-domainName"

	versionHistory := persistence.NewVersionHistory(baseBranchToken, []*persistence.VersionHistoryItem{
		persistence.NewVersionHistoryItem(10, 0),
		persistence.NewVersionHistoryItem(50, 100),
		persistence.NewVersionHistoryItem(baseBranchLCAEventID+10, baseBranchLCAEventVersion),
		persistence.NewVersionHistoryItem(150, 300),
	})
	versionHistories := persistence.NewVersionHistories(versionHistory)

	incomingVersionHistory := persistence.NewVersionHistory(nil, []*persistence.VersionHistoryItem{
		persistence.NewVersionHistoryItem(10, 0),
		persistence.NewVersionHistoryItem(50, 100),
		persistence.NewVersionHistoryItem(baseBranchLCAEventID, baseBranchLCAEventVersion),
		persistence.NewVersionHistoryItem(200, 400),
	})

	newBranchToken := []byte("some random new branch token")

	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockMutableState.EXPECT().HasBufferedEvents().Return(false).AnyTimes()
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistence.WorkflowExecutionInfo{
		DomainID:   s.domainID,
		WorkflowID: s.workflowID,
		RunID:      s.runID,
	}).AnyTimes()
	s.mockShard.Resource.DomainCache.EXPECT().GetDomainName(gomock.Any()).Return(domainName, nil).AnyTimes()
	s.mockHistoryV2Manager.On("ForkHistoryBranch", mock.Anything, mock.MatchedBy(func(input *persistence.ForkHistoryBranchRequest) bool {
		input.Info = ""
		s.Equal(&persistence.ForkHistoryBranchRequest{
			ForkBranchToken: baseBranchToken,
			ForkNodeID:      baseBranchLCAEventID + 1,
			Info:            "",
			ShardID:         common.IntPtr(s.mockShard.GetShardID()),
			DomainName:      domainName,
		}, input)
		return true
	})).Return(&persistence.ForkHistoryBranchResponse{
		NewBranchToken: newBranchToken,
	}, nil).Once()

	doContinue, index, err := s.branchManager.prepareVersionHistory(
		ctx.Background(),
		incomingVersionHistory,
		baseBranchLCAEventID+1,
		baseBranchLCAEventVersion,
	)
	s.NoError(err)
	s.True(doContinue)
	s.Equal(1, index)
}

func (s *branchManagerSuite) TestPrepareVersionHistory_BranchNotAppendable_MissingEventInBetween() {
	baseBranchToken := []byte("some random base branch token")
	baseBranchLCAEventID := int64(85)
	baseBranchLCAEventVersion := int64(200)
	baseBranchLastEventID := int64(150)
	baseBranchLastEventVersion := int64(300)

	versionHistory := persistence.NewVersionHistory(baseBranchToken, []*persistence.VersionHistoryItem{
		persistence.NewVersionHistoryItem(10, 0),
		persistence.NewVersionHistoryItem(50, 100),
		persistence.NewVersionHistoryItem(baseBranchLCAEventID+10, baseBranchLCAEventVersion),
		persistence.NewVersionHistoryItem(baseBranchLastEventID, baseBranchLastEventVersion),
	})
	versionHistories := persistence.NewVersionHistories(versionHistory)

	incomingVersionHistory := persistence.NewVersionHistory(nil, []*persistence.VersionHistoryItem{
		persistence.NewVersionHistoryItem(10, 0),
		persistence.NewVersionHistoryItem(50, 100),
		persistence.NewVersionHistoryItem(baseBranchLCAEventID, baseBranchLCAEventVersion),
		persistence.NewVersionHistoryItem(200, 400),
	})

	execution := &persistence.WorkflowExecutionInfo{
		DomainID:   s.domainID,
		WorkflowID: s.workflowID,
		RunID:      s.runID,
	}
	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockMutableState.EXPECT().HasBufferedEvents().Return(false).AnyTimes()
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(execution).AnyTimes()

	_, _, err := s.branchManager.prepareVersionHistory(
		ctx.Background(),
		incomingVersionHistory,
		baseBranchLCAEventID+2,
		baseBranchLCAEventVersion,
	)
	s.IsType(&types.RetryTaskV2Error{}, err)
}
