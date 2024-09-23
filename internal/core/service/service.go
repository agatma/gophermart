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
	GetUserId(accessToken string) (int, error)
}

type Order interface {
	CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error
	GetAllOrders(ctx context.Context, userID int) (domain.OrderOutList, error)
}

type Service struct {
	Authorization
	Order
}

func NewService(cfg *config.Config, storage *storage.Storage) *Service {
	return &Service{
		Authorization: newAuthService(storage, cfg),
		Order:         newOrderService(storage, cfg),
	}
}
