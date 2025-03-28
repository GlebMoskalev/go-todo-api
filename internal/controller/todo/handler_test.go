package todo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	"strings"
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

func TestCreate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	userID := uuid.New()

	testCases := []struct {
		name                string
		inputRequest        string
		inputToken          string
		prepareTodoService  func(serviceMock *mocks.TodoService)
		prepareTokenService func(serviceMock *mocks.TokenService)
		setupMiddleware     func(mux *chi.Mux, tokenServiceMock *mocks.TokenService)
		expectedHTTPStatus  int
		expectedResponse    string
	}{
		{
			name:         "successful creation",
			inputRequest: `{"title":"test","description":"test","due_date":"2025-04-01","tags":["api","test"]}`,
			inputToken:   "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("Create", mock.MatchedBy(func(ctx context.Context) bool {
					_, ok := ctx.Value("id").(uuid.UUID)
					return ok
				}), userID, mock.AnythingOfType("entity.Todo")).
					Return(12, nil)
			},
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusCreated,
			expectedResponse:   `{"code":201,"error":false,"message":"Successfully create","data":{"id":12}}`,
		},
		{
			name:               "user not authenticated",
			inputRequest:       `{"title":"test","description":"test","due_date":"2025-04-01","tags":["api","test"]}`,
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedResponse:   `{"code":401,"error":true,"message":"User not authenticated"}`,
		},
		{
			name:         "invalid todo",
			inputRequest: `{"description":"test","due_date":"2025-04-01","tags":["api","test"]}`,
			inputToken:   "valid_token",
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Validation error: Field 'title' is required"}`,
		},
		{
			name:         "internal server error",
			inputRequest: `{"title":"test","description":"test","due_date":"2025-04-01","tags":["api","test"]}`,
			inputToken:   "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("Create", mock.MatchedBy(func(ctx context.Context) bool {
					_, ok := ctx.Value("id").(uuid.UUID)
					return ok
				}), userID, mock.AnythingOfType("entity.Todo")).
					Return(0, errors.New("database error"))
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
			name:         "invalid json body",
			inputRequest: `{"title":"test","description":"test","due_date":"2025-04-01","tags":"not_an_array"}`, // Некорректный тип для tags
			inputToken:   "valid_token",
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Invalid json format"}`,
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
			r.Post("/todos", handler.Create)

			req, err := http.NewRequest("POST", "/todos", bytes.NewBufferString(tc.inputRequest))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tc.inputToken)

			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)
			assert.Equal(t, tc.expectedHTTPStatus, rr.Code)
			assert.JSONEq(t, tc.expectedResponse, rr.Body.String())
		})
	}
}

func TestUpdate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	userID := uuid.New()

	testCases := []struct {
		name                string
		inputRequest        string
		inputToken          string
		prepareTodoService  func(serviceMock *mocks.TodoService)
		prepareTokenService func(serviceMock *mocks.TokenService)
		setupMiddleware     func(mux *chi.Mux, tokenServiceMock *mocks.TokenService)
		expectedHTTPStatus  int
		expectedResponse    string
	}{
		{
			name:         "successful update",
			inputRequest: `{"id":12,"title":"test","description":"test","due_date":"2025-04-02","tags":["updated","test"]}`,
			inputToken:   "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("Update", mock.MatchedBy(func(ctx context.Context) bool {
					_, ok := ctx.Value("id").(uuid.UUID)
					return ok
				}), userID, mock.AnythingOfType("entity.Todo")).
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
			expectedResponse:   `{"code":200,"error":false,"message":"Successfully update"}`,
		},
		{
			name:               "missing authorization",
			inputRequest:       `{"id":12,"title":"test","description":"test","due_date":"2025-04-02","tags":["updated","test"]}`,
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedResponse:   `{"code":401,"error":true,"message":"User not authenticated"}`,
		},
		{
			name:         "invalid todo",
			inputRequest: `{"id":12,"description":"updated desc","due_date":"2025-04-02","tags":["updated","test"]}`,
			inputToken:   "valid_token",
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Validation error: Field 'title' is required"}`,
		},
		{
			name:         "todo not found",
			inputRequest: `{"id":12,"title":"updated","description":"updated desc","due_date":"2025-04-02","tags":["updated","test"]}`,
			inputToken:   "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("Update", mock.MatchedBy(func(ctx context.Context) bool {
					_, ok := ctx.Value("id").(uuid.UUID)
					return ok
				}), userID, mock.AnythingOfType("entity.Todo")).
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
			name:         "internal server",
			inputRequest: `{"id":12,"title":"updated","description":"updated desc","due_date":"2025-04-02","tags":["updated","test"]}`,
			inputToken:   "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("Update", mock.MatchedBy(func(ctx context.Context) bool {
					_, ok := ctx.Value("id").(uuid.UUID)
					return ok
				}), userID, mock.AnythingOfType("entity.Todo")).
					Return(errors.New("internal server"))
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
			name:         "invalid json body",
			inputRequest: `{"id":12,"title":"updated","description":"updated desc","due_date":"2025-04-02","tags":"not_an_array"}`,
			inputToken:   "valid_token",
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse:   `{"code":400,"error":true,"message":"Invalid json format"}`,
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
			r.Post("/todos", handler.Update)

			req, err := http.NewRequest("POST", "/todos", bytes.NewBufferString(tc.inputRequest))
			if err != nil {
				t.Fatalf("Failed to update request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tc.inputToken)

			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)
			assert.Equal(t, tc.expectedHTTPStatus, rr.Code)
			assert.JSONEq(t, tc.expectedResponse, rr.Body.String())
		})
	}
}

