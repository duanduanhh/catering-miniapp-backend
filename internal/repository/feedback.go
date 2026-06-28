package repository

import (
	"context"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type FeedbackRepository interface {
	Create(ctx context.Context, feedback *model.Feedback) error
	ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.Feedback, int64, error)
	AdminList(ctx context.Context, query AdminFeedbackListQuery) ([]*AdminFeedbackRow, int64, error)
}

// AdminFeedbackListQuery 管理后台意见反馈列表筛选条件
type AdminFeedbackListQuery struct {
	FeedbackID int64
	Type       *int
	UserID     int64
	StartTime  string
	EndTime    string
	PageNum    int
	PageSize   int
}

// AdminFeedbackRow 反馈 + JOIN 出的用户姓名手机号
type AdminFeedbackRow struct {
	model.Feedback
	UserName  string `gorm:"column:user_name"`
	UserPhone string `gorm:"column:user_phone"`
}

func NewFeedbackRepository(
	repository *Repository,
) FeedbackRepository {
	return &feedbackRepository{
		Repository: repository,
	}
}

type feedbackRepository struct {
	*Repository
}

func (r *feedbackRepository) Create(ctx context.Context, feedback *model.Feedback) error {
	return r.DB(ctx).Create(feedback).Error
}

func (r *feedbackRepository) ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.Feedback, int64, error) {
	var (
		list  []*model.Feedback
		total int64
	)
	db := r.DB(ctx).Model(&model.Feedback{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize
	if err := db.Order("create_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AdminList 管理后台意见反馈列表，JOIN user 拿姓名/手机号。
func (r *feedbackRepository) AdminList(ctx context.Context, query AdminFeedbackListQuery) ([]*AdminFeedbackRow, int64, error) {
	var (
		rows  []*AdminFeedbackRow
		total int64
	)
	db := r.DB(ctx).
		Table("feedback AS f").
		Select("f.*, u.name AS user_name, u.phone AS user_phone").
		Joins("LEFT JOIN user u ON f.user_id = u.id")

	if query.FeedbackID != 0 {
		db = db.Where("f.id = ?", query.FeedbackID)
	}
	if query.Type != nil {
		db = db.Where("f.type = ?", *query.Type)
	}
	if query.UserID > 0 {
		db = db.Where("f.user_id = ?", query.UserID)
	}
	if query.StartTime != "" {
		db = db.Where("f.create_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("f.create_at <= ?", query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	offset := (query.PageNum - 1) * query.PageSize
	if err := db.Order("f.create_at DESC").Offset(offset).Limit(query.PageSize).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
