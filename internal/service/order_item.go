package service

import (
	"context"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type OrderItemService interface {
	ListByOrderID(ctx context.Context, orderID int64) ([]*model.OrderItem, error)
}
func NewOrderItemService(
    service *Service,
    orderItemRepository repository.OrderItemRepository,
) OrderItemService {
	return &orderItemService{
		Service:        service,
		orderItemRepository: orderItemRepository,
	}
}

type orderItemService struct {
	*Service
	orderItemRepository repository.OrderItemRepository
}

func (s *orderItemService) ListByOrderID(ctx context.Context, orderID int64) ([]*model.OrderItem, error) {
	return s.orderItemRepository.ListByOrderID(ctx, orderID)
}
