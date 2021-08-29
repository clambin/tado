// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	context "context"
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// Authenticator is an autogenerated mock type for the Authenticator type
type Authenticator struct {
	mock.Mock
}

// AuthHeaders provides a mock function with given fields: ctx
func (_m *Authenticator) AuthHeaders(ctx context.Context) (http.Header, error) {
	ret := _m.Called(ctx)

	var r0 http.Header
	if rf, ok := ret.Get(0).(func(context.Context) http.Header); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Header)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Reset provides a mock function with given fields:
func (_m *Authenticator) Reset() {
	_m.Called()
}
