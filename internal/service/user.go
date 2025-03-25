package service

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/repository"
	"log/slog"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUsernameExists = errors.New("username already exists")
)

type UserService interface {
	Register(ctx context.Context, user entity.UserLogin) (entity.User, error)
	GetUser(ctx context.Context, username string) (entity.User, error)
}

type userService struct {
	repo   repository.UserRepository
	logger *slog.Logger
}

func NewUserService(repo repository.UserRepository, logger *slog.Logger) UserService {
	return &userService{repo: repo, logger: logger}
}

func (s *userService) Register(ctx context.Context, user entity.UserLogin) (entity.User, error) {
	return s.repo.Create(ctx, user)
}

func (s *userService) GetUser(ctx context.Context, username string) (entity.User, error) {
	return s.repo.GetByUsername(ctx, username)
}
