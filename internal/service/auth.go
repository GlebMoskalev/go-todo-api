package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/GlebMoskalev/go-todo-api/internal/config"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"time"
)

type AuthService struct {
	db     *sql.DB
	config config.Config
	logger *slog.Logger
}

func NewTokenService(db *sql.DB, config config.Config, logger *slog.Logger) *AuthService {
	return &AuthService{
		db:     db,
		config: config,
		logger: logger,
	}
}

func (ts *AuthService) GenerateTokenPair(username string) (string, string, error) {
	accessToken, err := ts.generateAccessToken(username)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := ts.generateRefreshToken(username)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (ts *AuthService) generateAccessToken(username string) (string, error) {
	payload := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Duration(ts.config.Token.AccessTokenExpire) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(ts.config.Token.AccessTokenSecret))
}

func (ts *AuthService) generateRefreshToken(username string) (string, error) {
	payload := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Duration(ts.config.Token.RefreshTokenExpire) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	refreshToken, err := token.SignedString([]byte(ts.config.Token.RefreshTokenSecret))
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	err = ts.saveRefreshToken(ctx, username, refreshToken)
	return refreshToken, err
}

func (ts *AuthService) saveRefreshToken(ctx context.Context, username, token string) error {
	logger := utils.SetupLogger(ctx, ts.logger, "token_service", "saveRefreshToken")

	var userID int
	err := ts.db.QueryRowContext(ctx, "SELECT id FROM users WHERE username=$1",
		username).Scan(&userID)
	if err != nil {
		logger.Error("Failed to find user ID", "error", err, "username", username)
		return err
	}
	_, err = ts.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE userid=$1", userID)
	if err != nil {
		logger.Error("Failed to delete old refresh tokens", "error", err)
		return err
	}
	_, err = ts.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (TOKEN, ExpiryDate, UserId) 
		 VALUES ($1, $2, $3)`,
		token,
		time.Now().Add(time.Duration(ts.config.Token.RefreshTokenExpire)*time.Minute),
		userID,
	)
	if err != nil {
		logger.Error("Failed to save refresh token", "error", err)
		return err
	}

	return nil
}

func (ts *AuthService) ValidateAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(ts.config.Token.AccessTokenSecret), nil
	})

	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["username"].(string), err
	}
	return "", errors.New("invalid token")
}

func (ts *AuthService) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(ts.config.Token.RefreshTokenSecret), nil
	})

	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["username"].(string), err
	}
	return "", errors.New("invalid token")
}

func (ts *AuthService) RefreshTokens(refreshTokenString string) (string, string, error) {
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(ts.config.Token.RefreshTokenSecret), nil
	})

	if err != nil {
		return "", "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", errors.New("invalid refresh token")
	}

	username := claims["username"].(string)
	var count int
	err = ts.db.QueryRow("SELECT COUNT(*) FROM refresh_tokens rt JOIN users u ON rt.userid = u.id "+
		"WHERE u.username = $1 AND rt.token = $2", username, refreshTokenString).Scan(&count)
	if err != nil || count == 0 {
		return "", "", errors.New("refresh token not found")
	}

	user, err := ts.getUserByUsername(username)
	if err != nil {
		return "", "", err
	}
	return ts.GenerateTokenPair(user.Username)
}

func (ts *AuthService) getUserByUsername(username string) (entity.User, error) {
	var user entity.User
	err := ts.db.QueryRow(
		"SELECT username, passwordhash FROM users WHERE username = $1",
		username,
	).Scan(&user.Username, &user.PasswordHash)

	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}
