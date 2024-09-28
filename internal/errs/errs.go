package errs

import "errors"

var (
	ErrValidationError = errors.New("validation error")
	ErrNotFound        = errors.New("element not found")

	ErrLoginAlreadyExist      = errors.New("login already exist")
	ErrInvalidLoginOrPassword = errors.New("invalid login or password")

	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrOrderAlreadyAdded  = errors.New("order has already been added")
	ErrUnreachableOrder   = errors.New("order has already been added by another user")

	ErrWithdrawAlreadyExist = errors.New("withdraw for this order already exist")
	ErrNotEnoughFunds       = errors.New("not enough bonuses to withdraw")
)
