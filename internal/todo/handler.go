package todo

import (
	"errors"
	"fmt"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/entity/response"
	"github.com/GlebMoskalev/go-todo-api/internal/utils"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo   Repository
	logger *slog.Logger
}

func NewHandler(repo Repository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "Get")
	logger.Debug("Attempting to fetching todo")

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("Invalid id", "todo_id", idStr)
		response.SendResponse[any](w, http.StatusBadRequest, "Invalid id", nil)
		return
	}

	logger = logger.With("todo_id", id)
	todo, err := h.repo.Get(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get todo", "error", err)
		if errors.Is(err, ErrNotFound) {
			response.SendResponse[any](w, http.StatusNotFound, "Todo not found", nil)
			return
		}

		response.SendResponse[any](w, http.StatusInternalServerError, response.ServerFailureMessage, nil)
		return
	}

	response.SendResponse(w, http.StatusOK, "Ok", todo)
	logger.Info("Successfully fetched todo")
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "Delete")
	logger.Debug("Attempting to delete todo")

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("Invalid id", "todo_id", idStr)
		response.SendResponse[any](w, http.StatusBadRequest, "Invalid id", nil)
		return
	}

	logger = logger.With("todo_id", id)
	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		logger.Error("Failed to delete todo", "error", err)
		if errors.Is(err, ErrNotFound) {
			response.SendResponse[any](w, http.StatusNotFound, "Todo not found", nil)
			return
		}
		response.SendResponse[any](w, http.StatusInternalServerError, response.ServerFailureMessage, nil)
		return
	}

	response.SendResponse[any](w, http.StatusOK, "Successfully delete", nil)
	logger.Info("Successfully delete todo")
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "Create")
	logger.Debug("Attempting to create todo")

	var todo entity.Todo
	err := utils.DecodeJSONStruct(r, &todo)
	if err != nil {
		logger.Warn("Failed to decode json", "error", err)
		response.SendResponse[any](w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if validationErrors := todo.Validate(); validationErrors != nil {
		msg := fmt.Sprintf("Validation error: %s", strings.Join(validationErrors, ";"))
		logger.Warn(msg)
		response.SendResponse[any](w, http.StatusBadRequest, msg, nil)
		return
	}

	id, err := h.repo.Create(r.Context(), todo)
	if err != nil {
		logger.Error("Failed to create todo")
		response.SendResponse[any](w, http.StatusInternalServerError, response.ServerFailureMessage, nil)
		return
	}

	response.SendResponse(w, http.StatusOK, "Successfully create", struct {
		Id int `json:"id"`
	}{
		Id: id,
	})
	logger.Info("Successfully create todo")
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "Update")
	logger.Debug("Attempting to update todo")

	var todo entity.Todo
	err := utils.DecodeJSONStruct(r, &todo)
	if err != nil {
		logger.Warn("Failed to decode json", "error", err)
		response.SendResponse[any](w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if validationErrors := todo.Validate(); validationErrors != nil {
		msg := fmt.Sprintf("Validation error: %s", strings.Join(validationErrors, ";"))
		logger.Warn(msg)
		response.SendResponse[any](w, http.StatusBadRequest, msg, nil)
		return
	}

	err = h.repo.Update(r.Context(), todo)
	if err != nil {
		logger.Error("Failed to update todo")
		if errors.Is(err, ErrNotFound) {
			response.SendResponse[any](w, http.StatusNotFound, "Todo not found", nil)
			return
		}

		response.SendResponse[any](w, http.StatusInternalServerError, response.ServerFailureMessage, nil)
		return
	}

	response.SendResponse[any](w, http.StatusOK, "Successfully update", nil)
	logger.Info("Successfully update todo")
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "todo_handler", "GetAll")
	logger.Debug("Attempting to get todos")

	query := r.URL.Query()
	limit := entity.DefaultLimit
	offset := entity.DefaultOffset

	if limitStr := query.Get("limit"); limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err != nil {
			logger.Warn("Invalid limit parameter", "limit", limitStr)
			response.SendResponse[any](w, http.StatusBadRequest, "Invalid limit parameter", nil)
			return
		}
		limit = limitInt
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		offsetInt, err := strconv.Atoi(offsetStr)
		if err != nil {
			logger.Warn("Invalid limit parameter", "limit", offsetStr)
			response.SendResponse[any](w, http.StatusBadRequest, "Invalid offset parameter", nil)
			return
		}
		offset = offsetInt
	}
	pagination := entity.Pagination{Offset: offset, Limit: limit}

	var filters Filters
	if dueDateStr := query.Get("due_date"); dueDateStr != "" {
		dueDate, err := time.Parse(time.DateOnly, dueDateStr)
		if err != nil {
			logger.Warn("Invalid due_date parameter", "due_date", dueDateStr)
			response.SendResponse[any](w, http.StatusBadRequest, "Invalid due_date format. Use YYYY-MM-DD", nil)
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

	todos, total, err := h.repo.GetAll(r.Context(), pagination, filters)
	if err != nil {
		logger.Error("Failed to fetch todos", "error", err)
		response.SendResponse[any](w, http.StatusInternalServerError, response.ServerFailureMessage, nil)
		return
	}

	response.SendListResponse(w, http.StatusOK, "Ok", pagination, total, todos)
	logger.Info("Successfully fetched todos")
}
