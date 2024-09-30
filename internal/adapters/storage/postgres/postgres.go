package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	PGConnectionTimeout = 10
)

type Storage struct {
	db *pgxpool.Pool
}

func NewPostgresStorage(cfg *Config) (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(PGConnectionTimeout)*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres %w", err)
	}
	err = migrate(pool)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate %w", err)
	}
	return &Storage{db: pool}, nil
}
