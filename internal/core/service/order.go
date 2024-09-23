package service

import (
	"context"
	"errors"
	"gophermart/cmd/pkg/errs"
	"gophermart/internal/config"

	"gophermart/internal/core/domain"
)

type OrderStorage interface {
	CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error
	UpdateOrder(ctx context.Context, userID int, order *domain.AccrualOut) error
	GetOrder(ctx context.Context, order *domain.OrderIn) (*domain.OrderOut, error)
	GetAllOrders(ctx context.Context, userID int) (domain.OrderOutList, error)
	GetAllOrdersByStatus(ctx context.Context, status string) (domain.OrderOutList, error)
}

type OrderService struct {
	storage OrderStorage
	config  *config.Config
}

func newOrderService(storage OrderStorage, config *config.Config) *OrderService {
	return &OrderService{storage: storage, config: config}
}

func (o *OrderService) CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error {
	var err error
	//if err = goluhn.Validate(order.Number); err != nil {
	//	return errs.ErrInvalidOrderNumber
	//}
	orderOut, err := o.storage.GetOrder(ctx, order)
	if err != nil {
		if !errors.Is(err, errs.ErrOrderNotFound) {
			return err
		}
	}
	if orderOut != nil && orderOut.Number != "" {
		if orderOut.UserID == userID {
			return errs.ErrOrderAlreadyAdded
		} else if orderOut.UserID != userID {
			return errs.ErrUnreachableOrder
		}
	}
	err = o.storage.CreateOrder(ctx, userID, order)
	return err
}

func (o *OrderService) GetAllOrders(ctx context.Context, userID int) (domain.OrderOutList, error) {
	return o.storage.GetAllOrders(ctx, userID)
}
