package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

const requestIDContextKey = "request_id"

func RequestIdHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := middleware.GetReqID(r.Context())
		if requestId != "" {
			w.Header().Set("X-Request-ID", requestId)
			newCtx := context.WithValue(r.Context(), requestIDContextKey, requestId)
			next.ServeHTTP(w, r.WithContext(newCtx))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetRequestId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestId, ok := ctx.Value(requestIDContextKey).(string); ok {
		return requestId
	}
	return ""
}
