package repository

import (
	"context"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type CollectRepository interface {
	Create(ctx context.Context, collect *model.Collect) error
	Update(ctx context.Context, collect *model.Collect) error
	Get(ctx context.Context, userID, contentID int64, bizType int) (*model.Collect, error)
	ListByUser(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.Collect, int64, error)
}

func NewCollectRepository(
	repository *Repository,
) CollectRepository {
	return &collectRepository{
		Repository: repository,
	}
}

type collectRepository struct {
	*Repository
}

func (r *collectRepository) Create(ctx context.Context, collect *model.Collect) error {
	return r.DB(ctx).Create(collect).Error
}

func (r *collectRepository) Update(ctx context.Context, collect *model.Collect) error {
	return r.DB(ctx).Save(collect).Error
}

func (r *collectRepository) Get(ctx context.Context, userID, contentID int64, bizType int) (*model.Collect, error) {
	var collect model.Collect
	if err := r.DB(ctx).
		Where("user_id = ? AND content_id = ? AND type = ?", userID, contentID, bizType).
		First(&collect).Error; err != nil {
		return nil, err
	}
	return &collect, nil
}

func (r *collectRepository) ListByUser(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.Collect, int64, error) {
	var (
		collects []*model.Collect
		total    int64
	)
	db := r.DB(ctx).Model(&model.Collect{}).Where("user_id = ? AND status = ?", userID, model.CollectStatusActive)
	if bizType > 0 {
		db = db.Where("type = ?", bizType)
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
	if err := db.Order("create_at DESC").Offset(offset).Limit(pageSize).Find(&collects).Error; err != nil {
		return nil, 0, err
	}
	return collects, total, nil
}
