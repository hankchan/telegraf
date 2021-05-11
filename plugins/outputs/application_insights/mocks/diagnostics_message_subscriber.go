// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import appinsights "github.com/microsoft/ApplicationInsights-Go/appinsights"

import mock "github.com/stretchr/testify/mock"

// DiagnosticsMessageSubscriber is an autogenerated mock type for the DiagnosticsMessageSubscriber type
type DiagnosticsMessageSubscriber struct {
	mock.Mock
}

// Subscribe provides a mock function with given fields: _a0
func (_m *DiagnosticsMessageSubscriber) Subscribe(_a0 appinsights.DiagnosticsMessageHandler) appinsights.DiagnosticsMessageListener {
	ret := _m.Called(_a0)

	var r0 appinsights.DiagnosticsMessageListener
	if rf, ok := ret.Get(0).(func(appinsights.DiagnosticsMessageHandler) appinsights.DiagnosticsMessageListener); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(appinsights.DiagnosticsMessageListener)
		}
	}

	return r0
}
