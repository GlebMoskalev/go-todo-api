package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/GlebMoskalev/go-todo-api/internal/config"
	_ "github.com/lib/pq"
)

func InitPostgres(cfg config.Config) (*sql.DB, error) {
	dbUrl := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Host,
		cfg.Database.Port,
	)

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("error openning database %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, errors.New("error pinging database")
	}

	return db, nil
}
