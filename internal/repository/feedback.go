package repository

import (
	"context"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type FeedbackRepository interface {
	Create(ctx context.Context, feedback *model.Feedback) error
	ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.Feedback, int64, error)
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
