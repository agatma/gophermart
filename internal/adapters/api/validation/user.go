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
