package repository

import (
	"context"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type ContactHistoryRepository interface {
	Create(ctx context.Context, history *model.ContactHistory) error
	ListOut(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.ContactHistory, int64, error)
	ListIn(ctx context.Context, purposeUserID int64, bizType int, pageNum, pageSize int) ([]*model.ContactHistory, int64, error)
	DeleteOut(ctx context.Context, userID, purposeID int64) error
	DeleteIn(ctx context.Context, purposeUserID, purposeID int64) error
	ExistsByUserAndJob(ctx context.Context, userID, jobID int64, purposeType int) (bool, error)
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
	db := r.DB(ctx).Model(&model.ContactHistory{}).Where("user_id = ? AND user_deleted = 0", userID)
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
	db := r.DB(ctx).Model(&model.ContactHistory{}).Where("purpose_user_id = ? AND purpose_user_deleted = 0", purposeUserID)
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

func (r *contactHistoryRepository) ExistsByUserAndJob(ctx context.Context, userID, jobID int64, purposeType int) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.ContactHistory{}).
		Where("user_id = ? AND purpose_id = ? AND purpose_type = ? AND user_deleted = 0", userID, jobID, purposeType).
		Count(&count).Error
	return count > 0, err
}

// DeleteOut 删除"我联系的"记录（软删除）
func (r *contactHistoryRepository) DeleteOut(ctx context.Context, userID, purposeID int64) error {
	return r.DB(ctx).Model(&model.ContactHistory{}).Where("user_id = ? AND purpose_id = ?", userID, purposeID).Update("user_deleted", 1).Error
}

// DeleteIn 删除"联系我的"记录（软删除）
func (r *contactHistoryRepository) DeleteIn(ctx context.Context, purposeUserID, purposeID int64) error {
	return r.DB(ctx).Model(&model.ContactHistory{}).Where("purpose_user_id = ? AND purpose_id = ?", purposeUserID, purposeID).Update("purpose_user_deleted", 1).Error
}
