package repository

import (
	"context"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type PositionCategoryRepository interface {
	GetAllWithSubcategories(ctx context.Context) ([]*model.PositionCategory, error)
}

type positionCategoryRepository struct {
	*Repository
}

func NewPositionCategoryRepository(repository *Repository) PositionCategoryRepository {
	return &positionCategoryRepository{
		Repository: repository,
	}
}

func (r *positionCategoryRepository) GetAllWithSubcategories(ctx context.Context) ([]*model.PositionCategory, error) {
	var categories []*model.PositionCategory
	err := r.DB(ctx).Preload("Subcategories").Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}
