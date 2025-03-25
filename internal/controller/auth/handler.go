package auth

import (
	"errors"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/service"
	"github.com/GlebMoskalev/go-todo-api/internal/utils"
	"log/slog"
	"net/http"
)

type Handler struct {
	userService  service.UserService
	tokenService service.TokenService
	logger       *slog.Logger
}

func NewHandler(userService service.UserService, tokenService service.TokenService, logger *slog.Logger) *Handler {
	return &Handler{
		userService:  userService,
		tokenService: tokenService,
		logger:       logger,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "auth_handler", "Register")
	logger.Debug("Attempting to register user")

	var userLogin entity.UserLogin
	if err := utils.DecodeJSONStruct(r, &userLogin); err != nil {
		logger.Warn("Failed to decode JSON", "error", err)
		entity.SendResponse[any](w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	createdUser, err := h.userService.Register(r.Context(), userLogin)
	if err != nil {
		if errors.Is(err, service.ErrUsernameExists) {
			logger.Warn("Username already exists")
			entity.SendResponse[any](w, http.StatusConflict, "Username already exists", nil)
			return
		}
		logger.Error("Failed to create user", "error", err)
		entity.SendResponse[any](w, http.StatusInternalServerError, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse(w, http.StatusOK, "User successfully created", map[string]string{
		"username": createdUser.Username,
	})
	logger.Info("Successfully registered user")
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "auth_handler", "Login")
	logger.Debug("Attempting to login user")

	var userLogin entity.UserLogin
	if err := utils.DecodeJSONStruct(r, &userLogin); err != nil {
		logger.Warn("Failed to decode JSON", "error", err)
		entity.SendResponse[any](w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	user, err := h.userService.GetByUsername(r.Context(), userLogin.Username)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			logger.Warn("User not found")
			entity.SendResponse[any](w, http.StatusUnauthorized, "Invalid credentials", nil)
			return
		}
		logger.Error("Failed to get user", "error", err)
		entity.SendResponse[any](w, http.StatusInternalServerError, entity.ServerFailureMessage, nil)
		return
	}

	if !entity.VerifyPassword(userLogin.Password, user.PasswordHash) {
		logger.Warn("Invalid password")
		entity.SendResponse[any](w, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	accessToken, refreshToken, err := h.tokenService.GenerateTokenPair(user.ID)
	if err != nil {
		logger.Error("Failed to generate tokens", "error", err)
		entity.SendResponse[any](w, http.StatusInternalServerError, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse(w, http.StatusOK, "Login successful", struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{AccessToken: accessToken, RefreshToken: refreshToken})
	logger.Info("Successfully logged in user")
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "auth_handler", "Refresh")
	logger.Debug("Attempting to refresh tokens")

	var refreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := utils.DecodeJSONStruct(r, &refreshRequest); err != nil {
		logger.Warn("Failed to decode JSON", "error", err)
		entity.SendResponse[any](w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	accessToken, refreshToken, err := h.tokenService.RefreshTokens(refreshRequest.RefreshToken)
	if err != nil {
		logger.Warn("Failed to refresh tokens", "error", err)
		entity.SendResponse[any](w, http.StatusUnauthorized, "Invalid refresh token", nil)
		return
	}

	entity.SendResponse(w, http.StatusOK, "Tokens refreshed", map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
	logger.Info("Successfully refreshed tokens")
}
