package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart/internal/core/domain"
	"gophermart/internal/errs"
	"gophermart/internal/logger"
	"time"

	"go.uber.org/zap"
)

const accrualFactor = 100

func (s *Storage) CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error {
	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3)",
		userID,
		order.Number,
		domain.New,
	)
	if err != nil {
		return fmt.Errorf("could not create order: %w", err)
	}
	return nil
}

func (s *Storage) UpdateOrder(ctx context.Context, order *domain.AccrualOut) error {
	var err error
	if order.Accrual != nil {
		accrual := int64(*order.Accrual * accrualFactor)
		_, err = s.db.ExecContext(
			ctx,
			`UPDATE orders SET status=$1, accrual=$2, updated_at=$3 WHERE number=$4`,
			order.Status,
			accrual,
			time.Now(),
			order.Order,
		)
	} else {
		_, err = s.db.ExecContext(
			ctx,
			`UPDATE orders SET status=$1, updated_at=$2 WHERE number=$3`,
			order.Status,
			time.Now(), order.Order,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to update order in PG: %w", err)
	}
	return nil
}

func (s *Storage) GetOrder(ctx context.Context, order *domain.OrderIn) (*domain.OrderOut, error) {
	var (
		orderOut domain.OrderOut
		accrual  sql.NullInt64
	)
	row := s.db.QueryRowContext(
		ctx,
		`SELECT number, status, user_id, accrual, updated_at 
			   FROM orders 
			   WHERE number=$1`,
		order.Number,
	)
	if row.Err() != nil {
		return nil, fmt.Errorf("failed to get order in PG: %w", row.Err())
	}
	err := row.Scan(&orderOut.Number, &orderOut.Status, &orderOut.UserID, &accrual, &orderOut.UploadedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get order in PG: %w", err)
	}

	if accrual.Valid {
		value := float32(accrual.Int64) / accrualFactor
		orderOut.Accrual = &value
	}
	return &orderOut, nil
}

func (s *Storage) GetAllOrders(ctx context.Context, userID int) (domain.OrderOutList, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT number, status, user_id, accrual, updated_at 
			   FROM orders 
			   WHERE user_id=$1 AND withdraw IS NULL ORDER BY updated_at`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders in PG: %w", err)
	}
	orders, err := s.parseOrderRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders in PG: %w", err)
	}
	return orders, nil
}

func (s *Storage) GetAllOrdersByStatus(ctx context.Context, status string) (domain.OrderOutList, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT number, status, user_id, accrual, updated_at 
			   FROM orders 
			   WHERE status=$1 AND withdraw IS NULL ORDER BY updated_at`,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders in PG: %w", err)
	}
	orders, err := s.parseOrderRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to parse orders in PG: %w", err)
	}
	return orders, nil
}

func (s *Storage) parseOrderRows(rows *sql.Rows) (domain.OrderOutList, error) {
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Log.Error("error occurred during closing rows", zap.Error(err))
		}
	}()
	orders := make(domain.OrderOutList, 0)
	for rows.Next() {
		var order domain.OrderOut
		var accrual sql.NullInt64
		err := rows.Scan(&order.Number, &order.Status, &order.UserID, &accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse order in PG: %w", err)
		}
		if accrual.Valid {
			value := float32(accrual.Int64) / accrualFactor
			order.Accrual = &value
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan order rows: %w", err)
	}
	return orders, nil
}
