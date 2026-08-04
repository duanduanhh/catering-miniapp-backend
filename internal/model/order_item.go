package model

import "time"

type ProductType int

const (
	ProductTypeTop            ProductType = 1
	ProductTypeContactVoucher ProductType = 2
	ProductTypeRefresh        ProductType = 3
	ProductTypePublishRent    ProductType = 4 // 发布招租（付费）
)

type OrderTargetType int

const (
	OrderTargetJob OrderTargetType = 1
)

type OrderItem struct {
	ID                       int64           `gorm:"primaryKey;column:id"`
	OrderID                  int64           `gorm:"column:order_id"`
	ProductType              ProductType     `gorm:"column:product_type"`
	ProductID                int64           `gorm:"column:product_id;index"`
	SKUID                    int64           `gorm:"column:sku_id;index"`
	SKUCode                  string          `gorm:"column:sku_code;size:64;index"`
	SKUVersion               int             `gorm:"column:sku_version"`
	VirtualProductIDSnapshot string          `gorm:"column:virtual_product_id_snapshot;size:128"`
	TitleSnapshot            string          `gorm:"column:title_snapshot"`
	UnitPriceSnapshot        float64         `gorm:"column:unit_price_snapshot"`
	PriceCentsSnapshot       int64           `gorm:"column:price_cents_snapshot"`
	BenefitSnapshot          string          `gorm:"column:benefit_snapshot;type:json"`
	TopHour                  int             `gorm:"column:top_hour"`
	ContactVoucherNum        int             `gorm:"column:contact_voucher_num"`
	TargetType               OrderTargetType `gorm:"column:target_type"`
	TargetID                 int64           `gorm:"column:target_id"`
	CreateAt                 time.Time       `gorm:"column:create_at"`
	UpdateAt                 time.Time       `gorm:"column:update_at"`
}

func (m *OrderItem) TableName() string {
	return "order_item"
}
