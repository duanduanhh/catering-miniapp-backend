package model

import "time"

type Collect struct {
	UserID    int64     `gorm:"primaryKey;column:user_id"`
	ContentID int64     `gorm:"primaryKey;column:content_id"`
	Type      int       `gorm:"primaryKey;column:type"`
	Status    int       `gorm:"column:status"`
	CreateAt  time.Time `gorm:"column:create_at"`
	UpdateAt  time.Time `gorm:"column:update_at"`
}

func (m *Collect) TableName() string {
	return "collect"
}
