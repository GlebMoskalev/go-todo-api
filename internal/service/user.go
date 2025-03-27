package service

import (
	"context"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/repository"
	"github.com/google/uuid"
	"log/slog"
)

//go:generate go run github.com/vektra/mockery/v2 --name=UserService --output=./mocks
type UserService interface {
	Register(ctx context.Context, user entity.UserLogin) (entity.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (entity.User, error)
	GetByUsername(ctx context.Context, username string) (entity.User, error)
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

func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (entity.User, error) {
	return s.repo.Get(ctx, id)
}

func (s *userService) GetByUsername(ctx context.Context, username string) (entity.User, error) {
	return s.repo.GetByUsername(ctx, username)
}
