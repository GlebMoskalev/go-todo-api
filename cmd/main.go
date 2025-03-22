package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/GlebMoskalev/go-todo-api/internal/config"
	"github.com/GlebMoskalev/go-todo-api/internal/database"
	"github.com/GlebMoskalev/go-todo-api/internal/middleware"
	"github.com/GlebMoskalev/go-todo-api/internal/todo"
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

	router := setupRouter(logger, db, cfg)

	http.ListenAndServe(":8888", router)
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
}

func setupRouter(logger *slog.Logger, db *sql.DB, cfg config.Config) *chi.Mux {
	todoRepo := todo.NewRepository(db, *logger)
	todoHandler := todo.NewHandler(todoRepo, logger)

	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.RequestIdHeader)
	r.Use(middleware.RequestIdLogger(logger))

	r.Route("/todos", func(r chi.Router) {
		todo.RegisterRoutes(r, todoHandler)
	})

	logger.Info("Starting server", "port", cfg.Database.Port)
	return r
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
