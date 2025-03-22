package todo

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
)

type Handler struct {
	repo   Repository
	logger *slog.Logger
}

func NewHandler(repo Repository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) GetTodo(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With("layer", "todo_handler", "operation", "GetTodo")
	logger.Debug("Attempting to fetching todo")

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid id", "error", err)
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	logger = logger.With("todo_id", id)
	ctx := r.Context()
	todo, err := h.repo.Get(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get todo", "error", err)
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Todo not found"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	jsonTodo, err := json.Marshal(todo)
	if err != nil {
		h.logger.Error("Failed to marshal todo to JSON", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonTodo)
}
