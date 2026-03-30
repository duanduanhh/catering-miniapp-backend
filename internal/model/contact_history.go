package model

import "time"

type ContactHistory struct {
	ID                 int64     `gorm:"primaryKey;column:id"`
	UserID             int64     `gorm:"column:user_id"`
	PurposeID          int64     `gorm:"column:purpose_id"`
	PurposeType        int       `gorm:"column:purpose_type"`
	PurposeUserID      int64     `gorm:"column:purpose_user_id"`
	PurposeUserName    string    `gorm:"column:purpose_user_name"`
	PurposeUserPhone   string    `gorm:"column:purpose_user_phone"`
	UserDeleted        int       `gorm:"column:user_deleted"`         // 发起方是否删除: 0-否, 1-是
	PurposeUserDeleted int       `gorm:"column:purpose_user_deleted"` // 被联系方是否删除: 0-否, 1-是
	CreateAt           time.Time `gorm:"column:create_at"`
	UpdateAt           time.Time `gorm:"column:update_at"`
}

func (m *ContactHistory) TableName() string {
	return "contact_history"
}
