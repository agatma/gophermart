package rest

import (
	"context"
	"errors"
	"fmt"
	"gophermart/internal/core/domain"
	"net/http"
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

type Service interface {
	CreateUser(ctx context.Context, user *domain.UserIn) error
	CreateToken(ctx context.Context, user *domain.UserIn) (string, error)
	GetUserID(accessToken string) (int, error)
	CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error
	GetAllOrders(ctx context.Context, userID int) (domain.OrderOutList, error)
	GetBalance(ctx context.Context, userID int) (*domain.BalanceOut, error)
	WithdrawBonuses(ctx context.Context, userID int, withdraw *domain.WithdrawalIn) error
	GetAllWithdrawals(ctx context.Context, userID int) (domain.WithdrawOutList, error)
}

type Handler struct {
	service Service
	config  *config.Config
}

type API struct {
	srv *http.Server
}

func (a *API) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		if err := a.srv.Shutdown(ctx); err != nil {
			logger.Log.Info("gophermart shutdown: ", zap.Error(err))
		}
	}()
	if err := a.srv.ListenAndServe(); err != nil {
		logger.Log.Error("error occurred during running gophermart: ", zap.Error(err))
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed run gophermart: %w", err)
		}
	}
	return nil
}

func NewAPI(cfg *config.Config, srv Service) *API {
	h := &Handler{
		config:  cfg,
		service: srv,
	}
	r := chi.NewRouter()

	r.Use(h.loggingRequestMiddleware)
	r.Use(middleware.Timeout(serverTimeout * time.Second))
	r.Post("/api/user/register", h.SignUp)
	r.Post("/api/user/login", h.SignIn)
	r.Mount("/api/user/", ordersRouter(h))
	return &API{
		srv: &http.Server{
			Addr:    cfg.Address,
			Handler: r,
		},
	}
}

func ordersRouter(h *Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(h.authorizeRequestMiddleware)
	r.Post("/orders", h.CreateOrder)
	r.Get("/orders", h.GetAllOrders)
	r.Get("/withdrawals", h.GetAllWithdrawals)
	r.Route("/balance", func(r chi.Router) {
		r.Get("/", h.GetBalance)
		r.Post("/withdraw", h.WithdrawBonuses)
	})
	return r
}
