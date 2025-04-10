package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"paymentSystem/internal/config"
)

type Storage struct {
	Postgres *sql.DB
}

func NewStorage(cfg *config.Config) (*Storage, error) {
	const op = "storage.NewStorage"
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{
		Postgres: db,
	}, nil
}
