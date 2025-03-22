package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

const loggerContextKey = "logger"

func RequestIdLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestId := middleware.GetReqID(r.Context())
			if requestId != "" {
				newLogger := logger.With("request_id", requestId)
				newCtx := context.WithValue(r.Context(), loggerContextKey, newLogger)
				next.ServeHTTP(w, r.WithContext(newCtx))
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

func GetLogger(ctx context.Context, defaultLogger *slog.Logger) *slog.Logger {
	if ctx == nil {
		return defaultLogger
	}
	if logger := ctx.Value(loggerContextKey).(*slog.Logger); logger != nil {
		return logger
	}

	return defaultLogger
}
