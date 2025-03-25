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

type UserRepository interface {
	Create(ctx context.Context, user entity.UserLogin) (entity.User, error)
	GetByUsername(ctx context.Context, username string) (entity.User, error)
}

type userRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUserRepository(db *sql.DB, logger *slog.Logger) UserRepository {
	return &userRepository{db: db, logger: logger}
}

func (r *userRepository) Create(ctx context.Context, user entity.UserLogin) (entity.User, error) {
	logger := utils.SetupLogger(ctx, r.logger, "user_repository", "Create", "username", user.Username)
	logger.Debug("Attempting to create user")

	passwordHash, err := entity.HashPassword(user.Password)
	if err != nil {
		logger.Error("Failed to hash password", "error", err)
		return entity.User{}, err
	}

	_, err = r.db.ExecContext(ctx, "INSERT INTO users(username, passwordhash) VALUES($1, $2)",
		user.Username,
		passwordHash,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			logger.Warn("Username already exists")
			return entity.User{}, ErrUsernameExists
		}
		logger.Error("Failed to insert user into database", "error", err)
	}
	logger.Info("Successfully created user")
	return entity.User{
		Username:     user.Username,
		PasswordHash: passwordHash,
	}, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (entity.User, error) {
	logger := utils.SetupLogger(ctx, r.logger, "user_repository", "Create", "username", username)
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
