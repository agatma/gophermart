package service

import (
	"context"
	"fmt"
	"gophermart/cmd/pkg/errs"
	"gophermart/internal/adapters/storage"
	"gophermart/internal/config"
	"gophermart/internal/core/domain"

	"github.com/ShiraazMoollatjie/goluhn"
)

type WithdrawService struct {
	storage storage.Withdrawal
	config  *config.Config
}

func newWithdrawService(storage storage.Withdrawal, config *config.Config) *WithdrawService {
	return &WithdrawService{storage: storage, config: config}
}

func (ws *WithdrawService) GetBalance(ctx context.Context, userID int) (*domain.BalanceOut, error) {
	balance, err := ws.storage.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance for user %d: %w", userID, err)
	}
	return balance, nil
}

func (ws *WithdrawService) WithdrawBonuses(ctx context.Context, userID int, withdraw *domain.WithdrawalIn) error {
	if err := goluhn.Validate(withdraw.OrderNumber); err != nil {
		return errs.ErrInvalidOrderNumber
	}
	if err := ws.storage.WithdrawBonuses(ctx, userID, withdraw); err != nil {
		return fmt.Errorf("failed to withdraw bonuses for user %d: %w", userID, err)
	}
	return nil
}

func (ws *WithdrawService) GetAllWithdrawals(ctx context.Context, userID int) (domain.WithdrawOutList, error) {
	withdrawals, err := ws.storage.GetAllWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all withdrawals for user %d: %w", userID, err)
	}
	return withdrawals, nil
}
