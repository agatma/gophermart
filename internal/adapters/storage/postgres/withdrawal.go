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

type WithdrawalPostgres struct {
	db *sqlx.DB
}

func NewWithdrawPostgres(db *sqlx.DB) *WithdrawalPostgres {
	return &WithdrawalPostgres{
		db: db,
	}
}

type Transaction interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (wp *WithdrawalPostgres) getBalance(ctx context.Context, tx Transaction, userID int) (*domain.BalanceOut, error) {
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
		return nil, fmt.Errorf("failed to get balance in PG: %w", row.Err())
	}

	if err := row.Scan(&balance.Current, &balance.Withdrawn); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return &domain.BalanceOut{Current: 0, Withdrawn: 0}, nil
		}
		return nil, fmt.Errorf("failed to get balance in PG: %w", err)
	}

	balance.Current = float32(balance.Current-balance.Withdrawn) / accrualFactor
	balance.Withdrawn = float32(balance.Withdrawn) / accrualFactor
	return &balance, nil
}

func (wp *WithdrawalPostgres) GetBalance(ctx context.Context, userID int) (*domain.BalanceOut, error) {
	balance, err := wp.getBalance(ctx, wp.db, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance in PG: %w", err)
	}
	return balance, nil
}

func (wp *WithdrawalPostgres) WithdrawBonuses(ctx context.Context, userID int, withdraw *domain.WithdrawalIn) error {
	var pgxErr *pgconn.PgError
	tx, err := wp.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	balance, err := wp.getBalance(ctx, tx, userID)
	if err != nil {
		return fmt.Errorf("failed to get balance for withdraw bonuses: %w", err)
	}
	if balance.Current < withdraw.Sum {
		return errs.ErrNotEnoughFunds
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO orders (user_id, number, status, withdraw) VALUES ($1, $2, $3, $4)",
		userID,
		withdraw.OrderNumber,
		domain.Processed,
		int64(withdraw.Sum*accrualFactor),
	)
	if err != nil {
		ok := errors.As(err, &pgxErr)
		if ok && pgxErr.Code == PGUniqueViolationCode {
			return errs.ErrWithdrawAlreadyExist
		}
		if txErr := tx.Rollback(); txErr != nil {
			if !errors.Is(txErr, sql.ErrTxDone) {
				logger.Log.Error("failed to rollback the transaction", zap.Error(txErr))
			}
		}
		return fmt.Errorf("failed to withdraw bonuses: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction %w", err)
	}
	return nil
}

func (wp *WithdrawalPostgres) GetAllWithdrawals(ctx context.Context, userID int) (domain.WithdrawOutList, error) {
	withdrawals := make(domain.WithdrawOutList, 0)
	rows, err := wp.db.QueryContext(
		ctx,
		`SELECT number, withdraw, updated_at
			   FROM orders 
			   WHERE withdraw IS NOT NULL AND user_id=$1 ORDER BY updated_at`,
		userID,
	)
	defer func() {
		err = rows.Close()
		if err != nil {
			logger.Log.Error("error occurred during closing rows", zap.Error(err))
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}
	for rows.Next() {
		var withdraw domain.WithdrawalsOut
		err = rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to get all withdrawals%w", err)
		}
		withdraw.Sum = float32(withdraw.Sum) / accrualFactor
		withdrawals = append(withdrawals, withdraw)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}
	return withdrawals, nil
}
