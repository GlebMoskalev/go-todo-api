package middleware

import (
	"context"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/service"
	"net/http"
	"strings"
)

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				entity.SendResponse[any](w, http.StatusUnauthorized, "Missing authorization token", nil)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				entity.SendResponse[any](w, http.StatusUnauthorized, "Invalid token format", nil)
				return
			}

			username, err := authService.ValidateAccessToken(parts[1])
			if err != nil {
				entity.SendResponse[any](w, http.StatusUnauthorized, "Invalid or expired token", nil)
				return
			}

			ctx := context.WithValue(r.Context(), "username", username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
