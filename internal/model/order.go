package model

import "time"

type Order struct {
	ID          int64      `gorm:"primaryKey;column:id"`
	OrderNo     string     `gorm:"column:order_no"`
	UserID      int64      `gorm:"column:user_id"`
	AmountTotal Decimal    `gorm:"column:amount_total;type:decimal(10,2)"`
	AmountPaid  Decimal    `gorm:"column:amount_paid;type:decimal(10,2)"`
	Currency    string     `gorm:"column:currency"`
	Status      int        `gorm:"column:status"`
	PayChannel  string     `gorm:"column:pay_channel"`
	PayTradeNo  string     `gorm:"column:pay_trade_no"`
	PaidAt      *time.Time `gorm:"column:paid_at"`
	CanceledAt  *time.Time `gorm:"column:canceled_at"`
	RefundedAt  *time.Time `gorm:"column:refunded_at"`
	Remark      string     `gorm:"column:remark"`
	CreateAt    time.Time  `gorm:"column:create_at"`
	UpdateAt    time.Time  `gorm:"column:update_at"`
}

func (m *Order) TableName() string {
	return "order"
}
