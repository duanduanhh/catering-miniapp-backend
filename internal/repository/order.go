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
	ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.OrderWithItem, int64, error)
	HasPaidOrderByUserIDs(ctx context.Context, userIDs []int64) (map[int64]bool, error)
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

func (r *orderRepository) ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.OrderWithItem, int64, error) {
	var (
		list  []*model.OrderWithItem
		total int64
	)
	db := r.DB(ctx).Table("orders").
		Select("orders.*, order_item.product_type, order_item.title_snapshot, order_item.unit_price_snapshot").
		Joins("LEFT JOIN order_item ON order_item.order_id = orders.id").
		Where("orders.user_id = ? AND orders.status = ?", userID, model.OrderStatusPaid)

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
	if err := db.Order("orders.create_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *orderRepository) HasPaidOrderByUserIDs(ctx context.Context, userIDs []int64) (map[int64]bool, error) {
	if len(userIDs) == 0 {
		return map[int64]bool{}, nil
	}
	var rows []struct {
		UserID int64 `gorm:"column:user_id"`
	}
	err := r.DB(ctx).Table("orders").
		Select("user_id").
		Where("user_id IN ? AND status = ?", userIDs, model.OrderStatusPaid).
		Group("user_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]bool, len(rows))
	for _, row := range rows {
		result[row.UserID] = true
	}
	return result, nil
}
