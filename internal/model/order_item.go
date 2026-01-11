package model

import "time"

type OrderItem struct {
	ID                int64     `gorm:"primaryKey;column:id"`
	OrderID           int64     `gorm:"column:order_id"`
	ProductType       int       `gorm:"column:product_type"`
	RuleID            int64     `gorm:"column:rule_id"`
	TitleSnapshot     string    `gorm:"column:title_snapshot"`
	UnitPriceSnapshot float64   `gorm:"column:unit_price_snapshot"`
	TargetType        int       `gorm:"column:target_type"`
	TargetID          int64     `gorm:"column:target_id"`
	PurposeUserID     int64     `gorm:"column:purpose_user_id"`
	CreateAt          time.Time `gorm:"column:create_at"`
	UpdateAt          time.Time `gorm:"column:update_at"`
}

func (m *OrderItem) TableName() string {
	return "order_item"
}
