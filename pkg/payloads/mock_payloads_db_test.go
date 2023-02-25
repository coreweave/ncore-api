// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/coreweave/ncore-api/pkg/payloads (interfaces: DB)

// Package payloads is a generated GoMock package.
package payloads

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockDB is a mock of DB interface.
type MockDB struct {
	ctrl     *gomock.Controller
	recorder *MockDBMockRecorder
}

// MockDBMockRecorder is the mock recorder for MockDB.
type MockDBMockRecorder struct {
	mock *MockDB
}

// NewMockDB creates a new mock instance.
func NewMockDB(ctrl *gomock.Controller) *MockDB {
	mock := &MockDB{ctrl: ctrl}
	mock.recorder = &MockDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDB) EXPECT() *MockDBMockRecorder {
	return m.recorder
}

// GetNodePayload mocks base method.
func (m *MockDB) GetNodePayload(arg0 context.Context, arg1 string) (*NodePayload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodePayload", arg0, arg1)
	ret0, _ := ret[0].(*NodePayload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNodePayload indicates an expected call of GetNodePayload.
func (mr *MockDBMockRecorder) GetNodePayload(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodePayload", reflect.TypeOf((*MockDB)(nil).GetNodePayload), arg0, arg1)
}

// GetPayloadParameters mocks base method.
func (m *MockDB) GetPayloadParameters(arg0 context.Context, arg1 string) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPayloadParameters", arg0, arg1)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPayloadParameters indicates an expected call of GetPayloadParameters.
func (mr *MockDBMockRecorder) GetPayloadParameters(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPayloadParameters", reflect.TypeOf((*MockDB)(nil).GetPayloadParameters), arg0, arg1)
}
