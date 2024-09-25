package validation

import (
	"gophermart/cmd/pkg/errs"
	"gophermart/internal/core/domain"
)

func ValidateUserIn(userIn *domain.UserIn) error {
	if userIn == nil || userIn.Password == "" || userIn.Login == "" {
		return errs.ErrValidationError
	}
	return nil
}

func ValidateWithdrawIn(withdrawIn *domain.WithdrawalIn) error {
	if withdrawIn == nil || withdrawIn.OrderNumber == "" || withdrawIn.Sum == 0.0 {
		return errs.ErrValidationError
	}
	return nil
}
