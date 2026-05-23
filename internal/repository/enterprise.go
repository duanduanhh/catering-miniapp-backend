package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type EnterpriseRepository interface {
	Create(ctx context.Context, e *model.Enterprise) error
	Update(ctx context.Context, e *model.Enterprise) error
	GetByID(ctx context.Context, id int64) (*model.Enterprise, error)
	GetByUserAndCode(ctx context.Context, userID int64, code string) (*model.Enterprise, error)
	ListByUser(ctx context.Context, userID int64) ([]*model.Enterprise, error)
	ListVerifiedByUser(ctx context.Context, userID int64) ([]*model.Enterprise, error)
	ClearDefault(ctx context.Context, userID int64) error
}

func NewEnterpriseRepository(r *Repository) EnterpriseRepository {
	return &enterpriseRepository{Repository: r}
}

type enterpriseRepository struct {
	*Repository
}

func (r *enterpriseRepository) Create(ctx context.Context, e *model.Enterprise) error {
	return r.DB(ctx).Create(e).Error
}

func (r *enterpriseRepository) Update(ctx context.Context, e *model.Enterprise) error {
	return r.DB(ctx).Save(e).Error
}

func (r *enterpriseRepository) GetByID(ctx context.Context, id int64) (*model.Enterprise, error) {
	var e model.Enterprise
	if err := r.DB(ctx).Where("id = ? AND status != ?", id, model.EnterpriseStatusDeleted).First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *enterpriseRepository) GetByUserAndCode(ctx context.Context, userID int64, code string) (*model.Enterprise, error) {
	var e model.Enterprise
	if err := r.DB(ctx).
		Where("user_id = ? AND social_credit_code = ? AND status != ?", userID, code, model.EnterpriseStatusDeleted).
		First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *enterpriseRepository) ListByUser(ctx context.Context, userID int64) ([]*model.Enterprise, error) {
	var list []*model.Enterprise
	if err := r.DB(ctx).
		Where("user_id = ? AND status != ?", userID, model.EnterpriseStatusDeleted).
		Order("is_default DESC, create_at ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *enterpriseRepository) ListVerifiedByUser(ctx context.Context, userID int64) ([]*model.Enterprise, error) {
	var list []*model.Enterprise
	if err := r.DB(ctx).
		Where("user_id = ? AND status = ?", userID, model.EnterpriseStatusVerified).
		Order("is_default DESC, create_at ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ClearDefault 将该用户所有企业的 is_default 置 0，在事务中调用。
func (r *enterpriseRepository) ClearDefault(ctx context.Context, userID int64) error {
	return r.DB(ctx).Model(&model.Enterprise{}).
		Where("user_id = ? AND status != ?", userID, model.EnterpriseStatusDeleted).
		Update("is_default", 0).Error
}
