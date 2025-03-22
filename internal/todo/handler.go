package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/entity/response"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/GlebMoskalev/go-todo-api/internal/middleware"
)

type Handler struct {
	repo   Repository
	logger *slog.Logger
}

func NewHandler(repo Repository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context(), h.logger)
	logger = logger.With("layer", "todo_handler", "operation", "Get")
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
	logger := middleware.GetLogger(r.Context(), h.logger)
	logger = logger.With("layer", "todo_handler", "operation", "Delete")
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
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context(), h.logger)
	logger = logger.With("layer", "todo_handler", "operation", "Create")
	logger.Debug("Attempting to create todo")

	var todo entity.Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		logger.Warn("Failed to decode json")
		response.SendResponse[any](w, http.StatusBadRequest, "Invalid todo", nil)
		return
	}

	if validationErrors := entity.ValidateTodo(todo); validationErrors != nil {
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
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context(), h.logger)
	logger = logger.With("layer", "todo_handler", "operation", "Update")
	logger.Debug("Attempting to update todo")

	var todo entity.Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		logger.Warn("Failed to decode json")
		response.SendResponse[any](w, http.StatusBadRequest, "Invalid todo", nil)
		return
	}

	if validationErrors := entity.ValidateTodo(todo); validationErrors != nil {
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
}
