package main

import (
	"context"
	"fmt"
	"gophermart/internal/adapters/api/rest"
	"gophermart/internal/adapters/storage"
	"gophermart/internal/config"
	"gophermart/internal/core/accrual"
	"gophermart/internal/core/service"
	"gophermart/internal/logger"
	"log"

	"github.com/go-resty/resty/v2"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("can't load config: %w", err)
	}
	if err = logger.Initialize(cfg.LogLevel); err != nil {
		return fmt.Errorf("can't load logger: %w", err)
	}
	activeStorage, err := storage.InitStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize a storage: %w", err)
	}
	newService := service.NewService(cfg, activeStorage)
	accrualService := accrual.NewAccrualService(
		activeStorage,
		cfg,
		resty.New(),
		accrual.NewWorkerTimeoutMap(cfg.AccrualRateLimit),
	)
	go accrualService.Run(context.Background())
	api := rest.NewAPI(cfg, newService)
	if err = api.Run(); err != nil {
		return fmt.Errorf("gophermart has failed: %w", err)
	}
	return nil
}
