package main

import (
	"errors"
	"fmt"
	"gophermart/internal/adapters/api/rest"
	"gophermart/internal/adapters/storage"
	"gophermart/internal/adapters/storage/postgres"
	"gophermart/internal/config"
	"gophermart/internal/core/service"
	"gophermart/internal/logger"
	"log"
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
	newStorage, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize a storage: %w", err)
	}
	newService := service.NewService(cfg, newStorage)
	api := rest.NewAPI(cfg, newService)
	if err = api.Run(); err != nil {
		return fmt.Errorf("server has failed: %w", err)
	}
	return nil
}

func initStorage(cfg *config.Config) (*storage.Storage, error) {
	if cfg.DatabaseURI == "" {

		return nil, errors.New("postgres uri is required")
	}
	postgresStorage, err := storage.NewStorage(storage.Config{
		Postgres: &postgres.Config{
			DSN: cfg.DatabaseURI,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init db storage %w", err)
	}
	logger.Log.Info("db storage is initialized")
	return postgresStorage, nil
}