func TestGetAll(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	userID := uuid.New()

	testCases := []struct {
		name                string
		queryParams         map[string]string
		inputToken          string
		prepareTodoService  func(serviceMock *mocks.TodoService)
		prepareTokenService func(serviceMock *mocks.TokenService)
		setupMiddleware     func(mux *chi.Mux, tokenServiceMock *mocks.TokenService)
		expectedHTTPStatus  int
		expectedResponse    string
	}{
		{
			name: "successful get all todos",
			queryParams: map[string]string{
				"limit":  "10",
				"offset": "0",
			},
			inputToken: "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("GetAll",
					mock.MatchedBy(func(ctx context.Context) bool {
						_, ok := ctx.Value("id").(uuid.UUID)
						return ok
					}),
					userID,
					entity.Pagination{Offset: 0, Limit: 10},
					entity.Filters{},
				).Return(
					[]entity.Todo{
						{
							ID:          12,
							Title:       "Buy groceries",
							Description: "Get milk, bread, and eggs",
							DueDate:     &entity.Date{Time: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)},
							Tags:        []string{"shopping", "urgent"},
						},
					},
					1,
					nil,
				)
			},
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusOK,
			expectedResponse: `{
                "code": 200,
                "count": 1,
                "data": [
                    {
                        "id": 12,
                        "title": "Buy groceries",
                        "description": "Get milk, bread, and eggs",
                        "due_date": "2025-04-01",
                        "tags": ["shopping", "urgent"]
                    }
                ],
                "error": false,
                "limit": 10,
                "message": "Successfully fetch",
                "offset": 0,
                "total": 1
            }`,
		},
		{
			name: "invalid query parameters - limit",
			queryParams: map[string]string{
				"limit":  "invalid",
				"offset": "0",
			},
			inputToken: "valid_token",
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse: `{
                "code": 400,
                "error": true,
                "message": "Invalid limit parameter"
            }`,
		},
		{
			name:               "missing authorization",
			queryParams:        map[string]string{},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedResponse: `{
                "code": 401,
                "error": true,
                "message": "User not authenticated"
            }`,
		},
		{
			name: "invalid query parameters - limit",
			queryParams: map[string]string{
				"limit":  "invalid",
				"offset": "0",
			},
			inputToken: "valid_token",
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResponse: `{
                "code": 400,
                "error": true,
                "message": "Invalid limit parameter"
            }`,
		},
		{
			name:               "missing authorization",
			queryParams:        map[string]string{},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedResponse: `{
                "code": 401,
                "error": true,
                "message": "User not authenticated"
            }`,
		},
		{
			name: "internal server error",
			queryParams: map[string]string{
				"limit":  "20",
				"offset": "0",
			},
			inputToken: "valid_token",
			prepareTodoService: func(serviceMock *mocks.TodoService) {
				serviceMock.On("GetAll",
					mock.MatchedBy(func(ctx context.Context) bool {
						_, ok := ctx.Value("id").(uuid.UUID)
						return ok
					}),
					userID,
					entity.Pagination{Offset: 0, Limit: 20},
					entity.Filters{},
				).Return(
					[]entity.Todo{},
					0,
					errors.New("database error"),
				)
			},
			prepareTokenService: func(serviceMock *mocks.TokenService) {
				serviceMock.On("ValidateAccessToken", "valid_token").
					Return(userID, nil)
			},
			setupMiddleware: func(mux *chi.Mux, tokenServiceMock *mocks.TokenService) {
				mux.Use(middleware.AuthMiddleware(tokenServiceMock))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedResponse: `{
                "code": 500,
                "error": true,
                "message": "Something went wrong, please try again later"
            }`,
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
			r.Get("/todos", handler.GetAll)

			url := "/todos"
			if len(tc.queryParams) > 0 {
				queryParts := []string{}
				for k, v := range tc.queryParams {
					queryParts = append(queryParts, fmt.Sprintf("%s=%s", k, v))
				}
				url += "?" + strings.Join(queryParts, "&")
			}

			req, err := http.NewRequest("GET", url, nil)
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
