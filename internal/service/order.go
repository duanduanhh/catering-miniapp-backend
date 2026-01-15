package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type OrderService interface {
	CreateTopOrder(ctx context.Context, userID, jobID int64, topHour int, price float64) (*model.Order, *model.OrderItem, error)
	CreateContactVoucherOrder(ctx context.Context, userID int64, price float64, voucherNum int) (*model.Order, *model.OrderItem, error)
	CreateRefreshOrder(ctx context.Context, userID, jobID int64, price float64) (*model.Order, *model.OrderItem, error)
	PayOrder(ctx context.Context, userID, orderID int64, orderNo string, amount float64, payChannel, payTradeNo string) (*model.Order, error)
	PayOrderByNotify(ctx context.Context, orderNo string, amount float64, payChannel, payTradeNo string) (*model.Order, error)
}

func NewOrderService(
	service *Service,
	orderRepository repository.OrderRepository,
	orderItemRepository repository.OrderItemRepository,
	jobRepository repository.JobRepository,
	userRepository repository.UserRepository,
	contactVoucherHistoryRepository repository.ContactVoucherHistoryRepository,
) OrderService {
	return &orderService{
		Service:                         service,
		orderRepository:                 orderRepository,
		orderItemRepository:             orderItemRepository,
		jobRepository:                   jobRepository,
		userRepository:                  userRepository,
		contactVoucherHistoryRepository: contactVoucherHistoryRepository,
	}
}

type orderService struct {
	*Service
	orderRepository                 repository.OrderRepository
	orderItemRepository             repository.OrderItemRepository
	jobRepository                   repository.JobRepository
	userRepository                  repository.UserRepository
	contactVoucherHistoryRepository repository.ContactVoucherHistoryRepository
}

func (s *orderService) CreateTopOrder(ctx context.Context, userID, jobID int64, topHour int, price float64) (*model.Order, *model.OrderItem, error) {
	job, err := s.jobRepository.GetByID(ctx, jobID)
	if err != nil {
		return nil, nil, err
	}
	if job.UserID != userID {
		return nil, nil, ErrForbidden
	}
	order := &model.Order{
		OrderNo:     s.generateOrderNo("TOP"),
		UserID:      userID,
		AmountTotal: model.NewDecimalFromFloat64(price),
		AmountPaid:  model.NewDecimalFromFloat64(0),
		Currency:    "CNY",
		Status:      model.OrderStatusPending,
		CreateAt:    time.Now(),
		UpdateAt:    time.Now(),
	}
	item := &model.OrderItem{
		ProductType:       model.ProductTypeTop,
		TitleSnapshot:     fmt.Sprintf("置顶套餐-%d小时", topHour),
		TopHour:           topHour,
		UnitPriceSnapshot: price,
		TargetType:        model.OrderTargetJob,
		TargetID:          jobID,
		CreateAt:          time.Now(),
		UpdateAt:          time.Now(),
	}

	err = s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.orderRepository.Create(ctx, order); err != nil {
			return err
		}
		item.OrderID = order.ID
		return s.orderItemRepository.Create(ctx, item)
	})
	if err != nil {
		return nil, nil, err
	}
	return order, item, nil
}

func (s *orderService) CreateContactVoucherOrder(ctx context.Context, userID int64, price float64, voucherNum int) (*model.Order, *model.OrderItem, error) {
	order := &model.Order{
		OrderNo:     s.generateOrderNo("CV"),
		UserID:      userID,
		AmountTotal: model.NewDecimalFromFloat64(price),
		AmountPaid:  model.NewDecimalFromFloat64(0),
		Currency:    "CNY",
		Status:      model.OrderStatusPending,
		CreateAt:    time.Now(),
		UpdateAt:    time.Now(),
	}
	item := &model.OrderItem{
		ProductType:       model.ProductTypeContactVoucher,
		TitleSnapshot:     fmt.Sprintf("联系券-%d张", voucherNum),
		UnitPriceSnapshot: price,
		ContactVoucherNum: voucherNum,
		CreateAt:          time.Now(),
		UpdateAt:          time.Now(),
	}
	err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.orderRepository.Create(ctx, order); err != nil {
			return err
		}
		item.OrderID = order.ID
		return s.orderItemRepository.Create(ctx, item)
	})
	if err != nil {
		return nil, nil, err
	}
	return order, item, nil
}

func (s *orderService) CreateRefreshOrder(ctx context.Context, userID, jobID int64, price float64) (*model.Order, *model.OrderItem, error) {
	job, err := s.jobRepository.GetByID(ctx, jobID)
	if err != nil {
		return nil, nil, err
	}
	if job.UserID != userID {
		return nil, nil, ErrForbidden
	}
	order := &model.Order{
		OrderNo:     s.generateOrderNo("REF"),
		UserID:      userID,
		AmountTotal: model.NewDecimalFromFloat64(price),
		AmountPaid:  model.NewDecimalFromFloat64(0),
		Currency:    "CNY",
		Status:      model.OrderStatusPending,
		CreateAt:    time.Now(),
		UpdateAt:    time.Now(),
	}
	item := &model.OrderItem{
		ProductType:       model.ProductTypeRefresh,
		TitleSnapshot:     "刷新招聘",
		UnitPriceSnapshot: price,
		TargetType:        model.OrderTargetJob,
		TargetID:          jobID,
		CreateAt:          time.Now(),
		UpdateAt:          time.Now(),
	}
	err = s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.orderRepository.Create(ctx, order); err != nil {
			return err
		}
		item.OrderID = order.ID
		return s.orderItemRepository.Create(ctx, item)
	})
	if err != nil {
		return nil, nil, err
	}
	return order, item, nil
}

