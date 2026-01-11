package repository

import (
	"context"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type OrderItemRepository interface {
	Create(ctx context.Context, item *model.OrderItem) error
	ListByOrderID(ctx context.Context, orderID int64) ([]*model.OrderItem, error)
}

func NewOrderItemRepository(
	repository *Repository,
) OrderItemRepository {
	return &orderItemRepository{
		Repository: repository,
	}
}

type orderItemRepository struct {
	*Repository
}

func (r *orderItemRepository) Create(ctx context.Context, item *model.OrderItem) error {
	return r.DB(ctx).Create(item).Error
}

func (r *orderItemRepository) ListByOrderID(ctx context.Context, orderID int64) ([]*model.OrderItem, error) {
	var items []*model.OrderItem
	if err := r.DB(ctx).Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
