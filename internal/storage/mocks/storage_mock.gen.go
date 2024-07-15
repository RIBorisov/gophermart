// Code generated by MockGen. DO NOT EDIT.
// Source: internal/storage/storage.go
//
// Generated by this command:
//
//	mockgen -source=internal/storage/storage.go -destination=internal/storage/mocks/storage_mock.gen.go -package=mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	balance "github.com/RIBorisov/gophermart/internal/models/balance"
	orders "github.com/RIBorisov/gophermart/internal/models/orders"
	register "github.com/RIBorisov/gophermart/internal/models/register"
	storage "github.com/RIBorisov/gophermart/internal/storage"
	gomock "go.uber.org/mock/gomock"
)

// MockStore is a mock of Store interface.
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore.
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance.
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// BalanceWithdraw mocks base method.
func (m *MockStore) BalanceWithdraw(ctx context.Context, req balance.WithdrawRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BalanceWithdraw", ctx, req)
	ret0, _ := ret[0].(error)
	return ret0
}

// BalanceWithdraw indicates an expected call of BalanceWithdraw.
func (mr *MockStoreMockRecorder) BalanceWithdraw(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BalanceWithdraw", reflect.TypeOf((*MockStore)(nil).BalanceWithdraw), ctx, req)
}

// ClosePool mocks base method.
func (m *MockStore) ClosePool() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ClosePool")
	ret0, _ := ret[0].(error)
	return ret0
}

// ClosePool indicates an expected call of ClosePool.
func (mr *MockStoreMockRecorder) ClosePool() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClosePool", reflect.TypeOf((*MockStore)(nil).ClosePool))
}

// GetBalance mocks base method.
func (m *MockStore) GetBalance(ctx context.Context) (*storage.BalanceEntity, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBalance", ctx)
	ret0, _ := ret[0].(*storage.BalanceEntity)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalance indicates an expected call of GetBalance.
func (mr *MockStoreMockRecorder) GetBalance(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalance", reflect.TypeOf((*MockStore)(nil).GetBalance), ctx)
}

// GetOrdersList mocks base method.
func (m *MockStore) GetOrdersList(ctx context.Context) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrdersList", ctx)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrdersList indicates an expected call of GetOrdersList.
func (mr *MockStoreMockRecorder) GetOrdersList(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrdersList", reflect.TypeOf((*MockStore)(nil).GetOrdersList), ctx)
}

// GetUser mocks base method.
func (m *MockStore) GetUser(ctx context.Context, login string) (*storage.UserRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", ctx, login)
	ret0, _ := ret[0].(*storage.UserRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockStoreMockRecorder) GetUser(ctx, login any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockStore)(nil).GetUser), ctx, login)
}

// GetUserOrders mocks base method.
func (m *MockStore) GetUserOrders(ctx context.Context) ([]storage.OrderEntity, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserOrders", ctx)
	ret0, _ := ret[0].([]storage.OrderEntity)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserOrders indicates an expected call of GetUserOrders.
func (mr *MockStoreMockRecorder) GetUserOrders(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserOrders", reflect.TypeOf((*MockStore)(nil).GetUserOrders), ctx)
}

// GetWithdrawals mocks base method.
func (m *MockStore) GetWithdrawals(ctx context.Context) ([]storage.WithdrawalsEntity, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWithdrawals", ctx)
	ret0, _ := ret[0].([]storage.WithdrawalsEntity)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWithdrawals indicates an expected call of GetWithdrawals.
func (mr *MockStoreMockRecorder) GetWithdrawals(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWithdrawals", reflect.TypeOf((*MockStore)(nil).GetWithdrawals), ctx)
}

// SaveOrder mocks base method.
func (m *MockStore) SaveOrder(ctx context.Context, orderNo string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveOrder", ctx, orderNo)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveOrder indicates an expected call of SaveOrder.
func (mr *MockStoreMockRecorder) SaveOrder(ctx, orderNo any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveOrder", reflect.TypeOf((*MockStore)(nil).SaveOrder), ctx, orderNo)
}

// SaveUser mocks base method.
func (m *MockStore) SaveUser(ctx context.Context, user *register.Request) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveUser", ctx, user)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveUser indicates an expected call of SaveUser.
func (mr *MockStoreMockRecorder) SaveUser(ctx, user any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveUser", reflect.TypeOf((*MockStore)(nil).SaveUser), ctx, user)
}

// UpdateOrder mocks base method.
func (m *MockStore) UpdateOrder(ctx context.Context, data *orders.UpdateOrder) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOrder", ctx, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOrder indicates an expected call of UpdateOrder.
func (mr *MockStoreMockRecorder) UpdateOrder(ctx, data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOrder", reflect.TypeOf((*MockStore)(nil).UpdateOrder), ctx, data)
}
