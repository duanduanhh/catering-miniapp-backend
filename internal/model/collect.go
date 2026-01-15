package model

import "time"

type CollectType int

const (
	CollectTypeJob  CollectType = 1
	CollectTypeResume CollectType = 2
	CollectTypeRent CollectType = 3
)

type CollectStatus int

const (
	CollectStatusActive  CollectStatus = 1
	CollectStatusDeleted CollectStatus = 2
)

type Collect struct {
	UserID    int64     `gorm:"primaryKey;column:user_id"`
	ContentID int64     `gorm:"primaryKey;column:content_id"`
	Type      CollectType `gorm:"primaryKey;column:type"`
	Status    CollectStatus `gorm:"column:status"`
	CreateAt  time.Time `gorm:"column:create_at"`
	UpdateAt  time.Time `gorm:"column:update_at"`
}

func (m *Collect) TableName() string {
	return "collect"
}
