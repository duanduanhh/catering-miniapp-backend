package model

import "time"

type ProductType int

const (
	ProductTypeTop            ProductType = 1
	ProductTypeContactVoucher ProductType = 2
	ProductTypeRefresh        ProductType = 3
)

type OrderTargetType int

const (
	OrderTargetJob OrderTargetType = 1
)

type OrderItem struct {
	ID                int64           `gorm:"primaryKey;column:id"`
	OrderID           int64           `gorm:"column:order_id"`
	ProductType       ProductType     `gorm:"column:product_type"`
	TitleSnapshot     string          `gorm:"column:title_snapshot"`
	UnitPriceSnapshot float64         `gorm:"column:unit_price_snapshot"`
	TopHour           int             `gorm:"column:top_hour"`
	ContactVoucherNum int             `gorm:"column:contact_voucher_num"`
	TargetType        OrderTargetType `gorm:"column:target_type"`
	TargetID          int64           `gorm:"column:target_id"`
	CreateAt          time.Time       `gorm:"column:create_at"`
	UpdateAt          time.Time       `gorm:"column:update_at"`
}

func (m *OrderItem) TableName() string {
	return "order_item"
}
