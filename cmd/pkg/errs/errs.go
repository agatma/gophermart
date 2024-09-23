package errs

import "errors"

var (
	ErrLoginAlreadyExist      = errors.New("login already exist")
	ErrInvalidLoginOrPassword = errors.New("invalid login or password")

	ErrOrderAlreadyExist  = errors.New("order already exist")
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrOrderAlreadyAdded  = errors.New("order has already been added")
	ErrUnreachableOrder   = errors.New("order has already been added by another user")

	ErrValidationError = errors.New("validation error")

	ErrWithdrawAlreadyExist = errors.New("withdraw for this order already exist")
	ErrNotEnoughFunds       = errors.New("not enough bonuses to withdraw")
)
