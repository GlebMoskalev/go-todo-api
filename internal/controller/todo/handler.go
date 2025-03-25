package todo

import (
	"errors"
	"fmt"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	todo2 "github.com/GlebMoskalev/go-todo-api/internal/repository"
	"github.com/GlebMoskalev/go-todo-api/internal/utils"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo   todo2.TodoRepository
	logger *slog.Logger
}

func NewHandler(repo todo2.TodoRepository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "Get")
	logger.Debug("Attempting to fetching todo")

	username, ok := r.Context().Value("username").(string)
	if !ok {
		logger.Error("Username not found in contextutils")
		entity.SendResponse[any](w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("Invalid id", "todo_id", idStr)
		entity.SendResponse[any](w, http.StatusBadRequest, "Invalid id", nil)
		return
	}

	logger = logger.With("todo_id", id, "username", username)
	todo, err := h.repo.Get(r.Context(), username, id)
	if err != nil {
		if errors.Is(err, todo2.ErrNotFound) {
			logger.Warn("Todo not found")
			entity.SendResponse[any](w, http.StatusNotFound, "Todo not found", nil)
			return
		}

		logger.Error("Failed to get todo", "error", err)
		entity.SendResponse[any](w, http.StatusInternalServerError, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse(w, http.StatusOK, "Ok", todo)
	logger.Info("Successfully fetched todo")
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "Delete")
	logger.Debug("Attempting to delete todo")

	username, ok := r.Context().Value("username").(string)
	if !ok {
		logger.Error("Username not found in contextutils")
		entity.SendResponse[any](w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("Invalid id", "todo_id", idStr)
		entity.SendResponse[any](w, http.StatusBadRequest, "Invalid id", nil)
		return
	}

	logger = logger.With("todo_id", id, "username", username)
	err = h.repo.Delete(r.Context(), username, id)
	if err != nil {
		logger.Error("Failed to delete todo", "error", err)
		if errors.Is(err, todo2.ErrNotFound) {
			entity.SendResponse[any](w, http.StatusNotFound, "Todo not found", nil)
			return
		}
		entity.SendResponse[any](w, http.StatusInternalServerError, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse[any](w, http.StatusOK, "Successfully delete", nil)
	logger.Info("Successfully delete todo")
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "Create")
	logger.Debug("Attempting to create todo")

	username, ok := r.Context().Value("username").(string)
	if !ok {
		logger.Error("Username not found in contextutils")
		entity.SendResponse[any](w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	logger = logger.With("username", username)

	var todo entity.Todo
	err := utils.DecodeJSONStruct(r, &todo)
	if err != nil {
		logger.Warn("Failed to decode json", "error", err)
		entity.SendResponse[any](w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if validationErrors := todo.Validate(); validationErrors != nil {
		msg := fmt.Sprintf("Validation error: %s", strings.Join(validationErrors, ";"))
		logger.Warn(msg)
		entity.SendResponse[any](w, http.StatusBadRequest, msg, nil)
		return
	}

	id, err := h.repo.Create(r.Context(), username, todo)
	if err != nil {
		logger.Error("Failed to create todo")
		entity.SendResponse[any](w, http.StatusInternalServerError, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse(w, http.StatusOK, "Successfully create", struct {
		Id int `json:"id"`
	}{
		Id: id,
	})
	logger.Info("Successfully create todo")
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "Update")
	logger.Debug("Attempting to update todo")

	username, ok := r.Context().Value("username").(string)
	if !ok {
		logger.Error("Username not found in contextutils")
		entity.SendResponse[any](w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	logger = logger.With("username", username)

	var todo entity.Todo
	err := utils.DecodeJSONStruct(r, &todo)
	if err != nil {
		logger.Warn("Failed to decode json", "error", err)
		entity.SendResponse[any](w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if validationErrors := todo.Validate(); validationErrors != nil {
		msg := fmt.Sprintf("Validation error: %s", strings.Join(validationErrors, ";"))
		logger.Warn(msg)
		entity.SendResponse[any](w, http.StatusBadRequest, msg, nil)
		return
	}

	err = h.repo.Update(r.Context(), username, todo)
	if err != nil {
		if errors.Is(err, todo2.ErrNotFound) {
			logger.Error("Todo not found")
			entity.SendResponse[any](w, http.StatusNotFound, "Todo not found", nil)
			return
		}

		logger.Error("Failed to update todo")
		entity.SendResponse[any](w, http.StatusInternalServerError, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse[any](w, http.StatusOK, "Successfully update", nil)
	logger.Info("Successfully update todo")
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "GetAll")
	logger.Debug("Attempting to get todos")

	username, ok := r.Context().Value("username").(string)
	if !ok {
		logger.Error("Username not found in contextutils")
		entity.SendResponse[any](w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	logger = logger.With("username", username)

	query := r.URL.Query()
	limit := entity.DefaultLimit
	offset := entity.DefaultOffset

	if limitStr := query.Get("limit"); limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err != nil {
			logger.Warn("Invalid limit parameter", "limit", limitStr)
			entity.SendResponse[any](w, http.StatusBadRequest, "Invalid limit parameter", nil)
			return
		}
		limit = limitInt
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		offsetInt, err := strconv.Atoi(offsetStr)
		if err != nil {
			logger.Warn("Invalid limit parameter", "limit", offsetStr)
			entity.SendResponse[any](w, http.StatusBadRequest, "Invalid offset parameter", nil)
			return
		}
		offset = offsetInt
	}
	pagination := entity.Pagination{Offset: offset, Limit: limit}

	var filters todo2.Filters
	if dueDateStr := query.Get("due_date"); dueDateStr != "" {
		dueDate, err := time.Parse(time.DateOnly, dueDateStr)
		if err != nil {
			logger.Warn("Invalid due_date parameter", "due_date", dueDateStr)
			entity.SendResponse[any](w, http.StatusBadRequest, "Invalid due_date format. Use YYYY-MM-DD", nil)
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

	todos, total, err := h.repo.GetAll(r.Context(), username, pagination, filters)
	if err != nil {
		logger.Error("Failed to fetch todos", "error", err)
		entity.SendResponse[any](w, http.StatusInternalServerError, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendListResponse(w, http.StatusOK, "Ok", pagination, total, todos)
	logger.Info("Successfully fetched todos")
}
