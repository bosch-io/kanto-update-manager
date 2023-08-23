// Copyright (c) 2023 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0

// Code generated by MockGen. DO NOT EDIT.
// Source: client_api.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	ditto "github.com/eclipse/ditto-clients-golang"
	protocol "github.com/eclipse/ditto-clients-golang/protocol"
	gomock "github.com/golang/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// Connect mocks base method.
func (m *MockClient) Connect() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect")
	ret0, _ := ret[0].(error)
	return ret0
}

// Connect indicates an expected call of Connect.
func (mr *MockClientMockRecorder) Connect() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockClient)(nil).Connect))
}

// Disconnect mocks base method.
func (m *MockClient) Disconnect() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Disconnect")
}

// Disconnect indicates an expected call of Disconnect.
func (mr *MockClientMockRecorder) Disconnect() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Disconnect", reflect.TypeOf((*MockClient)(nil).Disconnect))
}

// Reply mocks base method.
func (m *MockClient) Reply(requestID string, message *protocol.Envelope) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reply", requestID, message)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reply indicates an expected call of Reply.
func (mr *MockClientMockRecorder) Reply(requestID, message interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reply", reflect.TypeOf((*MockClient)(nil).Reply), requestID, message)
}

// Send mocks base method.
func (m *MockClient) Send(message *protocol.Envelope) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", message)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockClientMockRecorder) Send(message interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockClient)(nil).Send), message)
}

// Subscribe mocks base method.
func (m *MockClient) Subscribe(handlers ...ditto.Handler) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range handlers {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Subscribe", varargs...)
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockClientMockRecorder) Subscribe(handlers ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockClient)(nil).Subscribe), handlers...)
}

// Unsubscribe mocks base method.
func (m *MockClient) Unsubscribe(handlers ...ditto.Handler) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range handlers {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Unsubscribe", varargs...)
}

// Unsubscribe indicates an expected call of Unsubscribe.
func (mr *MockClientMockRecorder) Unsubscribe(handlers ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unsubscribe", reflect.TypeOf((*MockClient)(nil).Unsubscribe), handlers...)
}