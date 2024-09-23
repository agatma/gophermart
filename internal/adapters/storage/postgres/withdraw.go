package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart/cmd/pkg/errs"
	"gophermart/internal/core/domain"
	"gophermart/internal/logger"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type WithdrawPostgres struct {
	db *sqlx.DB
}

func NewWithdrawPostgres(db *sqlx.DB) *WithdrawPostgres {
	return &WithdrawPostgres{
		db: db,
	}
}

type Transaction interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (wp *WithdrawPostgres) getBalance(ctx context.Context, tx Transaction, userID int) (*domain.BalanceOut, error) {
	var balance domain.BalanceOut
	row := tx.QueryRowContext(
		ctx,
		`SELECT SUM(accrual) as total, SUM(withdraw) as withdraw 
               FROM ( 
            		SELECT COALESCE(accrual, 0) as accrual, 
            			   COALESCE(withdraw, 0) as withdraw 
            		FROM orders WHERE user_id=$1
                ) as t`,
		userID,
	)
	if row.Err() != nil {
		return nil, fmt.Errorf("%w", row.Err())
	}

	err := row.Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &balance, nil
}

func (wp *WithdrawPostgres) GetBalance(ctx context.Context, userID int) (*domain.BalanceOut, error) {
	return wp.getBalance(ctx, wp.db, userID)
}

func (wp *WithdrawPostgres) Withdraw(ctx context.Context, userID int, withdraw *domain.WithdrawIn) error {
	var pgxErr *pgconn.PgError
	tx, err := wp.db.Begin()
	balance, err := wp.getBalance(ctx, tx, userID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	if (balance.Current - balance.Withdrawn) < withdraw.Sum {
		return errs.ErrNotEnoughFunds
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO orders (user_id, number, status, withdraw) VALUES ($1, $2, $3, $4)",
		userID,
		withdraw.OrderNumber,
		domain.Processed,
		int64(withdraw.Sum*100),
	)
	if err != nil {
		ok := errors.As(err, &pgxErr)
		if ok && pgxErr.Code == PGUniqueViolationCode {
			return errs.ErrWithdrawAlreadyExist
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction %w", err)
	}
	return nil
}

func (wp *WithdrawPostgres) GetAllWithdraws(ctx context.Context, userID int) (domain.WithdrawOutList, error) {
	withdraws := make(domain.WithdrawOutList, 0)
	rows, err := wp.db.QueryContext(
		ctx,
		`SELECT number, withdraw, updated_at
			   FROM orders 
			   WHERE withdraw IS NOT NULL AND user_id=$1 ORDER BY updated_at`,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrOrderNotFound
		}
		return nil, fmt.Errorf("%w", err)
	}
	for rows.Next() {
		var withdraw domain.WithdrawalsOut
		err = rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		withdraws = append(withdraws, withdraw)
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			logger.Log.Error("error occurred during closing rows", zap.Error(err))
		}
	}()
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return withdraws, nil
}
