package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

const (
	tokenTTL = 3600
)

type Config struct {
	Address              string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	TokenKey             string `env:"FILE_STORAGE_PATH"`
	TokenTTLSeconds      int    `env:"RESTORE"`
	HashKey              string `env:"KEY"`
	LogLevel             string
}

func NewConfig() (*Config, error) {
	var cfg Config
	flag.StringVar(&cfg.Address, "a", ":8080", "port to run server")
	flag.StringVar(&cfg.DatabaseURI, "d", "postgres://user:password@localhost:5434/gophermart", "postgres uri")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", ":8081", "accrual system address")
	flag.StringVar(&cfg.TokenKey, "t", "<token_key>", "hashing key")
	flag.IntVar(&cfg.TokenTTLSeconds, "s", tokenTTL, "token ttl in seconds")
	flag.StringVar(&cfg.HashKey, "k", "<hash_key>", "recover data from files")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		return &cfg, fmt.Errorf("failed to get config for server: %w", err)
	}
	return &cfg, nil
}
