package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/GlebMoskalev/go-todo-api/internal/entity"
	"os"
	"time"

	"github.com/GlebMoskalev/go-todo-api/internal/config"
	"github.com/GlebMoskalev/go-todo-api/internal/database"
	"github.com/GlebMoskalev/go-todo-api/internal/todo"
	"log/slog"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

var flagConfig = flag.String("config", "./config/local.yaml", "path to the config file")

func main() {
	flag.Parse()
	cfg, err := config.Load(*flagConfig)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	logger := setupLogger(cfg.Env)
	if logger == nil {
		slog.Error("failed initialization logger", "env", cfg.Env)
		os.Exit(1)
	}

	db, err := database.InitPostgres(cfg)
	if err != nil {
		logger.Error("failed initialization database", "error", err)
		os.Exit(1)
	}

	todoRepo := todo.NewRepository(db, *logger)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := todoRepo.Create(ctx, entity.Todo{
		Title:       "test todo",
		Description: "",
		Tags:        []string{"test"},
		DueTime:     entity.Date{Time: time.Date(2020, 12, 3, 0, 0, 0, 0, time.UTC)},
	})
	fmt.Println(id, err)
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	default:
		return nil
	}
	return log
}
