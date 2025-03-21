package todo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log/slog"
	"strings"

	"github.com/GlebMoskalev/go-todo-api/internal/entity"
)

type Filters struct {
	DueTime *entity.Date
	Tags    []string
}

type Repository interface {
	Get(ctx context.Context, id int) (entity.Todo, error)
	Create(ctx context.Context, todo entity.Todo) (int, error)
	Update(ctx context.Context, todo entity.Todo) error
	Delete(ctx context.Context, id int) error
	GetAll(ctx context.Context, pagination entity.Pagination, filters Filters) ([]entity.Todo, error)
}

type repository struct {
	db     *sql.DB
	logger slog.Logger
}

func NewRepository(db *sql.DB, logger slog.Logger) Repository {
	return &repository{db: db, logger: logger}
}

func (r *repository) Get(ctx context.Context, id int) (entity.Todo, error) {
	logger := r.logger.With("layer", "repository", "operation", "Get", "todo_id", id)
	logger.Debug("Attempting to fetching todo")

	row := r.db.QueryRowContext(ctx, "SELECT id, title, tags, description, duetime FROM todos WHERE id = $1", id)

	var todo entity.Todo
	if err := row.Scan(&todo.ID, &todo.Title, pq.Array(&todo.Tags), &todo.Description, &todo.DueTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("Todo not found")
			return entity.Todo{}, fmt.Errorf("row does not exist")
		}
		logger.Error("Failed to scan todo row", "error", err)
		return entity.Todo{}, err
	}

	logger.Info("Successfully fetched todo")
	return todo, nil
}

func (r *repository) Delete(ctx context.Context, id int) error {
	logger := r.logger.With("layer", "repository", "operation", "Delete", "todo_id", id)
	logger.Debug("Attempting to delete todo")

	res, err := r.db.ExecContext(ctx, "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		logger.Error("Failed to execute delete query", "error", err)
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", "error", err)
		return err
	}

	if rowsAffected == 0 {
		logger.Warn("No todo found to delete")
		return fmt.Errorf("delete failed")
	}

	logger.Info("Successfully deleted todo")
	return nil
}

func (r *repository) Create(ctx context.Context, todo entity.Todo) (int, error) {
	logger := r.logger.With("layer", "repository", "operation", "Create")
	logger.Debug("Attempting to create todo", "title", todo.Title)

	var id int
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO todos(title, description, tags, duetime) values($1, $2, $3, $4) RETURNING id",
		todo.Title,
		todo.Description,
		pq.Array(todo.Tags),
		todo.DueTime,
	).Scan(&id)
	if err != nil {
		logger.Error("Failed to insert todo into database", "error", err)
		return 0, err
	}

	logger.Info("Successfully created todo", "todo_id", id)
	return id, nil
}

func (r *repository) Update(ctx context.Context, todo entity.Todo) error {
	logger := r.logger.With("layer", "repository", "operation", "Update")
	logger.Debug("Attempting to update todo", "todo_id", todo.ID)

	res, err := r.db.ExecContext(ctx, "UPDATE todos SET title = $1, description = $2, tags = $3, duetime = $4 "+
		"WHERE id = $5",
		todo.Title,
		todo.Description,
		pq.Array(todo.Tags),
		todo.DueTime,
		todo.ID,
	)
	if err != nil {
		logger.Error("Failed to execute update query", "error", err)
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", "error", err)
		return err
	}
	if rowsAffected == 0 {
		logger.Warn("No todo found to update")
		return fmt.Errorf("update failed")
	}

	logger.Info("Successfully updated todo")
	return nil
}

func (r *repository) GetAll(ctx context.Context, pagination entity.Pagination, filters Filters) ([]entity.Todo, error) {
	logger := r.logger.With("layer", "repository", "operation", "GetAll")
	if filters.DueTime != nil {
		logger = logger.With("due_time", *filters.DueTime)
	}
	if filters.Tags != nil {
		logger = logger.With("tags", filters.Tags)
	}
	logger.Debug("Attempting to fetching todos", "limit", pagination.Limit, "offset", pagination.Offset)

	query := "SELECT * FROM todos"
	var conditions []string
	var args []any
	argIndex := 1

	if filters.DueTime != nil {
		conditions = append(conditions, fmt.Sprintf("duetime = $%d", argIndex))
		args = append(args, *filters.DueTime)
		argIndex++
	}
	if filters.Tags != nil {
		conditions = append(conditions, fmt.Sprintf("tags && $%d", argIndex))
		args = append(args, pq.Array(filters.Tags))
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	logger.Debug("Executing query", "query", query, "args", args)

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)

	if err != nil {
		logger.Error("Failed to query todos", "error", err)
		return nil, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Error("Failed to close rows", "error", err)
		}
	}(rows)

	var all []entity.Todo
	for rows.Next() {
		var todo entity.Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, pq.Array(&todo.Tags), &todo.DueTime); err != nil {
			r.logger.Error("Failed to scan todo row", "error", err)
			return nil, err
		}
		all = append(all, todo)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("Error occurred during rows iteration", "error", err)
		return nil, err
	}

	logger.Info("Successfully fetching todos")
	return all, nil
}
