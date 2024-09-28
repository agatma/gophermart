package storage

import (
	"context"
	"errors"
	"fmt"
	"gophermart/internal/adapters/storage/postgres"
	"gophermart/internal/config"
	"gophermart/internal/core/domain"
)

type Authorization interface {
	CreateUser(ctx context.Context, user *domain.UserIn) error
	GetUserID(ctx context.Context, user *domain.UserIn) (int, error)
}

type Order interface {
	CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error
	UpdateOrder(ctx context.Context, order *domain.AccrualOut) error
	GetOrder(ctx context.Context, order *domain.OrderIn) (*domain.OrderOut, error)
	GetAllOrders(ctx context.Context, userID int) (domain.OrderOutList, error)
	GetAllOrdersByStatus(ctx context.Context, status string) (domain.OrderOutList, error)
}

type Withdrawal interface {
	GetBalance(ctx context.Context, userID int) (*domain.BalanceOut, error)
	WithdrawBonuses(ctx context.Context, userID int, withdraw *domain.WithdrawalIn) error
	GetAllWithdrawals(ctx context.Context, userID int) (domain.WithdrawOutList, error)
}

type Storage struct {
	Authorization
	Order
	Withdrawal
}

func NewStorage(cfg *config.Config) (*Storage, error) {
	if cfg.DatabaseURI == "" {
		return nil, errors.New("postgres uri is required")
	}
	db, err := postgres.NewDB(&postgres.Config{
		DSN: cfg.DatabaseURI,
	})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &Storage{
		Authorization: postgres.NewAuthPostgres(db),
		Order:         postgres.NewOrderPostgres(db),
		Withdrawal:    postgres.NewWithdrawPostgres(db),
	}, nil
}
