package todo

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/middleware"
	"github.com/GlebMoskalev/go-todo-api/internal/service/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	userID := uuid.New()

	testCases := []struct {
		name                string
		inputID             string
		inputToken          string
		prepareTodoService  func(serviceMock *mocks.TodoService)
		prepareTokenService func(serviceMock *mocks.TokenService)
		setupMiddleware     func(mux *chi.Mux, tokenServiceMock *mocks.TokenService)
		expectedHTTPStatus  int
		expectedResponse    string
	}{
		{
			name:       "successful get",
			inputID:    "12",
			inputToken: "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("Get", mock.MatchedBy(func(ctx context.Context) bool {
					_, ok := ctx.Value("id").(uuid.UUID)
					return ok
				}), userID, 12).
					Return(entity.Todo{
						ID:          12,
						Title:       "test_todo",
						Description: "test_description",
						Tags:        []string{"test"},
						DueDate:     &entity.Date{Time: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)},
					}, nil)
			},
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   `{"code":200,"error":false,"message":"Successfully fetch","data":{"id":12,"title":"test_todo","description":"test_description","due_date":"2025-04-01","tags":["test"]}}`,
		},
		{
			name:               "id not found in context",
			inputID:            "12",
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedResponse:   `{"code":401,"error":true,"message":"User not authenticated"}`,
		},
		{
			name:       "invalid id",
			inputID:    "12f",
			inputToken: "valid_token",
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Invalid ID"}`,
		},
		{
			name:       "todo not found",
			inputID:    "12",
			inputToken: "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("Get", mock.MatchedBy(func(ctx context.Context) bool {
					_, ok := ctx.Value("id").(uuid.UUID)
					return ok
				}), userID, 12).
					Return(entity.Todo{}, entity.ErrTodoNotFound)
			},
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusNotFound,
			expectedResponse:   `{"code":404,"error":true,"message":"Todo not found"}`,
		},
		{
			name:       "internal user server error",
			inputID:    "12",
			inputToken: "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("Get", mock.MatchedBy(func(ctx context.Context) bool {
					_, ok := ctx.Value("id").(uuid.UUID)
					return ok
				}), userID, 12).
					Return(entity.Todo{}, errors.New("internal user server error"))
			},
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   `{"code":500,"error":true,"message":"Something went wrong, please try again later"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			todoServiceMock := mocks.NewTodoService(t)
			tokenServiceMock := mocks.NewTokenService(t)

			if tc.prepareTokenService != nil {
				tc.prepareTokenService(tokenServiceMock)
			}
			if tc.prepareTodoService != nil {
				tc.prepareTodoService(todoServiceMock)
			}

			handler := NewHandler(todoServiceMock, logger)

			r := chi.NewRouter()
			if tc.setupMiddleware != nil {
				tc.setupMiddleware(r, tokenServiceMock)
			}

			r.Get("/todos/{id}", handler.Get)

			req, err := http.NewRequest("GET", "/todos/"+tc.inputID, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Authorization", "Bearer "+tc.inputToken)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedHTTPStatus, rr.Code)
			assert.JSONEq(t, tc.expectedResponse, rr.Body.String())
		})
	}
}

func TestDelete(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	userID := uuid.New()

	testCases := []struct {
		name                string
		inputID             string
		inputToken          string
		prepareTodoService  func(serviceMock *mocks.TodoService)
		prepareTokenService func(serviceMock *mocks.TokenService)
		setupMiddleware     func(mux *chi.Mux, tokenServiceMock *mocks.TokenService)
		expectedHTTPStatus  int
		expectedResponse    string
	}{
		{
			name:       "successful delete",
			inputID:    "12",
			inputToken: "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("Delete", mock.MatchedBy(func(ctx context.Context) bool {
					_, ok := ctx.Value("id").(uuid.UUID)
					return ok
				}), userID, 12).
					Return(nil)
			},
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   `{"code":200,"error":false,"message":"Successfully delete"}`,
		},
		{
			name:       "invalid id",
			inputID:    "invalid",
			inputToken: "valid_token",
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Invalid ID"}`,
		},
		{
			name:       "todo not found",
			inputID:    "12",
			inputToken: "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("Delete", mock.MatchedBy(func(ctx context.Context) bool {
					_, ok := ctx.Value("id").(uuid.UUID)
					return ok
				}), userID, 12).
					Return(entity.ErrTodoNotFound)
			},
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusNotFound,
			expectedResponse:   `{"code":404,"error":true,"message":"Todo not found"}`,
		},
		{
			name:       "internal server error",
			inputID:    "12",
			inputToken: "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("Delete", mock.MatchedBy(func(ctx context.Context) bool {
					_, ok := ctx.Value("id").(uuid.UUID)
					return ok
				}), userID, 12).
					Return(errors.New("unexpected error"))
			},
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse:   `{"code":500,"error":true,"message":"Something went wrong, please try again later"}`,
		},
		{
			name:               "user not authenticated",
			inputID:            "12",
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedResponse:   `{"code":401,"error":true,"message":"User not authenticated"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			todoServiceMock := mocks.NewTodoService(t)
			tokenServiceMock := mocks.NewTokenService(t)

			if tc.prepareTodoService != nil {
				tc.prepareTodoService(todoServiceMock)
			}
			if tc.prepareTokenService != nil {
				tc.prepareTokenService(tokenServiceMock)
			}

			handler := NewHandler(todoServiceMock, logger)

			r := chi.NewRouter()
			if tc.setupMiddleware != nil {
				tc.setupMiddleware(r, tokenServiceMock)
			}
			r.Delete("/todos/{id}", handler.Delete)

			req, err := http.NewRequest("DELETE", "/todos/"+tc.inputID, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Authorization", "Bearer "+tc.inputToken)

			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)
			assert.Equal(t, tc.expectedHTTPStatus, rr.Code)
			assert.JSONEq(t, tc.expectedResponse, rr.Body.String())
		})
	}
}