func (s *orderService) PayOrder(ctx context.Context, userID, orderID int64, orderNo string, amount float64, payChannel, payTradeNo string) (*model.Order, error) {
	var order *model.Order
	var err error
	if orderID > 0 {
		order, err = s.orderRepository.GetByID(ctx, orderID)
	} else {
		order, err = s.orderRepository.GetByOrderNo(ctx, orderNo)
	}
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrForbidden
	}
	return s.payOrderWithItems(ctx, order, amount, payChannel, payTradeNo)
}

func (s *orderService) PayOrderByNotify(ctx context.Context, orderNo string, amount float64, payChannel, payTradeNo string) (*model.Order, error) {
	order, err := s.orderRepository.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	return s.payOrderWithItems(ctx, order, amount, payChannel, payTradeNo)
}

func (s *orderService) payOrderWithItems(ctx context.Context, order *model.Order, amount float64, payChannel, payTradeNo string) (*model.Order, error) {
	if order.Status != model.OrderStatusPending {
		return order, nil
	}
	if amount > 0 {
		expected, err := order.AmountTotal.ToCents()
		if err != nil {
			return nil, err
		}
		given := int64(amount*100 + 0.5)
		if expected != given {
			return nil, ErrAmountMismatch
		}
	}
	items, err := s.orderItemRepository.ListByOrderID(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	err = s.tm.Transaction(ctx, func(ctx context.Context) error {
		order.Status = model.OrderStatusPaid
		order.AmountPaid = order.AmountTotal
		order.PayChannel = payChannel
		order.PayTradeNo = payTradeNo
		paidAt := time.Now()
		order.PaidAt = &paidAt
		order.UpdateAt = time.Now()
		if err := s.orderRepository.Update(ctx, order); err != nil {
			return err
		}
		for _, item := range items {
			switch item.ProductType {
			case model.ProductTypeTop:
				if err := s.applyTop(ctx, item); err != nil {
					return err
				}
			case model.ProductTypeContactVoucher:
				if err := s.applyContactVoucher(ctx, order.UserID, item); err != nil {
					return err
				}
			case model.ProductTypeRefresh:
				if err := s.applyRefresh(ctx, item); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) applyTop(ctx context.Context, item *model.OrderItem) error {
	job, err := s.jobRepository.GetByID(ctx, item.TargetID)
	if err != nil {
		return err
	}
	now := time.Now()
	baseTime := now
	if job.TopEndTime != nil && job.TopEndTime.After(now) {
		baseTime = *job.TopEndTime
	}
	if job.TopStartTime == nil || (job.TopEndTime != nil && !job.TopEndTime.After(now)) {
		job.TopStartTime = &now
	}
	end := baseTime.Add(time.Duration(item.TopHour) * time.Hour)
	job.TopEndTime = &end
	job.UpdateAt = now
	return s.jobRepository.Update(ctx, job)
}

func (s *orderService) applyContactVoucher(ctx context.Context, userID int64, item *model.OrderItem) error {
	voucherNum := item.ContactVoucherNum
	if voucherNum <= 0 {
		voucherNum = parseVoucherNum(item.TitleSnapshot)
	}
	if voucherNum <= 0 {
		return ErrInvalidVoucherNum
	}
	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	lastNum := user.ContactVoucherNum
	nextNum := lastNum + voucherNum
	user.ContactVoucherNum = nextNum
	user.UpdateAt = time.Now()
	if err := s.userRepository.Update(ctx, user); err != nil {
		return err
	}
	history := &model.ContactVoucherHistory{
		UserID:    userID,
		BizType:   model.ContactVoucherHistoryBuy,
		ChangeNum: voucherNum,
		LastNum:   lastNum,
		NextNum:   nextNum,
		Remark:    "购买联系券",
		CreateAt:  time.Now(),
	}
	return s.contactVoucherHistoryRepository.Create(ctx, history)
}

func (s *orderService) applyRefresh(ctx context.Context, item *model.OrderItem) error {
	job, err := s.jobRepository.GetByID(ctx, item.TargetID)
	if err != nil {
		return err
	}
	now := time.Now()
	job.RefreshTime = &now
	return s.jobRepository.Update(ctx, job)
}

func (s *orderService) generateOrderNo(prefix string) string {
	id, err := s.sid.GenUint64()
	if err != nil {
		return fmt.Sprintf("%s%s", prefix, time.Now().Format("20060102150405"))
	}
	return fmt.Sprintf("%s%s%06d", prefix, time.Now().Format("20060102150405"), id%1000000)
}

func parseVoucherNum(title string) int {
	var num int
	_, _ = fmt.Sscanf(title, "联系券-%d张", &num)
	return num
}
