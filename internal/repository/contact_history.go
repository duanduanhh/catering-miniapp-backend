package repository

import (
	"context"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type ContactHistoryRepository interface {
	Create(ctx context.Context, history *model.ContactHistory) error
	ListOut(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.ContactHistory, int64, error)
	ListIn(ctx context.Context, purposeUserID int64, bizType int, pageNum, pageSize int) ([]*model.ContactHistory, int64, error)
}

func NewContactHistoryRepository(
	repository *Repository,
) ContactHistoryRepository {
	return &contactHistoryRepository{
		Repository: repository,
	}
}

type contactHistoryRepository struct {
	*Repository
}

func (r *contactHistoryRepository) Create(ctx context.Context, history *model.ContactHistory) error {
	return r.DB(ctx).Create(history).Error
}

func (r *contactHistoryRepository) ListOut(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.ContactHistory, int64, error) {
	var (
		histories []*model.ContactHistory
		total     int64
	)
	db := r.DB(ctx).Model(&model.ContactHistory{}).Where("user_id = ?", userID)
	if bizType > 0 {
		db = db.Where("purpose_type = ?", bizType)
	}
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

func (r *contactHistoryRepository) ListIn(ctx context.Context, purposeUserID int64, bizType int, pageNum, pageSize int) ([]*model.ContactHistory, int64, error) {
	var (
		histories []*model.ContactHistory
		total     int64
	)
	db := r.DB(ctx).Model(&model.ContactHistory{}).Where("purpose_user_id = ?", purposeUserID)
	if bizType > 0 {
		db = db.Where("purpose_type = ?", bizType)
	}
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
