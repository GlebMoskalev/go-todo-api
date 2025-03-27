package auth

import (
	"errors"
	"fmt"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/service"
	"github.com/GlebMoskalev/go-todo-api/internal/utils"
	"log/slog"
	"net/http"
	"strings"
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

// Register handles user registration
// @Summary Register a new user
// @Description Creates a new user with the provided username and password.
// @Tags auth
// @Accept json
// @Produce json
// @Param user body swagger.UserRequest true "User registration data"
// @Success 200 {object} swagger.SuccessRegisterResponse "User successfully created"
// @Failure 400 {object} swagger.ErrorResponse "Invalid request data or validation error"
// @Failure 409 {object} swagger.ConflictResponse "Username already exists"
// @Failure 500 {object} swagger.ServerErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "auth_handler", "Register")
	logger.Debug("Attempting to register user")

	var userLogin entity.UserLogin
	if err := utils.DecodeJSONStruct(r, &userLogin); err != nil {
		logger.Warn("Failed to decode JSON", "error", err)
		entity.SendResponse[any](w, http.StatusBadRequest, true, err.Error(), nil)
		return
	}

	if validationErrors := userLogin.Validate(); validationErrors != nil {
		msg := fmt.Sprintf("Validation error: %s", strings.Join(validationErrors, ";"))
		logger.Warn(msg)
		entity.SendResponse[any](w, http.StatusBadRequest, true, msg, nil)
		return
	}

	createdUser, err := h.userService.Register(r.Context(), userLogin)
	if err != nil {
		if errors.Is(err, entity.ErrUsernameExists) {
			logger.Warn("Username already exists")
			entity.SendResponse[any](w, http.StatusConflict, true, "Username already exists", nil)
			return
		}
		logger.Error("Failed to create user", "error", err)
		entity.SendResponse[any](w, http.StatusInternalServerError, true, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse(w, http.StatusOK, false, "User successfully created", map[string]string{
		"username": createdUser.Username,
	})
	logger.Info("Successfully registered user")
}

// Login handles user login
// @Summary User login
// @Description Authenticates a user and returns access and refresh tokens.
// @Tags auth
// @Accept json
// @Produce json
// @Param user body swagger.UserRequest true "User login credentials"
// @Success 200 {object} swagger.LoginResponse "Login successful"
// @Failure 400 {object} swagger.ErrorResponse "Invalid request data"
// @Failure 401 {object} swagger.UnauthorizedResponse"Invalid credentials"
// @Failure 500 {object} swagger.ServerErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "auth_handler", "Login")
	logger.Debug("Attempting to login user")

	var userLogin entity.UserLogin
	if err := utils.DecodeJSONStruct(r, &userLogin); err != nil {
		logger.Warn("Failed to decode JSON", "error", err)
		entity.SendResponse[any](w, http.StatusBadRequest, true, err.Error(), nil)
		return
	}

	user, err := h.userService.GetByUsername(r.Context(), userLogin.Username)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			logger.Warn("User not found")
			entity.SendResponse[any](w, http.StatusUnauthorized, true, "Invalid credentials", nil)
			return
		}
		logger.Error("Failed to get user", "error", err)
		entity.SendResponse[any](w, http.StatusInternalServerError, true, entity.ServerFailureMessage, nil)
		return
	}

	if !entity.VerifyPassword(userLogin.Password, user.PasswordHash) {
		logger.Warn("Invalid password")
		entity.SendResponse[any](w, http.StatusUnauthorized, true, "Invalid credentials", nil)
		return
	}

	accessToken, refreshToken, err := h.tokenService.GenerateTokenPair(user.ID)
	if err != nil {
		logger.Error("Failed to generate tokens", "error", err)
		entity.SendResponse[any](w, http.StatusInternalServerError, true, entity.ServerFailureMessage, nil)
		return
	}

	entity.SendResponse(w, http.StatusOK, false, "Login successful", struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{AccessToken: accessToken, RefreshToken: refreshToken})
	logger.Info("Successfully logged in user")
}

// Refresh handles token refresh
// @Summary Refresh access and refresh tokens
// @Description Refreshes tokens using a valid refresh token.
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body swagger.RefreshRequest true "Refresh token"
// @Success 200 {object} swagger.RefreshResponse "Tokens refreshed"
// @Failure 400 {object} swagger.ErrorResponse"Invalid request data"
// @Failure 401 {object} swagger.UnauthorizedResponse "Invalid refresh token"
// @Failure 500 {object} swagger.ServerErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	logger := utils.SetupLogger(r.Context(), h.logger, "auth_handler", "Refresh")
	logger.Debug("Attempting to refresh tokens")

	var refreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := utils.DecodeJSONStruct(r, &refreshRequest); err != nil {
		logger.Warn("Failed to decode JSON", "error", err)
		entity.SendResponse[any](w, http.StatusBadRequest, true, err.Error(), nil)
		return
	}

	accessToken, refreshToken, err := h.tokenService.RefreshTokens(refreshRequest.RefreshToken)
	if err != nil {
		logger.Warn("Failed to refresh tokens", "error", err)
		entity.SendResponse[any](w, http.StatusUnauthorized, true, "Invalid refresh token", nil)
		return
	}

	entity.SendResponse(w, http.StatusOK, false, "Tokens refreshed", map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
	logger.Info("Successfully refreshed tokens")
}
