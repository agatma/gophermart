package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart/internal/core/domain"
	"gophermart/internal/errs"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	PGUniqueViolationCode = "23505"
)

func (s *Storage) GetUserID(ctx context.Context, user *domain.UserIn) (int, error) {
	var id int
	row := s.db.QueryRow(ctx, getUserIDSQL, user.Login, user.PasswordHash)
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errs.ErrInvalidLoginOrPassword
		}
		return 0, fmt.Errorf("failed to query user: %w", err)
	}
	return id, nil
}

func (s *Storage) CreateUser(ctx context.Context, user *domain.UserIn) error {
	var (
		id     int
		pgxErr *pgconn.PgError
	)
	row := s.db.QueryRow(ctx, createUserSQL, user.Login, user.PasswordHash)
	if err := row.Scan(&id); err != nil {
		ok := errors.As(err, &pgxErr)
		if ok && pgxErr.Code == PGUniqueViolationCode {
			return errs.ErrLoginAlreadyExist
		}
		return fmt.Errorf("failed to insert user in PG: %w", err)
	}
	return nil
}
