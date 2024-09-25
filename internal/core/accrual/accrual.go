package accrual

import (
	"context"
	"fmt"
	"gophermart/internal/adapters/storage"
	"gophermart/internal/config"
	"gophermart/internal/core/domain"
	"gophermart/internal/logger"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const tasksCapacity = 100

type Service struct {
	storage       storage.Order
	config        *config.Config
	client        *resty.Client
	workerTimeout *WorkerTimeoutMap
}

func NewAccrualService(
	storage storage.Order,
	cfg *config.Config,
	client *resty.Client,
	workerTimeoutMap *WorkerTimeoutMap,
) *Service {
	return &Service{
		storage:       storage,
		client:        client,
		config:        cfg,
		workerTimeout: workerTimeoutMap,
	}
}

func (s *Service) getOrders(ctx context.Context, orders chan<- string) error {
	accrualPollTicker := time.NewTicker(time.Duration(s.config.AccrualPollInterval) * time.Second)
	defer accrualPollTicker.Stop()
	for range accrualPollTicker.C {
		processingOrders, err := s.storage.GetAllOrdersByStatus(ctx, domain.Processing)
		if err != nil {
			logger.Log.Error("error occurred during collecting processing orders", zap.Error(err))
			return fmt.Errorf("error occurred during collecting processing orders: %w", err)
		}
		registeredOrders, err := s.storage.GetAllOrdersByStatus(ctx, domain.Registered)
		if err != nil {
			logger.Log.Error("error occurred during collecting registered orders", zap.Error(err))
			return fmt.Errorf("error occurred during collecting registered orders: %w", err)
		}
		newOrders, err := s.storage.GetAllOrdersByStatus(ctx, domain.New)
		if err != nil {
			logger.Log.Error("error occurred during collecting new orders", zap.Error(err))
			return fmt.Errorf("error occurred during collecting new orders: %w", err)
		}
		for _, order := range processingOrders {
			orders <- order.Number
		}
		for _, order := range registeredOrders {
			orders <- order.Number
		}
		for _, order := range newOrders {
			orders <- order.Number
		}
	}
	return nil
}

func (s *Service) getOrderStatus(orderNumber string) (*domain.AccrualOut, error) {
	var order domain.AccrualOut
	resp, err := s.client.R().
		SetResult(&order).
		Get(fmt.Sprintf("%s/api/orders/%s", s.config.AccrualSystemAddress, orderNumber))

	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &order, nil
	case http.StatusTooManyRequests:
		retryAfterHeader := resp.Header().Get("Retry-After")
		if retryAfterHeader != "" {
			timeout, err := strconv.Atoi(retryAfterHeader)
			if err != nil {
				s.workerTimeout.Broadcast(s.config.AccrualTimeout)
				logger.Log.Error("too many request to accrual service", zap.Error(err))
			}
			s.workerTimeout.Broadcast(timeout)
		}
		fallthrough
	default:
		return nil, fmt.Errorf("unexpected status_code: %d; order number: %s", resp.StatusCode(), orderNumber)
	}
}

func (s *Service) updateOrderStatus(ctx context.Context, order *domain.AccrualOut) error {
	if err := s.storage.UpdateOrder(ctx, order); err != nil {
		return fmt.Errorf("error occurred during updating order: %w", err)
	}
	return nil
}

func (s *Service) processOrder(ctx context.Context, orderNumber string) error {
	order, err := s.getOrderStatus(orderNumber)
	if err != nil {
		logger.Log.Error("error during processing order", zap.Error(err))
		return fmt.Errorf("error occurred during getting order statu: %w", err)
	}
	err = s.updateOrderStatus(ctx, order)
	if err != nil {
		logger.Log.Error("error during updating order status", zap.Error(err))
		return fmt.Errorf("error occurred during updating order status: %w", err)
	}
	return nil
}

func (s *Service) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	orders := make(chan string, tasksCapacity)

	g := new(errgroup.Group)
	g.Go(func() error {
		err := s.getOrders(ctx, orders)
		if err != nil {
			cancel()
			return fmt.Errorf("error occurred during getting orders: %w", err)
		}
		return nil
	})
	for w := 1; w <= s.config.AccrualRateLimit; w++ {
		g.Go(func() error {
			err := s.worker(ctx, orders, w)
			if err != nil {
				cancel()
				return fmt.Errorf("error occurred in worker: %w", err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logger.Log.Error("error occurred", zap.Error(err))
	}
}

func (s *Service) worker(ctx context.Context, orders <-chan string, id int) error {
	for {
		select {
		case orderNumber, ok := <-orders:
			if !ok {
				return nil
			}
			if err := s.processOrder(ctx, orderNumber); err != nil {
				return fmt.Errorf("error occurred during processing order: %w", err)
			}
		case timeout := <-s.workerTimeout.GetWorker(id):
			time.Sleep(time.Duration(timeout) * time.Second)
		case <-ctx.Done():
			return nil
		}
	}
}
