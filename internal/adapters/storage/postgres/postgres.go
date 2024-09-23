package postgres

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewDB(cfg *Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres %w", err)
	}
	err = migrate(db)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate %w", err)
	}
	return db, nil
}
