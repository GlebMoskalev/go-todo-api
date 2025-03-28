package service

import (
	"context"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/repository"
	"github.com/google/uuid"
)

//go:generate go run github.com/vektra/mockery/v2 --name=TodoService --output=./mocks
type TodoService interface {
	Get(ctx context.Context, userID uuid.UUID, id int) (entity.Todo, error)
	Create(ctx context.Context, userID uuid.UUID, todo entity.Todo) (int, error)
	Update(ctx context.Context, userID uuid.UUID, todo entity.Todo) error
	Delete(ctx context.Context, userID uuid.UUID, id int) error
	GetAll(ctx context.Context, userID uuid.UUID, pagination entity.Pagination, filters entity.Filters) ([]entity.Todo, int, error)
}

type todoService struct {
	repo repository.TodoRepository
}

func NewTodoService(repo repository.TodoRepository) TodoService {
	return &todoService{repo: repo}
}

func (s *todoService) Get(ctx context.Context, userID uuid.UUID, id int) (entity.Todo, error) {
	return s.repo.Get(ctx, userID, id)
}

func (s *todoService) Create(ctx context.Context, userID uuid.UUID, todo entity.Todo) (int, error) {
	return s.repo.Create(ctx, userID, todo)
}

func (s *todoService) Update(ctx context.Context, userID uuid.UUID, todo entity.Todo) error {
	return s.repo.Update(ctx, userID, todo)
}

func (s *todoService) Delete(ctx context.Context, userID uuid.UUID, id int) error {
	return s.repo.Delete(ctx, userID, id)
}

func (s *todoService) GetAll(ctx context.Context, userID uuid.UUID, pagination entity.Pagination, filters entity.Filters) ([]entity.Todo, int, error) {
	if pagination.Limit > 100 {
		pagination.Limit = 100
	}
	return s.repo.GetAll(ctx, userID, pagination, filters)
}
