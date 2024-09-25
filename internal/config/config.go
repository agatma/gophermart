package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

const (
	tokenTTL                   = 3600
	defaultAccrualPollInterval = 5
	defaultAccrualRateLimit    = 15
	defaultAccrualTimeout      = 2
)

type Config struct {
	Address              string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AccrualPollInterval  int    `env:"ACCRUAL_POLL_INTERVAL"`
	AccrualRateLimit     int    `env:"ACCRUAL_RATE_LIMIT"`
	AccrualTimeout       int    `env:"ACCRUAL_TIMEOUT"`
	TokenKey             string `env:"FILE_STORAGE_PATH"`
	TokenTTLSeconds      int    `env:"RESTORE"`
	HashKey              string `env:"KEY"`
	LogLevel             string
}

func NewConfig() (*Config, error) {
	var cfg Config
	flag.StringVar(&cfg.Address, "a", ":8080", "port to run gophermart")
	flag.StringVar(&cfg.DatabaseURI, "d", "postgres://user:password@localhost:5434/gophermart", "postgres uri")

	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://localhost:8081", "accrual system address")
	flag.IntVar(&cfg.AccrualPollInterval, "p", defaultAccrualPollInterval, "poll interval")
	flag.IntVar(&cfg.AccrualRateLimit, "l", defaultAccrualRateLimit, "accrual rate limit")
	flag.IntVar(&cfg.AccrualTimeout, "t", defaultAccrualTimeout, " accrual timeout after 429")

	flag.StringVar(&cfg.TokenKey, "k", "<token_key>", "hashing key")
	flag.IntVar(&cfg.TokenTTLSeconds, "s", tokenTTL, "token ttl in seconds")
	flag.StringVar(&cfg.HashKey, "h", "<hash_key>", "recover data from files")
	flag.StringVar(&cfg.LogLevel, "e", "info", "log level")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		return &cfg, fmt.Errorf("failed to get config for gophermart: %w", err)
	}

	return &cfg, nil
}
