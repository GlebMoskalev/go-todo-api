package repository

import (
	"context"
	"database/sql"
	"github.com/GlebMoskalev/go-todo-api/internal/utils"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type TokenRepository interface {
	SaveRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiry time.Duration) error
	ValidateRefreshToken(ctx context.Context, userID uuid.UUID, token string) (bool, error)
	DeleteRefreshToken(ctx context.Context, userID uuid.UUID) error
}

type tokenRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewTokenRepository(db *sql.DB, logger *slog.Logger) TokenRepository {
	return &tokenRepository{db: db, logger: logger}
}

func (r *tokenRepository) SaveRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiry time.Duration) error {
	logger := utils.SetupLogger(ctx, r.logger, "token_repository", "SaveRefreshToken")

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO refresh_tokens (token, expirydate, userid) VALUES ($1, $2, $3)",
		token,
		time.Now().Add(expiry),
		userID,
	)
	if err != nil {
		logger.Error("Failed to save refresh token", "error", err)
	}
	return nil
}

func (r *tokenRepository) ValidateRefreshToken(ctx context.Context, userID uuid.UUID, token string) (bool, error) {
	logger := utils.SetupLogger(ctx, r.logger, "token_repository", "ValidateRefreshToken")
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM refresh_tokens rt JOIN users u ON rt.userid = u.id WHERE u.id = $1 AND rt.token = $2",
		userID, token,
	).Scan(&count)
	if err != nil {
		logger.Error("Failed to validate refresh token", "error", err)
		return false, err
	}
	return count > 0, nil
}

func (r *tokenRepository) DeleteRefreshToken(ctx context.Context, userID uuid.UUID) error {
	logger := utils.SetupLogger(ctx, r.logger, "token_repository", "DeleteRefreshToken")

	_, err := r.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE userid = $1", userID)
	if err != nil {
		logger.Error("Failed to delete refresh tokens", "error", err)
		return err
	}
	return nil
}
