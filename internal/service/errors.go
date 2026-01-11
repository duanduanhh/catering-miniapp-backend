package service

import "errors"

var (
	ErrForbidden          = errors.New("forbidden")
	ErrInsufficientVoucher = errors.New("insufficient contact voucher")
	ErrAmountMismatch     = errors.New("amount mismatch")
	ErrInvalidVoucherNum  = errors.New("invalid voucher number")
)
