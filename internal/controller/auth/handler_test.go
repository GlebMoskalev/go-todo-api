package auth

import (
	"bytes"
	"context"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/service/mocks"
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
		name           string
		input          string
		userServiceFn  func(mock *mocks.UserService)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:  "successful registration",
			input: `{"username":"testuser", "password":"password123"}`,
			userServiceFn: func(mock *mocks.UserService) {
				mock.On("Register", context.Background(), entity.UserLogin{
					Username: "testuser",
					Password: "password123",
				}).Return(entity.User{Username: "testuser"}, nil)
			},
			wantStatusCode: http.StatusCreated,
			wantBody:       `{"code":201,"error":false,"message":"User successfully created","data":{"username":"testuser"}}`,
		},
		{
			name:  "username exists",
			input: `{"username":"testuser", "password":"password123"}`,
			userServiceFn: func(mock *mocks.UserService) {
				mock.On("Register", context.Background(), entity.UserLogin{
					Username: "testuser",
					Password: "password123",
				}).Return(entity.User{}, entity.ErrUsernameExists)
			},
			wantStatusCode: http.StatusConflict,
			wantBody:       `{"code":409,"error":true,"message":"Username already exists"}`,
		},
		{
			name:           "invalid json",
			input:          `{"username":"testuser", "password":`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "validation error: unexpected field 'usernfame'",
			input:          `{"usernfame":"testuser", "password":"password123"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":400,"error":true,"message":"Unknown field: usernfame"}`,
		},
		{
			name:           "validation error: unexpected field 'passwordf'",
			input:          `{"username":"testuser", "passwordf":"password123"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":400,"error":true,"message":"Unknown field: passwordf"}`,
		},
		{
			name:           "validation error: username to short",
			input:          `{"username":"ab", "password":"password123"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":400,"error":true,"message":"Validation error: Field 'username' must be at least 3 character"}`,
		},
		{
			name:           "validation error: password to short",
			input:          `{"username":"testuser", "password":"123"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":400,"error":true,"message":"Validation error: Field 'password' must be at least 8 character"}`,
		},
		{
			name:           "validation error: password to only digits",
			input:          `{"username":"testuser", "password":"12345678"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":400,"error":true,"message":"Validation error: Field 'password' must contain at least one letter and one digit"}`,
		},
		{
			name:           "validation error: password to only letters",
			input:          `{"username":"testuser", "password":"abcderfcd"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"code":400,"error":true,"message":"Validation error: Field 'password' must contain at least one letter and one digit"}`,
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
			assert.Equal(t, tc.wantStatusCode, rr.Code)
			if tc.wantBody != "" {
				assert.JSONEq(t, tc.wantBody, rr.Body.String())
			}
		})
	}

}
