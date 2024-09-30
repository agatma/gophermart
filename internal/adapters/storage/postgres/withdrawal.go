package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart/internal/core/domain"
	"gophermart/internal/errs"
	"gophermart/internal/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type Transaction interface {
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
}

func (s *Storage) getBalance(ctx context.Context, tx Transaction, userID int) (*domain.BalanceOut, error) {
	var balance domain.BalanceOut
	row := tx.QueryRow(ctx, getBalanceSQL, userID)
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

func (s *Storage) GetBalance(ctx context.Context, userID int) (*domain.BalanceOut, error) {
	balance, err := s.getBalance(ctx, s.db, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance in PG: %w", err)
	}
	return balance, nil
}

func (s *Storage) WithdrawBonuses(ctx context.Context, userID int, withdraw *domain.WithdrawalIn) error {
	var pgxErr *pgconn.PgError
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	balance, err := s.getBalance(ctx, tx, userID)
	if err != nil {
		return fmt.Errorf("failed to get balance for withdraw bonuses: %w", err)
	}
	if balance.Current < withdraw.Sum {
		return errs.ErrNotEnoughFunds
	}
	_, err = tx.Exec(
		ctx,
		withdrawBonusesSQL,
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
		if txErr := tx.Rollback(ctx); txErr != nil {
			if !errors.Is(txErr, sql.ErrTxDone) {
				logger.Log.Error("failed to rollback the transaction", zap.Error(txErr))
			}
		}
		return fmt.Errorf("failed to withdraw bonuses: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction %w", err)
	}
	return nil
}

func (s *Storage) GetAllWithdrawals(ctx context.Context, userID int) (domain.WithdrawOutList, error) {
	withdrawals := make(domain.WithdrawOutList, 0)
	rows, err := s.db.Query(ctx, getAllWithdrawalsSQL, userID)
	defer func() {
		rows.Close()
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
