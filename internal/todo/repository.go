package todo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/GlebMoskalev/go-todo-api/internal/utils"
	"github.com/lib/pq"
	"log/slog"
	"strings"

	"github.com/GlebMoskalev/go-todo-api/internal/entity"
)

var (
	ErrNotFound = errors.New("todo not found")
)

type Filters struct {
	DueTime *entity.Date
	Tags    []string
}

type Repository interface {
	Get(ctx context.Context, username string, id int) (entity.Todo, error)
	Create(ctx context.Context, username string, todo entity.Todo) (int, error)
	Update(ctx context.Context, username string, todo entity.Todo) error
	Delete(ctx context.Context, username string, id int) error
	GetAll(ctx context.Context, username string, pagination entity.Pagination, filters Filters) ([]entity.Todo, int, error)
}

type repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) Repository {
	return &repository{db: db, logger: logger}
}

func (r *repository) Get(ctx context.Context, username string, id int) (entity.Todo, error) {
	logger := utils.SetupLogger(ctx, r.logger, "todo_repository", "Get", "todo_id", id)
	logger.Debug("Attempting to fetching todo")

	row := r.db.QueryRowContext(ctx,
		`SELECT t.id, t.title, t.tags, t.description, t.duetime
			 	FROM todos t JOIN users u ON t.userid = u.id WHERE t.id = $1 and u.username = $2`,
		id, username)

	var todo entity.Todo
	if err := row.Scan(&todo.ID, &todo.Title, pq.Array(&todo.Tags), &todo.Description, &todo.DueDate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("Todo not found")
			return entity.Todo{}, ErrNotFound
		}
		logger.Error("Failed to scan todo row", "error", err)
		return entity.Todo{}, err
	}

	logger.Info("Successfully fetched todo")
	return todo, nil
}

func (r *repository) Delete(ctx context.Context, username string, id int) error {
	logger := utils.SetupLogger(ctx, r.logger, "todo_repository", "Delete", "todo_id", id)
	logger.Debug("Attempting to delete todo")

	res, err := r.db.ExecContext(ctx,
		`DELETE FROM todos t USING users u WHERE t.userid = u.id AND t.id = $1 AND u.username = $2`,
		id, username)
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
		return ErrNotFound
	}

	logger.Info("Successfully deleted todo")
	return nil
}

func (r *repository) Create(ctx context.Context, username string, todo entity.Todo) (int, error) {
	logger := utils.SetupLogger(ctx, r.logger, "todo_repository", "Create")
	logger.Debug("Attempting to create todo", "title", todo.Title)

	var id int
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO todos(title, description, tags, duetime, userid) SELECT $1, $2, $3, $4, id FROM users WHERE username = $5 Returning id`,

		todo.Title,
		todo.Description,
		pq.Array(todo.Tags),
		todo.DueDate,
		username,
	).Scan(&id)
	if err != nil {
		logger.Error("Failed to insert todo into database", "error", err)
		return 0, err
	}

	logger.Info("Successfully created todo", "todo_id", id)
	return id, nil
}

func (r *repository) Update(ctx context.Context, username string, todo entity.Todo) error {
	logger := utils.SetupLogger(ctx, r.logger, "todo_repository", "Update")
	logger.Debug("Attempting to update todo", "todo_id", todo.ID)

	res, err := r.db.ExecContext(ctx,
		`UPDATE todos t SET title = $1, description = $2, tags = $3, duetime = $4 FROM users u WHERE t.userid = u.id AND t.id =$5 AND u.username =$6`,
		todo.Title,
		todo.Description,
		pq.Array(todo.Tags),
		todo.DueDate,
		todo.ID,
		username,
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
		return ErrNotFound
	}

	logger.Info("Successfully updated todo")
	return nil
}

func (r *repository) GetAll(ctx context.Context, username string, pagination entity.Pagination, filters Filters) ([]entity.Todo, int, error) {
	logger := utils.SetupLogger(ctx, r.logger, "todo_repository", "GetAll")
	if filters.DueTime != nil {
		logger = logger.With("due_time", *filters.DueTime)
	}
	if filters.Tags != nil {
		logger = logger.With("tags", filters.Tags)
	}
	logger.Debug("Attempting to fetching todos", "limit", pagination.Limit, "offset", pagination.Offset)

	var conditions []string
	var args []any
	argIndex := 1

	conditions = append(conditions, fmt.Sprintf("u.username = $%d", argIndex))
	args = append(args, username)
	argIndex++

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

	whereClause := ""
	if len(conditions) > 0 {
		whereClause += " WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := `SELECT COUNT(*) FROM todos t JOIN users u ON t.userid = u.id` + whereClause
	logger.Debug("Executing count query", "query", countQuery, "args", args)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	query := `SELECT t.id, t.title, t.description, t.tags, t.duetime FROM todos t JOIN users u ON t.userid = u.id ` + whereClause
	logger.Debug("Executing query", "query", query, "args", args)

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)

	if err != nil {
		logger.Error("Failed to query todos", "error", err)
		return nil, 0, err
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
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, pq.Array(&todo.Tags), &todo.DueDate); err != nil {
			r.logger.Error("Failed to scan todo row", "error", err)
			return nil, 0, err
		}
		all = append(all, todo)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("Error occurred during rows iteration", "error", err)
		return nil, 0, err
	}

	logger.Info("Successfully fetching todos")
	return all, total, nil
}
