package main

import (
	"gophermart/internal/adapters/app"
	"gophermart/internal/config"
	"gophermart/internal/logger"
	"log"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
		return
	}
	if err = logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal(err)
		return
	}
	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatal(err)
		return
	}
	if err = application.Run(); err != nil {
		log.Fatal(err)
	}
}
