package service

import "errors"

var (
	ErrForbidden           = errors.New("forbidden")
	ErrInsufficientVoucher = errors.New("insufficient contact voucher")
	ErrAmountMismatch      = errors.New("amount mismatch")
	ErrInvalidVoucherNum   = errors.New("invalid voucher number")
	ErrUserExists          = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrJobLimitExceeded             = errors.New("job limit exceeded")
	ErrShareRefreshLimitExceeded    = errors.New("share refresh limit exceeded")

	ErrEnterpriseNotFound  = errors.New("enterprise not found")
	ErrEnterpriseDuplicate = errors.New("enterprise already exists")
	ErrInvalidCreditCode   = errors.New("invalid social credit code")
)
