package model

import "time"

type ReportStatus int

const (
	ReportStatusPending  ReportStatus = 1 // 待处理
	ReportStatusResolved ReportStatus = 2 // 已处理
	ReportStatusRejected ReportStatus = 3 // 已驳回
)

type Report struct {
	ID          int64        `gorm:"primaryKey;column:id"`
	UserID      int64        `gorm:"column:user_id"`
	JobID       int64        `gorm:"column:job_id"`
	BizType     int          `gorm:"column:biz_type"`
	Reason      int          `gorm:"column:reason"`
	Description string       `gorm:"column:description;size:500"`
	Status      ReportStatus `gorm:"column:status"`
	CreateAt    time.Time    `gorm:"column:create_at"`
}

func (m *Report) TableName() string {
	return "report"
}
