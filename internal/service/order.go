package service

import (
	"context"
	"fmt"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"time"
)

type OrderService interface {
	CreateTopOrder(ctx context.Context, userID, jobID int64) (*model.Order, *model.OrderItem, TopRule, error)
	CreateContactVoucherOrder(ctx context.Context, userID int64, price float64, voucherNum int) (*model.Order, *model.OrderItem, error)
	PayOrder(ctx context.Context, userID, orderID int64, orderNo string, amount float64, payChannel, payTradeNo string) (*model.Order, error)
}
func NewOrderService(
    service *Service,
    orderRepository repository.OrderRepository,
	orderItemRepository repository.OrderItemRepository,
	jobRepository repository.JobRepository,
	userRepository repository.UserRepository,
	costHistoryRepository repository.CostHistoryRepository,
) OrderService {
	return &orderService{
		Service:        service,
		orderRepository: orderRepository,
		orderItemRepository: orderItemRepository,
		jobRepository: jobRepository,
		userRepository: userRepository,
		costHistoryRepository: costHistoryRepository,
	}
}

type orderService struct {
	*Service
	orderRepository repository.OrderRepository
	orderItemRepository repository.OrderItemRepository
	jobRepository       repository.JobRepository
	userRepository      repository.UserRepository
	costHistoryRepository repository.CostHistoryRepository
}

const (
	OrderStatusPending = 1
	OrderStatusPaid    = 2
	OrderStatusCanceled = 3
	OrderStatusRefunded = 4

	ProductTypeTop           = 1
	ProductTypeContactVoucher = 2
)

const (
	CostHistoryRecharge = 1
	CostHistoryConsume  = 2
)

type TopRule struct {
	RuleID int64
	Hours  int
	Price  float64
}

func (s *orderService) CreateTopOrder(ctx context.Context, userID, jobID int64) (*model.Order, *model.OrderItem, TopRule, error) {
	job, err := s.jobRepository.GetByID(ctx, jobID)
	if err != nil {
		return nil, nil, TopRule{}, err
	}
	if job.UserID != userID {
		return nil, nil, TopRule{}, ErrForbidden
	}
	rule := TopRule{RuleID: 2, Hours: 72, Price: 5.00}
	order := &model.Order{
		OrderNo:     s.generateOrderNo("TOP"),
		UserID:      userID,
		AmountTotal: model.NewDecimalFromFloat64(rule.Price),
		AmountPaid:  model.NewDecimalFromFloat64(0),
		Currency:    "CNY",
		Status:      OrderStatusPending,
		CreateAt:    time.Now(),
		UpdateAt:    time.Now(),
	}
	item := &model.OrderItem{
		ProductType:       ProductTypeTop,
		RuleID:            rule.RuleID,
		TitleSnapshot:     fmt.Sprintf("置顶套餐-%d小时", rule.Hours),
		UnitPriceSnapshot: rule.Price,
		TargetType:        1,
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
		return nil, nil, rule, err
	}
	return order, item, rule, nil
}

func (s *orderService) CreateContactVoucherOrder(ctx context.Context, userID int64, price float64, voucherNum int) (*model.Order, *model.OrderItem, error) {
	order := &model.Order{
		OrderNo:     s.generateOrderNo("CV"),
		UserID:      userID,
		AmountTotal: model.NewDecimalFromFloat64(price),
		AmountPaid:  model.NewDecimalFromFloat64(0),
		Currency:    "CNY",
		Status:      OrderStatusPending,
		CreateAt:    time.Now(),
		UpdateAt:    time.Now(),
	}
	item := &model.OrderItem{
		ProductType:       ProductTypeContactVoucher,
		RuleID:            int64(voucherNum),
		TitleSnapshot:     fmt.Sprintf("联系券-%d张", voucherNum),
		UnitPriceSnapshot: price,
		PurposeUserID:     0,
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
	if order.Status != OrderStatusPending {
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
		order.Status = OrderStatusPaid
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
			case ProductTypeTop:
				if err := s.applyTop(ctx, item); err != nil {
					return err
				}
			case ProductTypeContactVoucher:
				if err := s.applyContactVoucher(ctx, order.UserID, item); err != nil {
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
	hours := s.topHoursByRule(item.RuleID)
	now := time.Now()
	job.IsTop = 1
	job.IsBuyTop = 1
	job.TopHour = hours
	job.TopStartTime = &now
	end := now.Add(time.Duration(hours) * time.Hour)
	job.TopEndTime = &end
	job.UpdateAt = now
	return s.jobRepository.Update(ctx, job)
}

func (s *orderService) applyContactVoucher(ctx context.Context, userID int64, item *model.OrderItem) error {
	voucherNum := int(item.RuleID)
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
	history := &model.CostHistory{
		UserID:    userID,
		BizType:   CostHistoryRecharge,
		ChangeNum: voucherNum,
		LastNum:   lastNum,
		NextNum:   nextNum,
		Remark:    "购买联系券",
		CreateAt:  time.Now(),
	}
	return s.costHistoryRepository.Create(ctx, history)
}

func (s *orderService) topHoursByRule(ruleID int64) int {
	switch ruleID {
	case 1:
		return 24
	case 2:
		return 72
	default:
		return 24
	}
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
