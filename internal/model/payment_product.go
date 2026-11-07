package model

import "time"

// PaymentProductCode 是支付产品的稳定标识。它同时确定 SKU 的权益类型，避免再维护一套数字类型。
const (
	PaymentProductCodeJobTop         = "job_top"
	PaymentProductCodeContactVoucher = "contact_voucher"
	PaymentProductCodePaidRefresh    = "paid_refresh"
	PaymentProductCodeRentPublish    = "rent_publish"
)

type PaymentSelectionMode int

const (
	PaymentSelectionModeSingle   PaymentSelectionMode = 1
	PaymentSelectionModeMultiple PaymentSelectionMode = 2
)

type PaymentProduct struct {
	ID             int64                `gorm:"primaryKey;column:id"`
	ProductCode    string               `gorm:"column:product_code;size:64;not null;uniqueIndex:uk_payment_product_code"`
	Name           string               `gorm:"column:name;size:100;not null"`
	SelectionMode  PaymentSelectionMode `gorm:"column:selection_mode;not null;check:chk_payment_product_selection_mode,selection_mode IN (1,2)"`
	PurchaseNotice string               `gorm:"column:purchase_notice;type:text"`
	CreateAt       time.Time            `gorm:"column:create_at;not null"`
	UpdateAt       time.Time            `gorm:"column:update_at;not null"`
}

func (m *PaymentProduct) TableName() string {
	return "payment_product"
}
