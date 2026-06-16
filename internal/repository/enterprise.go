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
	AdminList(ctx context.Context, query AdminEnterpriseListQuery) ([]*AdminEnterpriseRow, int64, error)
}

// AdminEnterpriseListQuery 管理后台企业列表筛选条件
type AdminEnterpriseListQuery struct {
	Keyword  string // name / social_credit_code 模糊匹配
	Status   *int
	UserID   int64
	PageNum  int
	PageSize int
}

// AdminEnterpriseRow 企业记录 + JOIN user 出的发布人信息
type AdminEnterpriseRow struct {
	model.Enterprise
	UserName  string `gorm:"column:user_name"`
	UserPhone string `gorm:"column:user_phone"`
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

// AdminList 管理后台企业列表，JOIN user 拿发布人姓名/手机号。
func (r *enterpriseRepository) AdminList(ctx context.Context, query AdminEnterpriseListQuery) ([]*AdminEnterpriseRow, int64, error) {
	var (
		rows  []*AdminEnterpriseRow
		total int64
	)
	db := r.DB(ctx).
		Table("enterprise AS e").
		Select("e.*, u.name AS user_name, u.phone AS user_phone").
		Joins("LEFT JOIN user u ON e.user_id = u.id").
		Where("e.status != ?", model.EnterpriseStatusDeleted)

	if kw := query.Keyword; kw != "" {
		like := "%" + kw + "%"
		db = db.Where("e.name LIKE ? OR e.social_credit_code LIKE ?", like, like)
	}
	if query.Status != nil {
		db = db.Where("e.status = ?", *query.Status)
	}
	if query.UserID > 0 {
		db = db.Where("e.user_id = ?", query.UserID)
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
	if err := db.Order("e.create_at DESC").Offset(offset).Limit(query.PageSize).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
