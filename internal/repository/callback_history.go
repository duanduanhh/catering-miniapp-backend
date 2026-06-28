package repository

import (
	"context"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type CallbackHistoryRepository interface {
	Create(ctx context.Context, history *model.CallbackHistory) error
}

func NewCallbackHistoryRepository(
	repository *Repository,
) CallbackHistoryRepository {
	return &callbackHistoryRepository{
		Repository: repository,
	}
}

type callbackHistoryRepository struct {
	*Repository
}

func (r *callbackHistoryRepository) Create(ctx context.Context, history *model.CallbackHistory) error {
	return r.DB(ctx).Create(history).Error
}
