package model

import "time"

type ContactFeedbackStatus int

const (
	ContactFeedbackStatusPending  ContactFeedbackStatus = 1 // 待处理
	ContactFeedbackStatusVerified ContactFeedbackStatus = 2 // 已核实(退券)
	ContactFeedbackStatusRejected ContactFeedbackStatus = 3 // 已驳回
)

type ContactFeedback struct {
	ID               int64                 `gorm:"primaryKey;column:id"`
	UserID           int64                 `gorm:"column:user_id"`
	ContactHistoryID int64                 `gorm:"column:contact_history_id;default:0;index"` // 关联 contact_history.id；提交时按 (user,job,biz_type) 反查匹配的最新一条；0 表示历史数据无法关联
	JobID            int64                 `gorm:"column:job_id"`
	BizType          int                   `gorm:"column:biz_type"`
	Reason           int                   `gorm:"column:reason"`
	Description      string                `gorm:"column:description;size:500"`
	Status           ContactFeedbackStatus `gorm:"column:status"`
	CreateAt         time.Time             `gorm:"column:create_at"`
}

func (m *ContactFeedback) TableName() string {
	return "contact_feedback"
}
