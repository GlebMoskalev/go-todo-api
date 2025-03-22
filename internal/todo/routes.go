package todo

import (
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/{id}", h.Get)
	r.Get("/", h.GetAll)
	r.Delete("/{id}", h.Delete)
	r.Post("/", h.Create)
	r.Put("/", h.Update)
}
