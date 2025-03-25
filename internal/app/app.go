package app

import (
	"database/sql"
	"github.com/GlebMoskalev/go-todo-api/internal/config"
	auth2 "github.com/GlebMoskalev/go-todo-api/internal/controller/auth"
	todo2 "github.com/GlebMoskalev/go-todo-api/internal/controller/todo"
	"github.com/GlebMoskalev/go-todo-api/internal/database"
	"github.com/GlebMoskalev/go-todo-api/internal/middleware"
	"github.com/GlebMoskalev/go-todo-api/internal/repository"
	"github.com/GlebMoskalev/go-todo-api/internal/service"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
)

const version = "v1"

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func Run(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	logger := setupLogger(cfg.Env)
	if logger == nil {
		return err
	}

	db, err := database.InitPostgres(cfg)
	if err != nil {
		logger.Error("Failed initialization database", "error", err)
		os.Exit(1)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed close database connection", "error", err)
		}
	}(db)

	router := setupRouter(logger, db, cfg)
	return http.ListenAndServe(":"+cfg.Server.Port, router)
}

func setupRouter(logger *slog.Logger, db *sql.DB, cfg config.Config) *chi.Mux {
	userRepo := repository.NewUserRepository(db, logger)
	tokenRepo := repository.NewTokenRepository(db, logger)
	todoRepo := repository.NewTodoRepository(db, logger)

	userService := service.NewUserService(userRepo, logger)
	tokenService := service.NewTokenService(userRepo, tokenRepo, cfg, logger)

	todoHandler := todo2.NewHandler(todoRepo, logger)
	authHandler := auth2.NewHandler(userService, tokenService, logger)

	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.RequestIdHeader)

	r.Route("/"+version, func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			auth2.RegisterRoutes(r, authHandler)
		})

		r.Route("/todos", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(tokenService))
				todo2.RegisterRoutes(r, todoHandler)
			})
		})
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
