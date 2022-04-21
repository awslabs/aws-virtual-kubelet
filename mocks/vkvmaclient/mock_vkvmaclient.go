// Package mock_vkvmaclient was generated initially by GoMock
// It may be edited to add functionality as-needed

package mock_vkvmaclient

import (
	context "context"
	reflect "reflect"

	grpc_health_v1 "github.com/aws/aws-virtual-kubelet/proto/grpc/health/v1"
	vkvmagent_v0 "github.com/aws/aws-virtual-kubelet/proto/vkvmagent/v0"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockGrpcClient is a mock of GrpcClient interface.
type MockGrpcClient struct {
	ctrl     *gomock.Controller
	recorder *MockGrpcClientMockRecorder
}

// MockGrpcClientMockRecorder is the mock recorder for MockGrpcClient.
type MockGrpcClientMockRecorder struct {
	mock *MockGrpcClient
}

// NewMockGrpcClient creates a new mock instance.
func NewMockGrpcClient(ctrl *gomock.Controller) *MockGrpcClient {
	mock := &MockGrpcClient{ctrl: ctrl}
	mock.recorder = &MockGrpcClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGrpcClient) EXPECT() *MockGrpcClientMockRecorder {
	return m.recorder
}

// Connect mocks base method.
func (m *MockGrpcClient) Connect(ctx context.Context) (*grpc.ClientConn, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect", ctx)
	ret0, _ := ret[0].(*grpc.ClientConn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Connect indicates an expected call of Connect.
func (mr *MockGrpcClientMockRecorder) Connect(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockGrpcClient)(nil).Connect), ctx)
}

// GetApplicationLifecycleClient mocks base method.
func (m *MockGrpcClient) GetApplicationLifecycleClient(ctx context.Context) (vkvmagent_v0.ApplicationLifecycleClient, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationLifecycleClient", ctx)
	ret0, _ := ret[0].(vkvmagent_v0.ApplicationLifecycleClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationLifecycleClient indicates an expected call of GetApplicationLifecycleClient.
func (mr *MockGrpcClientMockRecorder) GetApplicationLifecycleClient(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationLifecycleClient", reflect.TypeOf((*MockGrpcClient)(nil).GetApplicationLifecycleClient), ctx)
}

// GetHealthClient mocks base method.
func (m *MockGrpcClient) GetHealthClient(ctx context.Context) (grpc_health_v1.HealthClient, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHealthClient", ctx)
	ret0, _ := ret[0].(grpc_health_v1.HealthClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHealthClient indicates an expected call of GetHealthClient.
func (mr *MockGrpcClientMockRecorder) GetHealthClient(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHealthClient", reflect.TypeOf((*MockGrpcClient)(nil).GetHealthClient), ctx)
}

// IsConnected mocks base method.
func (m *MockGrpcClient) IsConnected(ctx context.Context) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsConnected", ctx)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsConnected indicates an expected call of IsConnected.
func (mr *MockGrpcClientMockRecorder) IsConnected(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsConnected", reflect.TypeOf((*MockGrpcClient)(nil).IsConnected), ctx)
}
