// Code generated by MockGen. DO NOT EDIT.
// Source: state_builder.go
//
// Generated by this command:
//
//	mockgen -package execution -source state_builder.go -destination state_builder_mock.go -self_package github.com/uber/cadence/service/history/execution
//

// Package execution is a generated GoMock package.
package execution

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"

	types "github.com/uber/cadence/common/types"
)

// MockStateBuilder is a mock of StateBuilder interface.
type MockStateBuilder struct {
	ctrl     *gomock.Controller
	recorder *MockStateBuilderMockRecorder
	isgomock struct{}
}

// MockStateBuilderMockRecorder is the mock recorder for MockStateBuilder.
type MockStateBuilderMockRecorder struct {
	mock *MockStateBuilder
}

// NewMockStateBuilder creates a new mock instance.
func NewMockStateBuilder(ctrl *gomock.Controller) *MockStateBuilder {
	mock := &MockStateBuilder{ctrl: ctrl}
	mock.recorder = &MockStateBuilderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStateBuilder) EXPECT() *MockStateBuilderMockRecorder {
	return m.recorder
}

// ApplyEvents mocks base method.
func (m *MockStateBuilder) ApplyEvents(domainID, requestID string, workflowExecution types.WorkflowExecution, history, newRunHistory []*types.HistoryEvent) (MutableState, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyEvents", domainID, requestID, workflowExecution, history, newRunHistory)
	ret0, _ := ret[0].(MutableState)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ApplyEvents indicates an expected call of ApplyEvents.
func (mr *MockStateBuilderMockRecorder) ApplyEvents(domainID, requestID, workflowExecution, history, newRunHistory any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyEvents", reflect.TypeOf((*MockStateBuilder)(nil).ApplyEvents), domainID, requestID, workflowExecution, history, newRunHistory)
}

// GetMutableState mocks base method.
func (m *MockStateBuilder) GetMutableState() MutableState {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMutableState")
	ret0, _ := ret[0].(MutableState)
	return ret0
}

// GetMutableState indicates an expected call of GetMutableState.
func (mr *MockStateBuilderMockRecorder) GetMutableState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMutableState", reflect.TypeOf((*MockStateBuilder)(nil).GetMutableState))
}
