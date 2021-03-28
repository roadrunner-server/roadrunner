package mocks

import (
	"reflect"
	"sync"

	"github.com/golang/mock/gomock"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// MockLogger is a mock of Logger interface.
type MockLogger struct {
	sync.Mutex
	ctrl     *gomock.Controller
	recorder *MockLoggerMockRecorder
}

// MockLoggerMockRecorder is the mock recorder for MockLogger.
type MockLoggerMockRecorder struct {
	mock *MockLogger
}

// NewMockLogger creates a new mock instance.
func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &MockLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogger) EXPECT() *MockLoggerMockRecorder {
	return m.recorder
}

func (m *MockLogger) Init() error {
	mock := &MockLogger{ctrl: m.ctrl}
	mock.recorder = &MockLoggerMockRecorder{mock}
	return nil
}

// Debug mocks base method.
func (m *MockLogger) Debug(msg string, keyvals ...interface{}) {
	m.Lock()
	defer m.Unlock()
	m.ctrl.T.Helper()
	varargs := []interface{}{msg}
	varargs = append(varargs, keyvals...)
	m.ctrl.Call(m, "Debug", varargs...)
}

// Warn mocks base method.
func (m *MockLogger) Warn(msg string, keyvals ...interface{}) {
	m.Lock()
	defer m.Unlock()
	m.ctrl.T.Helper()
	varargs := []interface{}{msg}
	varargs = append(varargs, keyvals...)
	m.ctrl.Call(m, "Warn", varargs...)
}

// Info mocks base method.
func (m *MockLogger) Info(msg string, keyvals ...interface{}) {
	m.Lock()
	defer m.Unlock()
	m.ctrl.T.Helper()
	varargs := []interface{}{msg}
	varargs = append(varargs, keyvals...)
	m.ctrl.Call(m, "Info", varargs...)
}

// Error mocks base method.
func (m *MockLogger) Error(msg string, keyvals ...interface{}) {
	m.Lock()
	defer m.Unlock()
	m.ctrl.T.Helper()
	varargs := []interface{}{msg}
	varargs = append(varargs, keyvals...)
	m.ctrl.Call(m, "Error", varargs...)
}

// Warn indicates an expected call of Warn.
func (mr *MockLoggerMockRecorder) Warn(msg interface{}, keyvals ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{msg}, keyvals...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockLogger)(nil).Warn), varargs...)
}

// Debug indicates an expected call of Debug.
func (mr *MockLoggerMockRecorder) Debug(msg interface{}, keyvals ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{msg}, keyvals...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLogger)(nil).Debug), varargs...)
}

// Error indicates an expected call of Error.
func (mr *MockLoggerMockRecorder) Error(msg interface{}, keyvals ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{msg}, keyvals...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLogger)(nil).Error), varargs...)
}

func (mr *MockLoggerMockRecorder) Init() error {
	return nil
}

// Info indicates an expected call of Info.
func (mr *MockLoggerMockRecorder) Info(msg interface{}, keyvals ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{msg}, keyvals...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLogger)(nil).Info), varargs...)
}

// MockWithLogger is a mock of WithLogger interface.
type MockWithLogger struct {
	ctrl     *gomock.Controller
	recorder *MockWithLoggerMockRecorder
}

// MockWithLoggerMockRecorder is the mock recorder for MockWithLogger.
type MockWithLoggerMockRecorder struct {
	mock *MockWithLogger
}

// NewMockWithLogger creates a new mock instance.
func NewMockWithLogger(ctrl *gomock.Controller) *MockWithLogger {
	mock := &MockWithLogger{ctrl: ctrl}
	mock.recorder = &MockWithLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWithLogger) EXPECT() *MockWithLoggerMockRecorder {
	return m.recorder
}

// With mocks base method.
func (m *MockWithLogger) With(keyvals ...interface{}) logger.Logger {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	varargs = append(varargs, keyvals...)
	ret := m.ctrl.Call(m, "With", varargs...)
	ret0, _ := ret[0].(logger.Logger)
	return ret0
}

// With indicates an expected call of With.
func (mr *MockWithLoggerMockRecorder) With(keyvals ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "With", reflect.TypeOf((*MockWithLogger)(nil).With), keyvals...)
}
