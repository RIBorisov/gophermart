package errs

import "errors"

var (
	ErrUserExists              = errors.New("user already exists")
	ErrIncorrectPassword       = errors.New("invalid password")
	ErrUserNotExists           = errors.New("user not exists")
	ErrOrderCreatedAlready     = errors.New("order number already created by this user")
	ErrAnotherUserOrderCreated = errors.New("order number already created by another user")
	ErrGetUserFromContext      = errors.New("failed get userID from context")
	ErrInsufficientFunds       = errors.New("insufficient funds")
	ErrNoWithdrawals           = errors.New("user has no withdrawals yet")
)
