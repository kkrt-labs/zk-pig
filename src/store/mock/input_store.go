// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/kkrt-labs/zk-pig/src/store (interfaces: ProverInputStore)
//
// Generated by this command:
//
//	mockgen -destination=./mock/input_store.go -package=mockstore github.com/kkrt-labs/zk-pig/src/store ProverInputStore
//

// Package mockstore is a generated GoMock package.
package mockstore

import (
	context "context"
	reflect "reflect"

	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	gomock "go.uber.org/mock/gomock"
)

// MockProverInputStore is a mock of ProverInputStore interface.
type MockProverInputStore struct {
	ctrl     *gomock.Controller
	recorder *MockProverInputStoreMockRecorder
	isgomock struct{}
}

// MockProverInputStoreMockRecorder is the mock recorder for MockProverInputStore.
type MockProverInputStoreMockRecorder struct {
	mock *MockProverInputStore
}

// NewMockProverInputStore creates a new mock instance.
func NewMockProverInputStore(ctrl *gomock.Controller) *MockProverInputStore {
	mock := &MockProverInputStore{ctrl: ctrl}
	mock.recorder = &MockProverInputStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProverInputStore) EXPECT() *MockProverInputStoreMockRecorder {
	return m.recorder
}

// LoadProverInput mocks base method.
func (m *MockProverInputStore) LoadProverInput(ctx context.Context, chainID, blockNumber uint64) (*input.ProverInput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadProverInput", ctx, chainID, blockNumber)
	ret0, _ := ret[0].(*input.ProverInput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoadProverInput indicates an expected call of LoadProverInput.
func (mr *MockProverInputStoreMockRecorder) LoadProverInput(ctx, chainID, blockNumber any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadProverInput", reflect.TypeOf((*MockProverInputStore)(nil).LoadProverInput), ctx, chainID, blockNumber)
}

// StoreProverInput mocks base method.
func (m *MockProverInputStore) StoreProverInput(ctx context.Context, inputs *input.ProverInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreProverInput", ctx, inputs)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreProverInput indicates an expected call of StoreProverInput.
func (mr *MockProverInputStoreMockRecorder) StoreProverInput(ctx, inputs any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreProverInput", reflect.TypeOf((*MockProverInputStore)(nil).StoreProverInput), ctx, inputs)
}
