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

type OrderPostgres struct {
	db *sqlx.DB
}

func NewOrderPostgres(db *sqlx.DB) *OrderPostgres {
	return &OrderPostgres{
		db: db,
	}
}

func (op *OrderPostgres) CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error {
	var pgxErr *pgconn.PgError
	_, err := op.db.ExecContext(
		ctx,
		"INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3)",
		userID,
		order.Number,
		domain.New,
	)
	if err != nil {
		ok := errors.As(err, &pgxErr)
		if ok && pgxErr.Code == PGUniqueViolationCode {
			return errs.ErrOrderAlreadyExist
		}
		return err
	}
	return nil
}

func (op *OrderPostgres) GetOrder(ctx context.Context, order *domain.OrderIn) (*domain.OrderOut, error) {
	var (
		orderOut domain.OrderOut
		accrual  sql.NullInt64
	)
	row := op.db.QueryRowContext(
		ctx,
		`SELECT number, status, user_id, accrual, updated_at 
			   FROM orders 
			   WHERE number=$1`,
		order.Number,
	)
	if row.Err() != nil {
		return nil, fmt.Errorf("%w", row.Err())
	}

	err := row.Scan(&orderOut.Number, &orderOut.Status, &orderOut.UserID, &accrual, &orderOut.UploadedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrOrderNotFound
		}
		return nil, fmt.Errorf("%w", err)
	}

	if accrual.Valid {
		value := float32(accrual.Int64) / 100
		orderOut.Accrual = &value
	}
	return &orderOut, nil
}

func (op *OrderPostgres) GetAllOrders(ctx context.Context, userID int) (domain.OrderOutList, error) {
	orders := make(domain.OrderOutList, 0)
	rows, err := op.db.QueryContext(
		ctx,
		`SELECT number, status, user_id, accrual, updated_at 
			   FROM orders 
			   WHERE user_id=$1 AND withdraw IS NULL ORDER BY updated_at`,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrOrderNotFound
		}
		return nil, fmt.Errorf("%w", err)
	}
	for rows.Next() {
		var order domain.OrderOut
		var accrual sql.NullInt64
		err = rows.Scan(&order.Number, &order.Status, &order.UserID, &accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		if accrual.Valid {
			value := float32(accrual.Int64) / 100
			order.Accrual = &value
		}
		orders = append(orders, order)
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
	return orders, nil
}

func (op *OrderPostgres) GetAllOrdersByStatus(ctx context.Context, status string) (domain.OrderOutList, error) {
	orders := make(domain.OrderOutList, 0)
	return orders, nil
}

func (op *OrderPostgres) UpdateOrder(ctx context.Context, userID int, order *domain.AccrualOut) error {
	return nil
}
