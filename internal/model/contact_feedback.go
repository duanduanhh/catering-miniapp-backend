package model

import "time"

type ContactFeedbackStatus int

const (
	ContactFeedbackStatusPending  ContactFeedbackStatus = 1 // 待处理
	ContactFeedbackStatusVerified ContactFeedbackStatus = 2 // 已核实(退券)
	ContactFeedbackStatusRejected ContactFeedbackStatus = 3 // 已驳回
)

type ContactFeedback struct {
	ID          int64                 `gorm:"primaryKey;column:id"`
	UserID      int64                 `gorm:"column:user_id"`
	JobID       int64                 `gorm:"column:job_id"`
	BizType     int                   `gorm:"column:biz_type"`
	Reason      int                   `gorm:"column:reason"`
	Description string                `gorm:"column:description;size:500"`
	Status      ContactFeedbackStatus `gorm:"column:status"`
	CreateAt    time.Time             `gorm:"column:create_at"`
}

func (m *ContactFeedback) TableName() string {
	return "contact_feedback"
}
