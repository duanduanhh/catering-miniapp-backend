package service

import "errors"

var (
	ErrForbidden                  = errors.New("forbidden")
	ErrInsufficientVoucher        = errors.New("insufficient contact voucher")
	ErrAmountMismatch             = errors.New("amount mismatch")
	ErrInvalidVoucherNum          = errors.New("invalid voucher number")
	ErrUserExists                 = errors.New("user already exists")
	ErrUserNotFound               = errors.New("user not found")
	ErrJobLimitExceeded           = errors.New("job limit exceeded")
	ErrShareRefreshLimitExceeded  = errors.New("share refresh limit exceeded")
	ErrImageRiskyContent          = errors.New("image risky content")
	ErrInvalidImageCheckMode      = errors.New("invalid image check mode")
	ErrPaymentPackageNotFound     = errors.New("payment package not found")
	ErrPaymentPackageInvalid      = errors.New("invalid payment package")
	ErrPaymentPackageConflict     = errors.New("payment package was modified, please refresh and retry")
	ErrPaymentPackageSKUExists    = errors.New("payment package sku already exists")
	ErrPaymentPackagePublished    = errors.New("published payment package cannot be edited or deleted")
	ErrPaymentProductNotFound     = errors.New("payment product not found")
	ErrPaymentPackageCardinality  = errors.New("single-selection product can only have one package")
	ErrPaymentPackageSingleDelete = errors.New("single-selection product package cannot be deleted")
	ErrPaymentPackageUnavailable  = errors.New("payment package is not available for this user")
	ErrPaymentPackageLimitReached = errors.New("payment package purchase limit reached")

	ErrEnterpriseNotFound  = errors.New("enterprise not found")
	ErrEnterpriseDuplicate = errors.New("enterprise already exists")
	ErrInvalidCreditCode   = errors.New("invalid social credit code")
)
