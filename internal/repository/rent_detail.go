package repository

import (
	"context"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type RentDetailRepository interface {
	Create(ctx context.Context, d *model.RentDetail) error
	Update(ctx context.Context, d *model.RentDetail) error
	GetByJobID(ctx context.Context, jobID int64) (*model.RentDetail, error)
	GetByJobIDs(ctx context.Context, ids []int64) (map[int64]*model.RentDetail, error)
	DeleteByJobID(ctx context.Context, jobID int64) error
}

func NewRentDetailRepository(
	repository *Repository,
) RentDetailRepository {
	return &rentDetailRepository{
		Repository: repository,
	}
}

type rentDetailRepository struct {
	*Repository
}

func (r *rentDetailRepository) Create(ctx context.Context, d *model.RentDetail) error {
	if d.CreateAt.IsZero() {
		d.CreateAt = time.Now()
	}
	d.UpdateAt = time.Now()
	return r.DB(ctx).Create(d).Error
}

func (r *rentDetailRepository) Update(ctx context.Context, d *model.RentDetail) error {
	d.UpdateAt = time.Now()
	return r.DB(ctx).Save(d).Error
}

func (r *rentDetailRepository) GetByJobID(ctx context.Context, jobID int64) (*model.RentDetail, error) {
	var detail model.RentDetail
	if err := r.DB(ctx).Where("job_id = ?", jobID).First(&detail).Error; err != nil {
		return nil, err
	}
	return &detail, nil
}

func (r *rentDetailRepository) GetByJobIDs(ctx context.Context, ids []int64) (map[int64]*model.RentDetail, error) {
	if len(ids) == 0 {
		return map[int64]*model.RentDetail{}, nil
	}
	var details []*model.RentDetail
	if err := r.DB(ctx).Where("job_id IN ?", ids).Find(&details).Error; err != nil {
		return nil, err
	}
	result := make(map[int64]*model.RentDetail, len(details))
	for _, d := range details {
		result[d.JobID] = d
	}
	return result, nil
}

func (r *rentDetailRepository) DeleteByJobID(ctx context.Context, jobID int64) error {
	return r.DB(ctx).Where("job_id = ?", jobID).Delete(&model.RentDetail{}).Error
}
