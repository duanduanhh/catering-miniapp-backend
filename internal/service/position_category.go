package service

import (
	"context"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type PositionCategoryService interface {
	GetAllCategories(ctx context.Context) ([]*model.PositionCategory, error)
}

type positionCategoryService struct {
	repo repository.PositionCategoryRepository
}

func NewPositionCategoryService(repo repository.PositionCategoryRepository) PositionCategoryService {
	return &positionCategoryService{
		repo: repo,
	}
}

func (s *positionCategoryService) GetAllCategories(ctx context.Context) ([]*model.PositionCategory, error) {
	return s.repo.GetAllWithSubcategories(ctx)
}
