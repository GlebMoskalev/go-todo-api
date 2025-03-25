package service

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-todo-api/internal/config"
	"github.com/GlebMoskalev/go-todo-api/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"time"
)

type TokenService interface {
	GenerateTokenPair(username string, userID int) (string, string, error)
	ValidateAccessToken(tokenString string) (string, error)
	ValidateRefreshToken(tokenString string) (string, error)
	RefreshTokens(refreshTokenString string) (string, string, error)
}

type tokenService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	config    config.Config
	logger    *slog.Logger
}

func NewTokenService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository,
	config config.Config, logger *slog.Logger) TokenService {
	return &tokenService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		config:    config,
		logger:    logger,
	}
}

func (s *tokenService) GenerateTokenPair(username string, userID int) (string, string, error) {
	accessToken, err := s.generateAccessToken(username)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.generateRefreshToken(username, userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *tokenService) generateAccessToken(username string) (string, error) {
	payload := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Duration(s.config.Token.AccessTokenExpire) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(s.config.Token.AccessTokenSecret))
}

func (s *tokenService) generateRefreshToken(username string, userID int) (string, error) {
	payload := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Duration(s.config.Token.RefreshTokenExpire) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	refreshToken, err := token.SignedString([]byte(s.config.Token.RefreshTokenSecret))
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	err = s.tokenRepo.SaveRefreshToken(ctx, userID, refreshToken,
		time.Duration(s.config.Token.RefreshTokenExpire)*time.Minute)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *tokenService) ValidateAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.Token.AccessTokenSecret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["username"].(string), nil
	}
	return "", errors.New("invalid token")
}

func (s *tokenService) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.Token.RefreshTokenSecret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["username"].(string), nil
	}
	return "", errors.New("invalid token")
}

func (s *tokenService) RefreshTokens(refreshTokenString string) (string, string, error) {
	username, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	ctx := context.Background()
	isValid, err := s.tokenRepo.ValidateRefreshToken(ctx, username, refreshTokenString)
	if err != nil || !isValid {
		return "", "", errors.New("invalid refresh token")
	}

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}

	err = s.tokenRepo.DeleteRefreshToken(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	return s.GenerateTokenPair(user.Username, user.ID)
}
