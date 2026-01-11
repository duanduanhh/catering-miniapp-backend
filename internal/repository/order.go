package repository

import (
	"context"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	Update(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id int64) (*model.Order, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error)
}

func NewOrderRepository(
	repository *Repository,
) OrderRepository {
	return &orderRepository{
		Repository: repository,
	}
}

type orderRepository struct {
	*Repository
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	return r.DB(ctx).Create(order).Error
}

func (r *orderRepository) Update(ctx context.Context, order *model.Order) error {
	return r.DB(ctx).Save(order).Error
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	var order model.Order
	if err := r.DB(ctx).Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	var order model.Order
	if err := r.DB(ctx).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}
