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
		inputRequest       string
		prepareUserService func(mock *mocks.UserService)
		expectedHTTPStatus int
		expectedResponse   string
	}{
		{
			name:         "successful registration",
			inputRequest: `{"username":"test_user", "password":"password123"}`,
			prepareUserService: func(mock *mocks.UserService) {
				mock.On("Register", context.Background(), entity.UserLogin{
					Username: "test_user",
					Password: "password123",
				}).Return(entity.User{Username: "testuser"}, nil)
			},
			expectedHTTPStatus: http.StatusCreated,
			expectedResponse:   `{"code":201,"error":false,"message":"User successfully created","data":{"username":"testuser"}}`,
		},
		{
			name:         "username exists",
			inputRequest: `{"username":"test_user", "password":"password123"}`,
			prepareUserService: func(mock *mocks.UserService) {
				mock.On("Register", context.Background(), entity.UserLogin{
					Username: "test_user",
					Password: "password123",
				}).Return(entity.User{}, entity.ErrUsernameExists)
			},
			expectedHTTPStatus: http.StatusConflict,
			expectedResponse:   `{"code":409,"error":true,"message":"Username already exists"}`,
		},
		{
			name:               "invalid json",
			inputRequest:       `{"username":"test_user", "password":`,
			expectedHTTPStatus: http.StatusBadRequest,
		},
		{
			name:               "validation error: unexpected field 'usernfame'",
			inputRequest:       `{"usernfame":"test_user", "password":"password123"}`,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Unknown field: usernfame"}`,
		},
		{
			name:               "validation error: unexpected field 'passwordf'",
			inputRequest:       `{"username":"test_user", "passwordf":"password123"}`,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Unknown field: passwordf"}`,
		},
		{
			name:               "validation error: username to short",
			inputRequest:       `{"username":"ab", "password":"password123"}`,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Validation error: Field 'username' must be at least 3 character"}`,
		},
		{
			name:               "validation error: password to short",
			inputRequest:       `{"username":"test_user", "password":"123"}`,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Validation error: Field 'password' must be at least 8 character"}`,
		},
		{
			name:               "validation error: password to only digits",
			inputRequest:       `{"username":"test_user", "password":"12345678"}`,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Validation error: Field 'password' must contain at least one letter and one digit"}`,
		},
		{
			name:               "validation error: password to only letters",
			inputRequest:       `{"username":"test_user", "password":"abcderfcd"}`,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Validation error: Field 'password' must contain at least one letter and one digit"}`,
		},
		{
			name:         "internal server error",
			inputRequest: `{"username":"test_user", "password":"password123"}`,
			prepareUserService: func(mock *mocks.UserService) {
				mock.On("Register", context.Background(), entity.UserLogin{
					Username: "test_user",
					Password: "password123",
				}).Return(entity.User{}, errors.New("internal server error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   `{"code":500,"error":true,"message":"Something went wrong, please try again later"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := mocks.NewUserService(t)
			tokenServiceMock := mocks.NewTokenService(t)

			if tc.prepareUserService != nil {
				tc.prepareUserService(userServiceMock)
			}

			handler := NewHandler(userServiceMock, tokenServiceMock, logger)
			req, err := http.NewRequest("POST", "auth/register", bytes.NewBufferString(tc.inputRequest))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()

			handler.Register(rr, req)
			assert.Equal(t, tc.expectedHTTPStatus, rr.Code)
			if tc.expectedResponse != "" {

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
		name                string
		inputRequest        string
		prepareUserService  func(mock *mocks.UserService)
		prepareTokenService func(mock *mocks.TokenService)
		expectedHTTPStatus  int
		expectedResponse    string
	}{
		{

			name:         "successful login",
			inputRequest: `{"username":"test_user", "password":"password123"}`,
			prepareUserService: func(mock *mocks.UserService) {
				mock.On("GetByUsername", context.Background(), "test_user").
					Return(entity.User{
						ID:           userID,
						Username:     "test_user",
						PasswordHash: passwordHash,
					}, nil)
			},
			prepareTokenService: func(mock *mocks.TokenService) {
				mock.On("GenerateTokenPair", context.Background(), userID).
					Return("access_token", "refresh_token", nil)
			},
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   `{"code":200,"error":false,"message":"Login successful","data":{"access_token":"access_token", "refresh_token":"refresh_token"}}`,
		},
		{

			name:               "invalid json",
			inputRequest:       `{"username":"test_user", "password":"password123",}`,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Invalid json format"}`,
		},
		{
			name:               "validation error: unexpected field 'usernfame'",
			inputRequest:       `{"usernfame":"test_user", "password":"password123"}`,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Unknown field: usernfame"}`,
		},
		{
			name:               "validation error: unexpected field 'passwordf'",
			inputRequest:       `{"username":"test_user", "passwordf":"password123"}`,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Unknown field: passwordf"}`,
		},
		{
			name:         "invalid credentials",
			inputRequest: `{"username":"test_user", "password":"password123"}`,
			prepareUserService: func(mock *mocks.UserService) {
				mock.On("GetByUsername", context.Background(), "test_user").
					Return(entity.User{}, entity.ErrUserNotFound)
			},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedResponse:   `{"code":401,"error":true,"message":"Invalid credentials"}`,
		},
		{
			name:         "internal user server error",
			inputRequest: `{"username":"test_user", "password":"password123"}`,
			prepareUserService: func(mock *mocks.UserService) {
				mock.On("GetByUsername", context.Background(), "test_user").
					Return(entity.User{}, errors.New("internal server error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   `{"code":500,"error":true,"message":"Something went wrong, please try again later"}`,
		},
		{
			name:         "internal token server error",
			inputRequest: `{"username":"test_user", "password":"password123"}`,
			prepareUserService: func(mock *mocks.UserService) {
				mock.On("GetByUsername", context.Background(), "test_user").
					Return(entity.User{
						ID:           userID,
						Username:     "password123",
						PasswordHash: passwordHash,
					}, nil)
			},
			prepareTokenService: func(mock *mocks.TokenService) {
				mock.On("GenerateTokenPair", context.Background(), userID).
					Return("", "", errors.New("internal server error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   `{"code":500,"error":true,"message":"Something went wrong, please try again later"}`,
		},
		{
			name:         "invalid password",
			inputRequest: `{"username":"test_user", "password":"password123"}`,
			prepareUserService: func(mock *mocks.UserService) {
				mock.On("GetByUsername", context.Background(), "test_user").
					Return(entity.User{
						ID:           userID,
						Username:     "test_user",
						PasswordHash: "invalid_hash",
					}, nil)
			},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedResponse:   `{"code":401,"error":true,"message":"Invalid credentials"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := mocks.NewUserService(t)
			tokenServiceMock := mocks.NewTokenService(t)

			if tc.prepareUserService != nil {
				tc.prepareUserService(userServiceMock)
			}
			if tc.prepareTokenService != nil {
				tc.prepareTokenService(tokenServiceMock)
			}

			handler := NewHandler(userServiceMock, tokenServiceMock, logger)
			req, err := http.NewRequest("POST", "auth/login", bytes.NewBufferString(tc.inputRequest))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()

			handler.Login(rr, req)
			assert.Equal(t, tc.expectedHTTPStatus, rr.Code)
			if tc.expectedResponse != "" {
				assert.JSONEq(t, tc.expectedResponse, rr.Body.String())
			}

		})
	}
}

func TestRefresh(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	testCases := []struct {
		name                string
		inputRequest        string
		prepareTokenService func(mock *mocks.TokenService)
		expectedHTTPStatus  int
		expectedResponse    string
	}{
		{
			name:         "successful refresh",
			inputRequest: `{"refresh_token": "refresh_token"}`,
			prepareTokenService: func(mock *mocks.TokenService) {
				mock.On("RefreshTokens", context.Background(), "refresh_token").
					Return("access_token", "new_refresh_token", nil)
			},
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   `{"code":200, "data":{"access_token":"access_token", "refresh_token":"new_refresh_token"}, "error":false, "message":"Tokens refreshed"}`,
		},
		{

			name:               "invalid json",
			inputRequest:       `{"refresh_token": "refresh_token"`,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Invalid json format"}`,
		},
		{

			name:         "invalid refresh token",
			inputRequest: `{"refresh_token": "refresh_token"}`,
			prepareTokenService: func(mock *mocks.TokenService) {
				mock.On("RefreshTokens", context.Background(), "refresh_token").
					Return("", "", errors.New("invalid refresh token"))
			},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedResponse:   `{"code":401,"error":true,"message":"Invalid refresh token"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := mocks.NewUserService(t)
			tokenServiceMock := mocks.NewTokenService(t)

			if tc.prepareTokenService != nil {
				tc.prepareTokenService(tokenServiceMock)
			}

			handler := NewHandler(userServiceMock, tokenServiceMock, logger)
			req, err := http.NewRequest("POST", "auth/refresh", bytes.NewBufferString(tc.inputRequest))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			handler.Refresh(rr, req)

			assert.Equal(t, tc.expectedHTTPStatus, rr.Code)
			if tc.expectedResponse != "" {
				assert.JSONEq(t, tc.expectedResponse, rr.Body.String())
			}
		})
	}
}
