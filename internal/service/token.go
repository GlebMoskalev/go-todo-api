package service

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-todo-api/internal/config"
	"github.com/GlebMoskalev/go-todo-api/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type TokenService interface {
	GenerateTokenPair(id uuid.UUID) (string, string, error)
	ValidateAccessToken(tokenString string) (uuid.UUID, error)
	ValidateRefreshToken(tokenString string) (uuid.UUID, error)
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

func (s *tokenService) GenerateTokenPair(id uuid.UUID) (string, string, error) {
	accessToken, err := s.generateAccessToken(id)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.generateRefreshToken(id)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *tokenService) generateAccessToken(id uuid.UUID) (string, error) {
	payload := jwt.MapClaims{
		"id":  id,
		"exp": time.Now().UTC().Add(time.Duration(s.config.Token.AccessTokenExpire) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(s.config.Token.AccessTokenSecret))
}

func (s *tokenService) generateRefreshToken(id uuid.UUID) (string, error) {
	payload := jwt.MapClaims{
		"id":  id,
		"exp": time.Now().UTC().Add(time.Duration(s.config.Token.RefreshTokenExpire) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	refreshToken, err := token.SignedString([]byte(s.config.Token.RefreshTokenSecret))
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	err = s.tokenRepo.SaveRefreshToken(ctx, id, refreshToken,
		time.Duration(s.config.Token.RefreshTokenExpire)*time.Minute)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *tokenService) ValidateAccessToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.Token.AccessTokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		return uuid.Parse(claims["id"].(string))
	}
	return uuid.Nil, errors.New("invalid token")
}

func (s *tokenService) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.Token.RefreshTokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return uuid.Parse(claims["id"].(string))
	}
	return uuid.Nil, errors.New("invalid token")
}

func (s *tokenService) RefreshTokens(refreshTokenString string) (string, string, error) {
	id, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	ctx := context.Background()
	isValid, err := s.tokenRepo.ValidateRefreshToken(ctx, id, refreshTokenString)
	if err != nil || !isValid {
		return "", "", errors.New("invalid refresh token")
	}

	err = s.tokenRepo.DeleteRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return "", "", err
	}

	return s.GenerateTokenPair(id)
}
