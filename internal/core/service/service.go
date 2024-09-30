package service

import (
	"context"
	"gophermart/internal/config"
	"gophermart/internal/core/domain"
)

type Storage interface {
	CreateUser(ctx context.Context, user *domain.UserIn) error
	GetUserID(ctx context.Context, user *domain.UserIn) (int, error)
	CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error
	UpdateOrder(ctx context.Context, order *domain.AccrualOut) error
	GetOrder(ctx context.Context, order *domain.OrderIn) (*domain.OrderOut, error)
	GetAllOrders(ctx context.Context, userID int) (domain.OrderOutList, error)
	GetAllOrdersByStatus(ctx context.Context, status string) (domain.OrderOutList, error)
	GetBalance(ctx context.Context, userID int) (*domain.BalanceOut, error)
	WithdrawBonuses(ctx context.Context, userID int, withdraw *domain.WithdrawalIn) error
	GetAllWithdrawals(ctx context.Context, userID int) (domain.WithdrawOutList, error)
}

type Authorization interface {
	CreateUser(ctx context.Context, user *domain.UserIn) error
	CreateToken(ctx context.Context, user *domain.UserIn) (string, error)
	GetUserID(accessToken string) (int, error)
}

type Order interface {
	CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error
	GetAllOrders(ctx context.Context, userID int) (domain.OrderOutList, error)
}

type Withdrawal interface {
	GetBalance(ctx context.Context, userID int) (*domain.BalanceOut, error)
	WithdrawBonuses(ctx context.Context, userID int, withdraw *domain.WithdrawalIn) error
	GetAllWithdrawals(ctx context.Context, userID int) (domain.WithdrawOutList, error)
}

type Service struct {
	Authorization
	Order
	Withdrawal
}

func NewService(cfg *config.Config, storage Storage) *Service {
	return &Service{
		Authorization: newAuthService(storage, cfg),
		Order:         newOrderService(storage, cfg),
		Withdrawal:    newWithdrawService(storage, cfg),
	}
}
