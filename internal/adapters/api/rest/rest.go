package rest

import (
	"context"
	"fmt"
	"gophermart/internal/core/service"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gophermart/internal/config"
	"gophermart/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

const (
	serverTimeout = 3
)

type Handler struct {
	service *service.Service
	config  *config.Config
}

type API struct {
	srv *http.Server
}

func (a *API) Run() error {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigint
		if err := a.srv.Shutdown(context.Background()); err != nil {
			logger.Log.Info("gophermart shutdown: ", zap.Error(err))
		}
	}()
	if err := a.srv.ListenAndServe(); err != nil {
		logger.Log.Error("error occurred during running gophermart: ", zap.Error(err))
		return fmt.Errorf("failed run gophermart: %w", err)
	}
	return nil
}

func NewAPI(cfg *config.Config, srv *service.Service) *API {
	h := &Handler{
		config:  cfg,
		service: srv,
	}
	r := chi.NewRouter()

	r.Use(h.loggingRequestMiddleware)
	r.Use(middleware.Timeout(serverTimeout * time.Second))
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.SignUp)
		r.Post("/login", h.SignIn)

		r.Route("/", func(r chi.Router) {
			r.Use(h.authorizeRequestMiddleware)
			r.Post("/orders", h.CreateOrder)
			r.Get("/orders", h.GetAllOrders)
			r.Get("/withdrawals", h.GetAllWithdrawals)

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", h.GetBalance)
				r.Post("/withdraw", h.WithdrawBonuses)
			})
		})
	})

	return &API{
		srv: &http.Server{
			Addr:    cfg.Address,
			Handler: r,
		},
	}
}
