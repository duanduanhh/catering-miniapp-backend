package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByPhone(ctx context.Context, phone string) (*model.User, error)
	GetByOpenID(ctx context.Context, openID string) (*model.User, error)
	ListByIDs(ctx context.Context, ids []int64) ([]*model.User, error)
	ListByInviterID(ctx context.Context, inviterID int64, pageNum, pageSize int) ([]*model.User, int64, error)
}

func NewUserRepository(
	r *Repository,
) UserRepository {
	return &userRepository{
		Repository: r,
	}
}

type userRepository struct {
	*Repository
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.DB(ctx).Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	if err := r.DB(ctx).Save(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, userId int64) (*model.User, error) {
	var user model.User
	if err := r.DB(ctx).Where("id = ?", userId).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	if err := r.DB(ctx).Where("phone = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByOpenID(ctx context.Context, openID string) (*model.User, error) {
	var user model.User
	if err := r.DB(ctx).Where("wechat_open_id = ?", openID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ListByIDs(ctx context.Context, ids []int64) ([]*model.User, error) {
	if len(ids) == 0 {
		return []*model.User{}, nil
	}
	var users []*model.User
	if err := r.DB(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) ListByInviterID(ctx context.Context, inviterID int64, pageNum, pageSize int) ([]*model.User, int64, error) {
	var (
		list  []*model.User
		total int64
	)
	db := r.DB(ctx).Model(&model.User{}).Where("inviter_id = ?", inviterID)
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
