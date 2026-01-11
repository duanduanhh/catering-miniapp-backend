package repository

import (
	"context"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type CostHistoryRepository interface {
	Create(ctx context.Context, history *model.CostHistory) error
	ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.CostHistory, int64, error)
}

func NewCostHistoryRepository(
	repository *Repository,
) CostHistoryRepository {
	return &costHistoryRepository{
		Repository: repository,
	}
}

type costHistoryRepository struct {
	*Repository
}

func (r *costHistoryRepository) Create(ctx context.Context, history *model.CostHistory) error {
	return r.DB(ctx).Create(history).Error
}

func (r *costHistoryRepository) ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.CostHistory, int64, error) {
	var (
		histories []*model.CostHistory
		total     int64
	)
	db := r.DB(ctx).Model(&model.CostHistory{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (pageNum - 1) * pageSize
	if err := db.Order("create_at DESC").Offset(offset).Limit(pageSize).Find(&histories).Error; err != nil {
		return nil, 0, err
	}
	return histories, total, nil
}
