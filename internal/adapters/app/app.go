package app

import (
	"context"
	"fmt"
	"gophermart/internal/adapters/api/rest"
	"gophermart/internal/adapters/storage"
	"gophermart/internal/config"
	"gophermart/internal/core/accrual"
	"gophermart/internal/core/service"
	"gophermart/internal/logger"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type App struct {
	accrual *accrual.Service
	api     *rest.API
}

func NewApp(cfg *config.Config) (*App, error) {
	activeStorage, err := storage.NewStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a storage: %w", err)
	}
	newService := service.NewService(cfg, activeStorage)
	accrualService := accrual.NewAccrualService(
		activeStorage,
		cfg,
		resty.New(),
		accrual.NewWorkerTimeoutMap(cfg.AccrualRateLimit),
	)
	api := rest.NewAPI(cfg, newService)
	return &App{
		accrual: accrualService,
		api:     api,
	}, nil
}

func (a *App) Run() error {
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		err := a.accrual.Run(ctx)
		if err != nil {
			logger.Log.Error("accrual run failed:", zap.Error(err))
			return fmt.Errorf("accrual run failed: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := a.api.Run(ctx)
		if err != nil {
			logger.Log.Error("api run failed:", zap.Error(err))
			return fmt.Errorf("api run failed: %w", err)
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		logger.Log.Error("app run failed:", zap.Error(err))
		return fmt.Errorf("app run failed: %w", err)
	}
	return nil

}
