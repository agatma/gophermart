package service

import (
	"context"
	"errors"
	"fmt"
	"gophermart/internal/adapters/storage"
	"gophermart/internal/config"
	"gophermart/internal/errs"

	"gophermart/internal/core/domain"

	"github.com/ShiraazMoollatjie/goluhn"
)

type OrderService struct {
	storage storage.Order
	config  *config.Config
}

func newOrderService(storage storage.Order, config *config.Config) *OrderService {
	return &OrderService{storage: storage, config: config}
}

func (o *OrderService) CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error {
	if err := goluhn.Validate(order.Number); err != nil {
		return errs.ErrInvalidOrderNumber
	}
	orderOut, err := o.storage.GetOrder(ctx, order)
	if err != nil {
		if !errors.Is(err, errs.ErrNotFound) {
			return fmt.Errorf("failed to get order: %w", err)
		}
	}
	if orderOut != nil && orderOut.Number != "" {
		if orderOut.UserID == userID {
			return errs.ErrOrderAlreadyAdded
		}
		return errs.ErrUnreachableOrder
	}
	if err = o.storage.CreateOrder(ctx, userID, order); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

func (o *OrderService) GetAllOrders(ctx context.Context, userID int) (domain.OrderOutList, error) {
	orders, err := o.storage.GetAllOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all orders: %w", err)
	}
	return orders, nil
}
