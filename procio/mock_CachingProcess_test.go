// Code generated by mockery v2.12.2. DO NOT EDIT.

package procio

import (
	testing "testing"

	mock "github.com/stretchr/testify/mock"
)

// MockCachingProcess is an autogenerated mock type for the CachingProcess type
type MockCachingProcess struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *MockCachingProcess) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Crash provides a mock function with given fields: _a0
func (_m *MockCachingProcess) Crash(_a0 CrashMethod) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(CrashMethod) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Handle provides a mock function with given fields:
func (_m *MockCachingProcess) Handle() interface{} {
	ret := _m.Called()

	var r0 interface{}
	if rf, ok := ret.Get(0).(func() interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// Info provides a mock function with given fields:
func (_m *MockCachingProcess) Info() (*ProcessInfo, error) {
	ret := _m.Called()

	var r0 *ProcessInfo
	if rf, ok := ret.Get(0).(func() *ProcessInfo); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ProcessInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InvalidateCache provides a mock function with given fields:
func (_m *MockCachingProcess) InvalidateCache() {
	_m.Called()
}

// MemorySegments provides a mock function with given fields:
func (_m *MockCachingProcess) MemorySegments() ([]*MemorySegmentInfo, error) {
	ret := _m.Called()

	var r0 []*MemorySegmentInfo
	if rf, ok := ret.Get(0).(func() []*MemorySegmentInfo); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*MemorySegmentInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PID provides a mock function with given fields:
func (_m *MockCachingProcess) PID() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Resume provides a mock function with given fields:
func (_m *MockCachingProcess) Resume() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// String provides a mock function with given fields:
func (_m *MockCachingProcess) String() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Suspend provides a mock function with given fields:
func (_m *MockCachingProcess) Suspend() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockCachingProcess creates a new instance of MockCachingProcess. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockCachingProcess(t testing.TB) *MockCachingProcess {
	mock := &MockCachingProcess{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
