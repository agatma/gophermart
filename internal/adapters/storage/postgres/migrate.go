package postgres

import (
	"embed"
	"errors"
	"fmt"
	"gophermart/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations
var migrations embed.FS

func migrate(pool *pgxpool.Pool) error {
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("postgres migrate set dialect postgres: %w", err)
	}
	db := stdlib.OpenDBFromPool(pool)
	if err := goose.Up(db, "migrations"); err != nil {
		if !errors.Is(err, goose.ErrNoNextVersion) {
			return fmt.Errorf("postgres migrate up: %w", err)
		}
	}
	if err := db.Close(); err != nil {
		return fmt.Errorf("postgres migrate close db: %w", err)
	}
	logger.Log.Info("successful migrations")
	return nil
}
