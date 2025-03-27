package todo

import (
	"errors"
	"fmt"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/service"
	"github.com/GlebMoskalev/go-todo-api/internal/utils"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service service.TodoService
	logger  *slog.Logger
}

func NewHandler(service service.TodoService, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// Get retrieves a todo by ID
// @Summary Get
// @Description Retrieves a todo by its ID for the authenticated user.
// @Tags todo
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Security BearerAuth
// @Success 200 {object} swagger.GetTodoResponse "Successfully create"
// @Failure 400 {object} swagger.InvalidIDResponse "Invalid ID"
// @Failure 401 {object} swagger.UnauthorizedResponse "User not authenticated or invalid token"
// @Failure 404 {object} swagger.NotFoundResponse "Todo not found"
// @Failure 500 {object} swagger.ServerErrorResponse "Internal server error"
// @Router /todos/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "Get")
	logger.Debug("Attempting to fetching todo")

	userID, ok := r.Context().Value("id").(uuid.UUID)
	if !ok {
		logger.Error("Id not found in context")
		entity.SendResponse[any](w, http.StatusUnauthorized, true, "User not authenticated", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("Invalid id", "todo_id", idStr)
		entity.SendResponse[any](w, http.StatusBadRequest, true, "Invalid ID", nil)
		return
	}

	logger = logger.With("todo_id", id)
	todo, err := h.service.Get(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, entity.ErrTodoNotFound) {
			logger.Warn("Todo not found")
			entity.SendResponse[any](w, http.StatusNotFound, true, "Todo not found", nil)
			return
		}

		logger.Error("Failed to get todo", "error", err)
		entity.SendResponse[any](w, http.StatusInternalServerError, true, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse(w, http.StatusOK, false, "Successfully fetch", todo)
	logger.Info("Successfully fetched todo")
}

// Delete removes a todo by ID
// @Summary Delete a todo
// @Description Deletes a todo by its ID for the authenticated user.
// @Tags todo
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Security BearerAuth
// @Success 200 {object} swagger.DeleteResponse "Successfully delete"
// @Failure 400 {object} swagger.InvalidIDResponse "Invalid ID"
// @Failure 401 {object} swagger.UnauthorizedResponse "User not authenticated or invalid token"
// @Failure 404 {object} swagger.NotFoundResponse "Todo not found"
// @Failure 500 {object} swagger.ServerErrorResponse "Internal server error"
// @Router /todos/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "Delete")
	logger.Debug("Attempting to delete todo")

	userID, ok := r.Context().Value("id").(uuid.UUID)
	if !ok {
		logger.Error("Id not found in context")
		entity.SendResponse[any](w, http.StatusUnauthorized, true, "User not authenticated", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("Invalid id", "todo_id", idStr)
		entity.SendResponse[any](w, http.StatusBadRequest, true, "Invalid ID", nil)
		return
	}

	logger = logger.With("todo_id", id)
	err = h.service.Delete(r.Context(), userID, id)
	if err != nil {
		logger.Error("Failed to delete todo", "error", err)
		if errors.Is(err, entity.ErrTodoNotFound) {
			entity.SendResponse[any](w, http.StatusNotFound, true, "Todo not found", nil)
			return
		}
		entity.SendResponse[any](w, http.StatusInternalServerError, true, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse[any](w, http.StatusOK, false, "Successfully delete", nil)
	logger.Info("Successfully delete todo")
}

// Create adds a new todo
// @Summary Create a todo
// @Description Creates a new todo for the authenticated user.
// @Tags todo
// @Accept json
// @Produce json
// @Param todo body swagger.TodoRequest true "Todo data"
// @Security BearerAuth
// @Success 200 {object} swagger.CreateTodoResponse "Todo successfully created"
// @Failure 400 {object} swagger.ErrorResponse "Invalid request data or validation error"
// @Failure 401 {object} swagger.UnauthorizedResponse "User not authenticated or invalid token"
// @Failure 500 {object} swagger.ServerErrorResponse "Internal server error"
// @Router /todos [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "Create")
	logger.Debug("Attempting to create todo")

	userID, ok := r.Context().Value("id").(uuid.UUID)
	if !ok {
		logger.Error("Id not found in context")
		entity.SendResponse[any](w, http.StatusUnauthorized, true, "User not authenticated", nil)
		return
	}

	var todo entity.Todo
	err := utils.DecodeJSONStruct(r, &todo)
	if err != nil {
		logger.Warn("Failed to decode json", "error", err)
		entity.SendResponse[any](w, http.StatusBadRequest, true, err.Error(), nil)
		return
	}

	if validationErrors := todo.Validate(); validationErrors != nil {
		msg := fmt.Sprintf("Validation error: %s", strings.Join(validationErrors, ";"))
		logger.Warn(msg)
		entity.SendResponse[any](w, http.StatusBadRequest, true, msg, nil)
		return
	}

	id, err := h.service.Create(r.Context(), userID, todo)
	if err != nil {
		logger.Error("Failed to create todo")
		entity.SendResponse[any](w, http.StatusInternalServerError, true, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse(w, http.StatusOK, false, "Successfully create", map[string]int{
		"id": id,
	})
	logger.Info("Successfully create todo")
}

// Update modifies an existing todo
// @Summary Update a todo
// @Description Updates an existing todo for the authenticated user.
// @Tags todo
// @Accept json
// @Produce json
// @Param todo body swagger.TodoRequest true "Updated todo data"
// @Security BearerAuth
// @Success 200 {object} swagger.SuccessEmptyResponse "Todo successfully updated"
// @Failure 400 {object} swagger.ErrorResponse "Invalid request data or validation error"
// @Failure 401 {object} swagger.UnauthorizedResponse "User not authenticated or invalid token"
// @Failure 404 {object} swagger.NotFoundResponse "Todo not found"
// @Failure 500 {object} swagger.ServerErrorResponse "Internal server error"
// @Router /todos [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "Update")
	logger.Debug("Attempting to update todo")

	userID, ok := r.Context().Value("id").(uuid.UUID)
	if !ok {
		logger.Error("Id not found in context")
		entity.SendResponse[any](w, http.StatusUnauthorized, true, "User not authenticated", nil)
		return
	}

	var todo entity.Todo
	err := utils.DecodeJSONStruct(r, &todo)
	if err != nil {
		logger.Warn("Failed to decode json", "error", err)
		entity.SendResponse[any](w, http.StatusBadRequest, true, err.Error(), nil)
		return
	}

	if validationErrors := todo.Validate(); validationErrors != nil {
		msg := fmt.Sprintf("Validation error: %s", strings.Join(validationErrors, ";"))
		logger.Warn(msg)
		entity.SendResponse[any](w, http.StatusBadRequest, true, msg, nil)
		return
	}

	err = h.service.Update(r.Context(), userID, todo)
	if err != nil {
		if errors.Is(err, entity.ErrTodoNotFound) {
			logger.Warn("Todo not found")
			entity.SendResponse[any](w, http.StatusNotFound, true, "Todo not found", nil)
			return
		}

		logger.Error("Failed to update todo")
		entity.SendResponse[any](w, http.StatusInternalServerError, true, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse[any](w, http.StatusOK, false, "Successfully update", nil)
	logger.Info("Successfully update todo")
}

// GetAll retrieves all todos with pagination and filters
// @Summary Get all todos
// @Description Retrieves a paginated list of todos for the authenticated user with optional filters.
// @Tags todo
// @Accept json
// @Produce json
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Param due_date query string false "Filter by due date (YYYY-MM-DD)"
// @Param tags query string false "Filter by tags (comma-separated)"
// @Security BearerAuth
// @Success 200 {object} swagger.ListTodoResponse "Todos successfully retrieved"
// @Failure 400 {object} swagger.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} swagger.UnauthorizedResponse "User not authenticated or invalid token"
// @Failure 500 {object} swagger.ServerErrorResponse "Something went wrong, please try again later"
// @Router /todos [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "GetAll")
	logger.Debug("Attempting to get todos")

	userID, ok := r.Context().Value("id").(uuid.UUID)
	if !ok {
		logger.Error("Id not found in context")
		entity.SendResponse[any](w, http.StatusUnauthorized, true, "User not authenticated", nil)
		return
	}

	query := r.URL.Query()
	limit := entity.DefaultLimit
	offset := entity.DefaultOffset

	if limitStr := query.Get("limit"); limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err != nil {
			logger.Warn("Invalid limit parameter", "limit", limitStr)
			entity.SendResponse[any](w, http.StatusBadRequest, true, "Invalid limit parameter", nil)
			return
		}
		limit = limitInt
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		offsetInt, err := strconv.Atoi(offsetStr)
		if err != nil {
			logger.Warn("Invalid limit parameter", "limit", offsetStr)
			entity.SendResponse[any](w, http.StatusBadRequest, true, "Invalid offset parameter", nil)
			return
		}
		offset = offsetInt
	}
	pagination := entity.Pagination{Offset: offset, Limit: limit}

	var filters entity.Filters
	if dueDateStr := query.Get("due_date"); dueDateStr != "" {
		dueDate, err := time.Parse(time.DateOnly, dueDateStr)
		if err != nil {
			logger.Warn("Invalid due_date parameter", "due_date", dueDateStr)
			entity.SendResponse[any](w, http.StatusBadRequest, true, "Invalid due_date format. Use YYYY-MM-DD", nil)
			return
		}
		date := entity.Date{Time: dueDate}
		filters.DueTime = &date
	}

	if tagsStr := query.Get("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		var cleanedTags []string
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				cleanedTags = append(cleanedTags, tag)
			}
		}
		if len(cleanedTags) > 0 {
			filters.Tags = cleanedTags
		}
	}

	todos, total, err := h.service.GetAll(r.Context(), userID, pagination, filters)
	if err != nil {
		logger.Error("Failed to fetch todos", "error", err)
		entity.SendResponse[any](w, http.StatusInternalServerError, true, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendListResponse(w, http.StatusOK, false, "Successfully fetch", pagination, total, todos)
	logger.Info("Successfully fetched todos")
}
