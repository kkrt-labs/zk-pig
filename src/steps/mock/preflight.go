// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/kkrt-labs/zk-pig/src/steps (interfaces: Preflight)
//
// Generated by this command:
//
//	mockgen -destination=./mock/preflight.go -package=mocksteps github.com/kkrt-labs/zk-pig/src/steps Preflight
//

// Package mocksteps is a generated GoMock package.
package mocksteps

import (
	context "context"
	reflect "reflect"

	types "github.com/ethereum/go-ethereum/core/types"
	steps "github.com/kkrt-labs/zk-pig/src/steps"
	gomock "go.uber.org/mock/gomock"
)

// MockPreflight is a mock of Preflight interface.
type MockPreflight struct {
	ctrl     *gomock.Controller
	recorder *MockPreflightMockRecorder
	isgomock struct{}
}

// MockPreflightMockRecorder is the mock recorder for MockPreflight.
type MockPreflightMockRecorder struct {
	mock *MockPreflight
}

// NewMockPreflight creates a new mock instance.
func NewMockPreflight(ctrl *gomock.Controller) *MockPreflight {
	mock := &MockPreflight{ctrl: ctrl}
	mock.recorder = &MockPreflightMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPreflight) EXPECT() *MockPreflightMockRecorder {
	return m.recorder
}

// Preflight mocks base method.
func (m *MockPreflight) Preflight(ctx context.Context, block *types.Block) (*steps.PreflightData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Preflight", ctx, block)
	ret0, _ := ret[0].(*steps.PreflightData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Preflight indicates an expected call of Preflight.
func (mr *MockPreflightMockRecorder) Preflight(ctx, block any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Preflight", reflect.TypeOf((*MockPreflight)(nil).Preflight), ctx, block)
}
