package middleware

import (
	"github.com/GlebMoskalev/go-todo-api/internal/utils/contextutils"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func RequestIdHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := middleware.GetReqID(r.Context())
		if requestId != "" {
			w.Header().Set("X-Request-ID", requestId)
			newCtx := contextutils.WithRequestId(r.Context(), requestId)
			next.ServeHTTP(w, r.WithContext(newCtx))
			return
		}
		next.ServeHTTP(w, r)
	})
}
