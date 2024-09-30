package storage

import "gophermart/internal/adapters/storage/postgres"

type Config struct {
	Postgres *postgres.Config
}
