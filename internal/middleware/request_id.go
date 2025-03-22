package middleware

import (
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func RequestIdHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := middleware.GetReqID(r.Context())
		if requestId != "" {
			w.Header().Set("X-Request-ID", requestId)
		}
		next.ServeHTTP(w, r)
	})
}
