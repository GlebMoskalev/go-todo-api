package auth

import (
	"errors"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/repository"
	auth2 "github.com/GlebMoskalev/go-todo-api/internal/service"
	"github.com/GlebMoskalev/go-todo-api/internal/utils"
	"log/slog"
	"net/http"
)

type Handler struct {
	repo         repository.AuthRepository
	tokenService *auth2.AuthService
	logger       *slog.Logger
}

func NewHandler(repo repository.AuthRepository, service *auth2.AuthService, logger *slog.Logger) *Handler {
	return &Handler{
		repo:         repo,
		tokenService: service,
		logger:       logger,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "auth_handler", "Register")
	logger.Debug("Attempting to register user")

	var userLogin entity.UserLogin
	err := utils.DecodeJSONStruct(r, &userLogin)
	if err != nil {
		logger.Warn("Failed to decode json", "error", err)
		entity.SendResponse[any](w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err = h.repo.Create(r.Context(), userLogin)
	if err != nil {
		if errors.Is(err, repository.ErrUsernameExists) {
			logger.Warn("Username already exists")
			entity.SendResponse[any](w, http.StatusConflict, "Username already exists", nil)
			return
		}

		logger.Error("Failed to create user", "error", err)
		entity.SendResponse[any](w, http.StatusInternalServerError, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse[any](w, http.StatusOK, "User successfully created", nil)
	logger.Info("Successfully registered user")
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "auth_handler", "Login")
	logger.Debug("Attempting to login user")

	var userLogin entity.UserLogin
	err := utils.DecodeJSONStruct(r, &userLogin)
	if err != nil {
		logger.Warn("Failed to decode json", "error", err)
		entity.SendResponse[any](w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	user, err := h.repo.GetByUsername(r.Context(), userLogin.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
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
	accessToken, refreshToken, err := h.tokenService.GenerateTokenPair(user.Username)
	if err != nil {
		logger.Error("Failed to generate tokens", "error", err)
		entity.SendResponse[any](w, http.StatusInternalServerError, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse(w, http.StatusOK, "Login successful", struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
	logger.Info("Successfully logged in user")
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "auth_handler", "Refresh")
	logger.Debug("Attempting to refresh tokens")

	var refreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}
	err := utils.DecodeJSONStruct(r, &refreshRequest)
	if err != nil {
		logger.Warn("Failed to decode json", "error", err)
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
