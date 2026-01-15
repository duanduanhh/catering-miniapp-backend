package repository

import (
	"context"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type ContactVoucherHistoryRepository interface {
	Create(ctx context.Context, history *model.ContactVoucherHistory) error
	ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.ContactVoucherHistory, int64, error)
}

func NewContactVoucherHistoryRepository(
	repository *Repository,
) ContactVoucherHistoryRepository {
	return &contactVoucherHistoryRepository{
		Repository: repository,
	}
}

type contactVoucherHistoryRepository struct {
	*Repository
}

func (r *contactVoucherHistoryRepository) Create(ctx context.Context, history *model.ContactVoucherHistory) error {
	return r.DB(ctx).Create(history).Error
}

func (r *contactVoucherHistoryRepository) ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.ContactVoucherHistory, int64, error) {
	var (
		histories []*model.ContactVoucherHistory
		total     int64
	)
	db := r.DB(ctx).Model(&model.ContactVoucherHistory{}).Where("user_id = ?", userID)
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
