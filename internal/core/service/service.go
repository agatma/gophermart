package service

import (
	"context"
	"gophermart/internal/adapters/storage"
	"gophermart/internal/config"
	"gophermart/internal/core/domain"
)

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

func NewService(cfg *config.Config, storage *storage.Storage) *Service {
	return &Service{
		Authorization: newAuthService(storage, cfg),
		Order:         newOrderService(storage, cfg),
		Withdrawal:    newWithdrawService(storage, cfg),
	}
}
