// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// TokenService is an autogenerated mock type for the TokenService type
type TokenService struct {
	mock.Mock
}

// GenerateTokenPair provides a mock function with given fields: ctx, id
func (_m *TokenService) GenerateTokenPair(ctx context.Context, id uuid.UUID) (string, string, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GenerateTokenPair")
	}

	var r0 string
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (string, string, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) string); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) string); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(context.Context, uuid.UUID) error); ok {
		r2 = rf(ctx, id)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// RefreshTokens provides a mock function with given fields: ctx, refreshTokenString
func (_m *TokenService) RefreshTokens(ctx context.Context, refreshTokenString string) (string, string, error) {
	ret := _m.Called(ctx, refreshTokenString)

	if len(ret) == 0 {
		panic("no return value specified for RefreshTokens")
	}

	var r0 string
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, string, error)); ok {
		return rf(ctx, refreshTokenString)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, refreshTokenString)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) string); ok {
		r1 = rf(ctx, refreshTokenString)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(context.Context, string) error); ok {
		r2 = rf(ctx, refreshTokenString)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ValidateAccessToken provides a mock function with given fields: tokenString
func (_m *TokenService) ValidateAccessToken(tokenString string) (uuid.UUID, error) {
	ret := _m.Called(tokenString)

	if len(ret) == 0 {
		panic("no return value specified for ValidateAccessToken")
	}

	var r0 uuid.UUID
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (uuid.UUID, error)); ok {
		return rf(tokenString)
	}
	if rf, ok := ret.Get(0).(func(string) uuid.UUID); ok {
		r0 = rf(tokenString)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uuid.UUID)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tokenString)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateRefreshToken provides a mock function with given fields: tokenString
func (_m *TokenService) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	ret := _m.Called(tokenString)

	if len(ret) == 0 {
		panic("no return value specified for ValidateRefreshToken")
	}

	var r0 uuid.UUID
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (uuid.UUID, error)); ok {
		return rf(tokenString)
	}
	if rf, ok := ret.Get(0).(func(string) uuid.UUID); ok {
		r0 = rf(tokenString)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uuid.UUID)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tokenString)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewTokenService creates a new instance of TokenService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTokenService(t interface {
	mock.TestingT
	Cleanup(func())
}) *TokenService {
	mock := &TokenService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
