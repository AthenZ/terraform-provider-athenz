// Code generated by MockGen. DO NOT EDIT.
// Source: client.go

// Package client is a generated GoMock package.
package client

import (
	reflect "reflect"

	zms "github.com/AthenZ/athenz/clients/go/zms"
	gomock "github.com/golang/mock/gomock"
)

// MockZmsClient is a mock of ZmsClient interface.
type MockZmsClient struct {
	ctrl     *gomock.Controller
	recorder *MockZmsClientMockRecorder
}

// MockZmsClientMockRecorder is the mock recorder for MockZmsClient.
type MockZmsClientMockRecorder struct {
	mock *MockZmsClient
}

// NewMockZmsClient creates a new mock instance.
func NewMockZmsClient(ctrl *gomock.Controller) *MockZmsClient {
	mock := &MockZmsClient{ctrl: ctrl}
	mock.recorder = &MockZmsClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockZmsClient) EXPECT() *MockZmsClientMockRecorder {
	return m.recorder
}

// DeleteGroup mocks base method.
func (m *MockZmsClient) DeleteGroup(domain, groupName, auditRef string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteGroup", domain, groupName, auditRef)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteGroup indicates an expected call of DeleteGroup.
func (mr *MockZmsClientMockRecorder) DeleteGroup(domain, groupName, auditRef interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteGroup", reflect.TypeOf((*MockZmsClient)(nil).DeleteGroup), domain, groupName, auditRef)
}

// DeleteGroupMembership mocks base method.
func (m *MockZmsClient) DeleteGroupMembership(domain, groupName string, member zms.GroupMemberName, auditRef string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteGroupMembership", domain, groupName, member, auditRef)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteGroupMembership indicates an expected call of DeleteGroupMembership.
func (mr *MockZmsClientMockRecorder) DeleteGroupMembership(domain, groupName, member, auditRef interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteGroupMembership", reflect.TypeOf((*MockZmsClient)(nil).DeleteGroupMembership), domain, groupName, member, auditRef)
}

// DeleteMembership mocks base method.
func (m *MockZmsClient) DeleteMembership(domain, roleMember string, member zms.MemberName, auditRef string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMembership", domain, roleMember, member, auditRef)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteMembership indicates an expected call of DeleteMembership.
func (mr *MockZmsClientMockRecorder) DeleteMembership(domain, roleMember, member, auditRef interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMembership", reflect.TypeOf((*MockZmsClient)(nil).DeleteMembership), domain, roleMember, member, auditRef)
}

// DeletePolicy mocks base method.
func (m *MockZmsClient) DeletePolicy(domain, policyName, auditRef string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePolicy", domain, policyName, auditRef)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePolicy indicates an expected call of DeletePolicy.
func (mr *MockZmsClientMockRecorder) DeletePolicy(domain, policyName, auditRef interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePolicy", reflect.TypeOf((*MockZmsClient)(nil).DeletePolicy), domain, policyName, auditRef)
}

// DeleteRole mocks base method.
func (m *MockZmsClient) DeleteRole(domain, roleName, auditRef string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteRole", domain, roleName, auditRef)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteRole indicates an expected call of DeleteRole.
func (mr *MockZmsClientMockRecorder) DeleteRole(domain, roleName, auditRef interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRole", reflect.TypeOf((*MockZmsClient)(nil).DeleteRole), domain, roleName, auditRef)
}

// GetGroup mocks base method.
func (m *MockZmsClient) GetGroup(domain, groupName string) (*zms.Group, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroup", domain, groupName)
	ret0, _ := ret[0].(*zms.Group)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGroup indicates an expected call of GetGroup.
func (mr *MockZmsClientMockRecorder) GetGroup(domain, groupName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroup", reflect.TypeOf((*MockZmsClient)(nil).GetGroup), domain, groupName)
}

// GetPolicy mocks base method.
func (m *MockZmsClient) GetPolicy(domain, policy string) (*zms.Policy, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPolicy", domain, policy)
	ret0, _ := ret[0].(*zms.Policy)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPolicy indicates an expected call of GetPolicy.
func (mr *MockZmsClientMockRecorder) GetPolicy(domain, policy interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPolicy", reflect.TypeOf((*MockZmsClient)(nil).GetPolicy), domain, policy)
}

// GetRole mocks base method.
func (m *MockZmsClient) GetRole(domain, roleName string) (*zms.Role, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRole", domain, roleName)
	ret0, _ := ret[0].(*zms.Role)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRole indicates an expected call of GetRole.
func (mr *MockZmsClientMockRecorder) GetRole(domain, roleName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRole", reflect.TypeOf((*MockZmsClient)(nil).GetRole), domain, roleName)
}

// PutGroup mocks base method.
func (m *MockZmsClient) PutGroup(domain, groupName, auditRef string, group *zms.Group) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutGroup", domain, groupName, auditRef, group)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutGroup indicates an expected call of PutGroup.
func (mr *MockZmsClientMockRecorder) PutGroup(domain, groupName, auditRef, group interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutGroup", reflect.TypeOf((*MockZmsClient)(nil).PutGroup), domain, groupName, auditRef, group)
}

// PutGroupMembership mocks base method.
func (m *MockZmsClient) PutGroupMembership(domain, groupName string, memberName zms.GroupMemberName, auditRef string, membership *zms.GroupMembership) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutGroupMembership", domain, groupName, memberName, auditRef, membership)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutGroupMembership indicates an expected call of PutGroupMembership.
func (mr *MockZmsClientMockRecorder) PutGroupMembership(domain, groupName, memberName, auditRef, membership interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutGroupMembership", reflect.TypeOf((*MockZmsClient)(nil).PutGroupMembership), domain, groupName, memberName, auditRef, membership)
}

// PutMembership mocks base method.
func (m *MockZmsClient) PutMembership(domain, roleName string, memberName zms.MemberName, auditRef string, membership *zms.Membership) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutMembership", domain, roleName, memberName, auditRef, membership)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutMembership indicates an expected call of PutMembership.
func (mr *MockZmsClientMockRecorder) PutMembership(domain, roleName, memberName, auditRef, membership interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutMembership", reflect.TypeOf((*MockZmsClient)(nil).PutMembership), domain, roleName, memberName, auditRef, membership)
}

// PutPolicy mocks base method.
func (m *MockZmsClient) PutPolicy(domain, policyName, auditRef string, policy *zms.Policy) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutPolicy", domain, policyName, auditRef, policy)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutPolicy indicates an expected call of PutPolicy.
func (mr *MockZmsClientMockRecorder) PutPolicy(domain, policyName, auditRef, policy interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutPolicy", reflect.TypeOf((*MockZmsClient)(nil).PutPolicy), domain, policyName, auditRef, policy)
}

// PutRole mocks base method.
func (m *MockZmsClient) PutRole(domain, roleName, auditRef string, role *zms.Role) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutRole", domain, roleName, auditRef, role)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutRole indicates an expected call of PutRole.
func (mr *MockZmsClientMockRecorder) PutRole(domain, roleName, auditRef, role interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutRole", reflect.TypeOf((*MockZmsClient)(nil).PutRole), domain, roleName, auditRef, role)
}
