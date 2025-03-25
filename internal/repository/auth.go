package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/utils"
	"github.com/lib/pq"
	"log/slog"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUsernameExists = errors.New("username already exists")
)

type AuthRepository interface {
	Create(ctx context.Context, user entity.UserLogin) error
	GetByUsername(ctx context.Context, username string) (entity.User, error)
}

type authRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) AuthRepository {
	return &authRepository{db: db, logger: logger}
}

func (r *authRepository) Create(ctx context.Context, user entity.UserLogin) error {
	logger := utils.SetupLogger(ctx, r.logger, "auth_repository", "Create", "username", user.Username)
	logger.Debug("Attempting to create user")

	passwordHash, err := entity.HashPassword(user.Password)
	if err != nil {
		logger.Error("Failed to hash password", "error", err)
	}

	_, err = r.db.ExecContext(ctx, "INSERT INTO users(username, passwordhash) VALUES($1, $2)",
		user.Username,
		passwordHash,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			logger.Warn("Username already exists")
			return ErrUsernameExists
		}
		logger.Error("Failed to insert user into database", "error", err)
	}
	logger.Info("Successfully created user")
	return nil
}

func (r *authRepository) GetByUsername(ctx context.Context, username string) (entity.User, error) {
	logger := utils.SetupLogger(ctx, r.logger, "auth_repository", "Create", "username", username)
	logger.Debug("Attempting to fetch user")

	user := entity.User{
		Username: username,
	}
	err := r.db.QueryRowContext(ctx, "SELECT passwordhash FROM users WHERE username=$1", username).Scan(&user.PasswordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("User not found")
			return entity.User{}, ErrUserNotFound
		}
		logger.Error("Failed to scan user row", "error", err)
		return entity.User{}, err
	}

	logger.Info("Successfully fetched user")
	return user, nil
}
