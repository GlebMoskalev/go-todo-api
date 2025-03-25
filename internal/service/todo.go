package service

import (
	"context"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"github.com/GlebMoskalev/go-todo-api/internal/repository"
)

type TodoService interface {
	Get(ctx context.Context, username string, id int) (entity.Todo, error)
	Create(ctx context.Context, username string, todo entity.Todo) (int, error)
	Update(ctx context.Context, username string, todo entity.Todo) error
	Delete(ctx context.Context, username string, id int) error
	GetAll(ctx context.Context, username string, pagination entity.Pagination, filters repository.Filters) ([]entity.Todo, int, error)
}

type todoService struct {
	repo repository.TodoRepository
}

func NewTodoService(repo repository.TodoRepository) TodoService {
	return &todoService{repo: repo}
}

func (s *todoService) Get(ctx context.Context, username string, id int) (entity.Todo, error) {
	return s.repo.Get(ctx, username, id)
}

func (s *todoService) Create(ctx context.Context, username string, todo entity.Todo) (int, error) {
	return s.repo.Create(ctx, username, todo)
}

func (s *todoService) Update(ctx context.Context, username string, todo entity.Todo) error {
	return s.repo.Update(ctx, username, todo)
}

func (s *todoService) Delete(ctx context.Context, username string, id int) error {
	return s.repo.Delete(ctx, username, id)
}

func (s *todoService) GetAll(ctx context.Context, username string, pagination entity.Pagination, filters repository.Filters) ([]entity.Todo, int, error) {
	if pagination.Limit > 100 {
		pagination.Limit = 100
	}
	return s.repo.GetAll(ctx, username, pagination, filters)
}
