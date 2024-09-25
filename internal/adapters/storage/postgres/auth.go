package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart/internal/core/domain"

	"gophermart/cmd/pkg/errs"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

const (
	PGUniqueViolationCode = "23505"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{
		db: db,
	}
}

func (ap *AuthPostgres) GetUserID(ctx context.Context, user *domain.UserIn) (int, error) {
	var id int
	row := ap.db.QueryRowContext(
		ctx,
		"SELECT id FROM users WHERE login=$1 AND password_hash=$2",
		user.Login,
		user.Password,
	)
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errs.ErrInvalidLoginOrPassword
		}
		return 0, fmt.Errorf("failed to query user: %w", err)
	}
	return id, nil
}

func (ap *AuthPostgres) CreateUser(ctx context.Context, user *domain.UserIn) error {
	var id int
	var pgxErr *pgconn.PgError
	row := ap.db.QueryRowContext(
		ctx,
		"INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id",
		user.Login,
		user.Password,
	)
	if err := row.Scan(&id); err != nil {
		ok := errors.As(err, &pgxErr)
		if ok && pgxErr.Code == PGUniqueViolationCode {
			return errs.ErrLoginAlreadyExist
		}
		return fmt.Errorf("failed to insert user in PG: %w", err)
	}
	return nil
}
