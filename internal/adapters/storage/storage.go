package storage

import (
	"context"
	"errors"
	"fmt"
	"gophermart/internal/adapters/storage/postgres"
	"gophermart/internal/core/domain"
)

type Authorization interface {
	CreateUser(ctx context.Context, user *domain.UserIn) error
	GetUserId(ctx context.Context, user *domain.UserIn) (int, error)
}

type Order interface {
	CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error
	UpdateOrder(ctx context.Context, userID int, order *domain.AccrualOut) error
	GetOrder(ctx context.Context, order *domain.OrderIn) (*domain.OrderOut, error)
	GetAllOrders(ctx context.Context, userID int) (domain.OrderOutList, error)
	GetAllOrdersByStatus(ctx context.Context, status string) (domain.OrderOutList, error)
}

type Withdraw interface {
	GetBalance(ctx context.Context, userID int) (*domain.BalanceOut, error)
	Withdraw(ctx context.Context, userID int, withdraw *domain.WithdrawIn) error
	GetAllWithdraws(ctx context.Context, userID int) (domain.WithdrawOutList, error)
}

type Storage struct {
	Authorization
	Order
	Withdraw
}

func NewStorage(cfg Config) (*Storage, error) {
	if cfg.Postgres != nil {
		db, err := postgres.NewDB(cfg.Postgres)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		return &Storage{
			Authorization: postgres.NewAuthPostgres(db),
			Order:         postgres.NewOrderPostgres(db),
			Withdraw:      postgres.NewWithdrawPostgres(db),
		}, nil
	}
	return nil, errors.New("no available storage")
}
