package auth

import (
	"bytes"
	"context"
	"errors"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/service/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRegister(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	testCases := []struct {
		name               string
		input              string
		userServiceFn      func(mock *mocks.UserService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:  "successful registration",
			input: `{"username":"test_user", "password":"password123"}`,
			userServiceFn: func(mock *mocks.UserService) {
				mock.On("Register", context.Background(), entity.UserLogin{
					Username: "test_user",
					Password: "password123",
				}).Return(entity.User{Username: "testuser"}, nil)
			},
			expectedStatusCode: http.StatusCreated,
			expectedBody:       `{"code":201,"error":false,"message":"User successfully created","data":{"username":"testuser"}}`,
		},
		{
			name:  "username exists",
			input: `{"username":"test_user", "password":"password123"}`,
			userServiceFn: func(mock *mocks.UserService) {
				mock.On("Register", context.Background(), entity.UserLogin{
					Username: "test_user",
					Password: "password123",
				}).Return(entity.User{}, entity.ErrUsernameExists)
			},
			expectedStatusCode: http.StatusConflict,
			expectedBody:       `{"code":409,"error":true,"message":"Username already exists"}`,
		},
		{
			name:               "invalid json",
			input:              `{"username":"test_user", "password":`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "validation error: unexpected field 'usernfame'",
			input:              `{"usernfame":"test_user", "password":"password123"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"code":400,"error":true,"message":"Unknown field: usernfame"}`,
		},
		{
			name:               "validation error: unexpected field 'passwordf'",
			input:              `{"username":"test_user", "passwordf":"password123"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"code":400,"error":true,"message":"Unknown field: passwordf"}`,
		},
		{
			name:               "validation error: username to short",
			input:              `{"username":"ab", "password":"password123"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"code":400,"error":true,"message":"Validation error: Field 'username' must be at least 3 character"}`,
		},
		{
			name:               "validation error: password to short",
			input:              `{"username":"test_user", "password":"123"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"code":400,"error":true,"message":"Validation error: Field 'password' must be at least 8 character"}`,
		},
		{
			name:               "validation error: password to only digits",
			input:              `{"username":"test_user", "password":"12345678"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"code":400,"error":true,"message":"Validation error: Field 'password' must contain at least one letter and one digit"}`,
		},
		{
			name:               "validation error: password to only letters",
			input:              `{"username":"test_user", "password":"abcderfcd"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"code":400,"error":true,"message":"Validation error: Field 'password' must contain at least one letter and one digit"}`,
		},
		{
			name:  "internal server error",
			input: `{"username":"test_user", "password":"password123"}`,
			userServiceFn: func(mock *mocks.UserService) {
				mock.On("Register", context.Background(), entity.UserLogin{
					Username: "test_user",
					Password: "password123",
				}).Return(entity.User{}, errors.New("internal server error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"code":500,"error":true,"message":"Something went wrong, please try again later"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := mocks.NewUserService(t)
			tokenServiceMock := mocks.NewTokenService(t)

			if tc.userServiceFn != nil {
				tc.userServiceFn(userServiceMock)
			}

			handler := NewHandler(userServiceMock, tokenServiceMock, logger)
			req, _ := http.NewRequest("POST", "auth/register", bytes.NewBufferString(tc.input))
			rr := httptest.NewRecorder()

			handler.Register(rr, req)
			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			if tc.expectedBody != "" {

			}
		})
	}
}

func TestLogin(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	userID := uuid.New()
	passwordHash, err := entity.HashPassword("password123")
	assert.NoError(t, err)

	testCases := []struct {
		name               string
		input              string
		userServiceFn      func(mock *mocks.UserService)
		tokenServiceFn     func(mock *mocks.TokenService)
		expectedStatusCode int
		expectedBody       string
	}{
		{

			name:  "successful login",
			input: `{"username":"test_user", "password":"password123"}`,
			userServiceFn: func(mock *mocks.UserService) {
				mock.On("GetByUsername", context.Background(), "test_user").
					Return(entity.User{
						ID:           userID,
						Username:     "test_user",
						PasswordHash: passwordHash,
					}, nil)
			},
			tokenServiceFn: func(mock *mocks.TokenService) {
				mock.On("GenerateTokenPair", userID).
					Return("access_token", "refresh_token", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"code":200,"error":false,"message":"Login successful","data":{"access_token":"access_token", "refresh_token":"refresh_token"}}`,
		},
		{

			name:               "invalid json",
			input:              `{"username":"test_user", "password":"password123",}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"code":400,"error":true,"message":"Invalid json format"}`,
		},
		{
			name:               "validation error: unexpected field 'usernfame'",
			input:              `{"usernfame":"test_user", "password":"password123"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"code":400,"error":true,"message":"Unknown field: usernfame"}`,
		},
		{
			name:               "validation error: unexpected field 'passwordf'",
			input:              `{"username":"test_user", "passwordf":"password123"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"code":400,"error":true,"message":"Unknown field: passwordf"}`,
		},
		{
			name:  "invalid credentials",
			input: `{"username":"test_user", "password":"password123"}`,
			userServiceFn: func(mock *mocks.UserService) {
				mock.On("GetByUsername", context.Background(), "test_user").
					Return(entity.User{}, entity.ErrUserNotFound)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       `{"code":401,"error":true,"message":"Invalid credentials"}`,
		},
		{
			name:  "internal user server error",
			input: `{"username":"test_user", "password":"password123"}`,
			userServiceFn: func(mock *mocks.UserService) {
				mock.On("GetByUsername", context.Background(), "test_user").
					Return(entity.User{}, errors.New("internal server error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"code":500,"error":true,"message":"Something went wrong, please try again later"}`,
		},
		{
			name:  "internal token server error",
			input: `{"username":"test_user", "password":"password123"}`,
			userServiceFn: func(mock *mocks.UserService) {
				mock.On("GetByUsername", context.Background(), "test_user").
					Return(entity.User{
						ID:           userID,
						Username:     "password123",
						PasswordHash: passwordHash,
					}, nil)
			},
			tokenServiceFn: func(mock *mocks.TokenService) {
				mock.On("GenerateTokenPair", userID).
					Return("", "", errors.New("internal server error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"code":500,"error":true,"message":"Something went wrong, please try again later"}`,
		},
		{
			name:  "invalid password",
			input: `{"username":"test_user", "password":"password123"}`,
			userServiceFn: func(mock *mocks.UserService) {
				mock.On("GetByUsername", context.Background(), "test_user").
					Return(entity.User{
						ID:           userID,
						Username:     "test_user",
						PasswordHash: "invalid_hash",
					}, nil)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       `{"code":401,"error":true,"message":"Invalid credentials"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := mocks.NewUserService(t)
			tokenServiceMock := mocks.NewTokenService(t)

			if tc.userServiceFn != nil {
				tc.userServiceFn(userServiceMock)
			}
			if tc.tokenServiceFn != nil {
				tc.tokenServiceFn(tokenServiceMock)
			}

			handler := NewHandler(userServiceMock, tokenServiceMock, logger)
			req, _ := http.NewRequest("POST", "auth/login", bytes.NewBufferString(tc.input))
			rr := httptest.NewRecorder()

			handler.Login(rr, req)
			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}

		})
	}
}

func TestRefresh(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	testCases := []struct {
		name               string
		input              string
		tokenServiceFn     func(mock *mocks.TokenService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:  "successful refresh",
			input: `{"refresh_token": "refresh_token"}`,
			tokenServiceFn: func(mock *mocks.TokenService) {
				mock.On("RefreshTokens", "refresh_token").
					Return("access_token", "new_refresh_token", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"code":200, "data":{"access_token":"access_token", "refresh_token":"new_refresh_token"}, "error":false, "message":"Tokens refreshed"}`,
		},
		{

			name:               "invalid json",
			input:              `{"refresh_token": "refresh_token"`,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"code":400,"error":true,"message":"Invalid json format"}`,
		},
		{

			name:  "invalid refresh token",
			input: `{"refresh_token": "refresh_token"}`,
			tokenServiceFn: func(mock *mocks.TokenService) {
				mock.On("RefreshTokens", "refresh_token").
					Return("", "", errors.New("invalid refresh token"))
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       `{"code":401,"error":true,"message":"Invalid refresh token"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := mocks.NewUserService(t)
			tokenServiceMock := mocks.NewTokenService(t)

			if tc.tokenServiceFn != nil {
				tc.tokenServiceFn(tokenServiceMock)
			}

			handler := NewHandler(userServiceMock, tokenServiceMock, logger)
			req, _ := http.NewRequest("POST", "auth/refresh", bytes.NewBufferString(tc.input))
			rr := httptest.NewRecorder()
			handler.Refresh(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}
