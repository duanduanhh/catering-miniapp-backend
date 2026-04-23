package model

import "time"

type FeedbackType int

const (
	FeedbackTypeProductSuggestion FeedbackType = 1 // 产品建议
	FeedbackTypeFunctionIssue     FeedbackType = 2 // 功能问题
	FeedbackTypeContentCorrection FeedbackType = 3 // 内容修正
	FeedbackTypeOther             FeedbackType = 4 // 其他
)

type Feedback struct {
	ID        int64        `gorm:"primaryKey;column:id"`
	UserID    int64        `gorm:"column:user_id"`
	Type      FeedbackType `gorm:"column:type"`
	Content   string       `gorm:"column:content"`
	PhotoURLs string       `gorm:"column:photo_urls"`
	CreateAt  time.Time    `gorm:"column:create_at"`
}

func (m *Feedback) TableName() string {
	return "feedback"
}
