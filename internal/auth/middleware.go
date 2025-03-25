package auth

import (
	"context"
	"github.com/GlebMoskalev/go-todo-api/internal/entity/response"
	"net/http"
	"strings"
)

func (ts *TokenService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.SendResponse[any](w, http.StatusUnauthorized, "Missing authorization token", nil)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.SendResponse[any](w, http.StatusUnauthorized, "Invalid token format", nil)
			return
		}

		username, err := ts.ValidateAccessToken(parts[1])
		if err != nil {
			response.SendResponse[any](w, http.StatusUnauthorized, "Invalid or expired token", nil)
			return
		}
		ctx := context.WithValue(r.Context(), "username", username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
