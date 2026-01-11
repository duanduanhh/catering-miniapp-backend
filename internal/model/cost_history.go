package model

import "time"

type CostHistory struct {
	ID        int64     `gorm:"primaryKey;column:id"`
	UserID    int64     `gorm:"column:user_id"`
	BizType   int       `gorm:"column:biz_type"`
	ChangeNum int       `gorm:"column:change_num"`
	LastNum   int       `gorm:"column:last_num"`
	NextNum   int       `gorm:"column:next_num"`
	Remark    string    `gorm:"column:remark"`
	CreateAt  time.Time `gorm:"column:create_at"`
}

func (m *CostHistory) TableName() string {
	return "cost_history"
}
