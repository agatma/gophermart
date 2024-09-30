package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart/internal/core/domain"
	"gophermart/internal/errs"
	"time"

	"github.com/jackc/pgx/v5"
)

const accrualFactor = 100

func (s *Storage) CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error {
	_, err := s.db.Exec(ctx, createOrderSQL, userID, order.Number, domain.New)
	if err != nil {
		return fmt.Errorf("could not create order: %w", err)
	}
	return nil
}

func (s *Storage) UpdateOrder(ctx context.Context, order *domain.AccrualOut) error {
	var err error
	if order.Accrual != nil {
		accrual := int64(*order.Accrual * accrualFactor)
		_, err = s.db.Exec(ctx, updateOrderWithAccrualSQL, order.Status, accrual, time.Now(), order.Order)
	} else {
		_, err = s.db.Exec(ctx, updateOrderSQL, order.Status, time.Now(), order.Order)
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
	row := s.db.QueryRow(ctx, getOrderSQL, order.Number)
	err := row.Scan(&orderOut.Number, &orderOut.Status, &orderOut.UserID, &accrual, &orderOut.UploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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
	rows, err := s.db.Query(ctx, getAllOrdersByUserIDSQL, userID)
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
	rows, err := s.db.Query(ctx, getAllOrdersByStatusSQL, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders in PG: %w", err)
	}
	orders, err := s.parseOrderRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to parse orders in PG: %w", err)
	}
	return orders, nil
}

func (s *Storage) parseOrderRows(rows pgx.Rows) (domain.OrderOutList, error) {
	defer func() {
		rows.Close()
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
