package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type OrderService interface {
	CreateTopOrder(ctx context.Context, userID, jobID int64, skuCode string) (*model.Order, *model.OrderItem, error)
	CreateContactVoucherOrder(ctx context.Context, userID int64, skuCode string) (*model.Order, *model.OrderItem, error)
	CreateRefreshOrder(ctx context.Context, userID, jobID int64, skuCode string) (*model.Order, *model.OrderItem, error)
	PayOrder(ctx context.Context, userID, orderID int64, orderNo string, amount float64, payChannel, payTradeNo string) (*model.Order, error)
	PayOrderByNotify(ctx context.Context, orderNo string, amount float64, payChannel, payTradeNo string) (*model.Order, error)
	ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.OrderWithItem, int64, error)
	GetByOrderNo(ctx context.Context, userID int64, orderNo string) (*model.Order, error)
}

func NewOrderService(
	service *Service,
	orderRepository repository.OrderRepository,
	orderItemRepository repository.OrderItemRepository,
	jobRepository repository.JobRepository,
	userRepository repository.UserRepository,
	contactVoucherHistoryRepository repository.ContactVoucherHistoryRepository,
	paymentPackageService PaymentPackageService,
) OrderService {
	return &orderService{
		Service:                         service,
		orderRepository:                 orderRepository,
		orderItemRepository:             orderItemRepository,
		jobRepository:                   jobRepository,
		userRepository:                  userRepository,
		contactVoucherHistoryRepository: contactVoucherHistoryRepository,
		paymentPackageService:           paymentPackageService,
	}
}

type orderService struct {
	*Service
	orderRepository                 repository.OrderRepository
	orderItemRepository             repository.OrderItemRepository
	jobRepository                   repository.JobRepository
	userRepository                  repository.UserRepository
	contactVoucherHistoryRepository repository.ContactVoucherHistoryRepository
	paymentPackageService           PaymentPackageService
}

func (s *orderService) CreateTopOrder(
	ctx context.Context,
	userID, jobID int64,
	skuCode string,
) (*model.Order, *model.OrderItem, error) {
	job, err := s.jobRepository.GetByID(ctx, jobID)
	if err != nil {
		return nil, nil, err
	}
	if job.UserID != userID {
		return nil, nil, ErrForbidden
	}
	var order *model.Order
	var item *model.OrderItem
	err = s.tm.Transaction(ctx, func(ctx context.Context) error {
		pkg, err := s.paymentPackageService.ResolveForPurchase(
			ctx, userID, skuCode, model.PaymentProductCodeJobTop, job.BizType,
		)
		if err != nil {
			return err
		}
		topHour := pkg.BenefitConfig.TopHours
		contactVoucherNum := pkg.BenefitConfig.GiftContactVouchers
		order, item = buildSKUOrder(generatePaymentOrderNo(s.Service, "TOP"), userID, model.ProductTypeTop, pkg)
		item.TitleSnapshot = fmt.Sprintf("置顶%d天", topHour/24)
		item.TopHour = topHour
		item.ContactVoucherNum = contactVoucherNum
		item.TargetType = model.OrderTargetJob
		item.TargetID = jobID
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

func (s *orderService) CreateContactVoucherOrder(
	ctx context.Context,
	userID int64,
	skuCode string,
) (*model.Order, *model.OrderItem, error) {
	var order *model.Order
	var item *model.OrderItem
	err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		pkg, err := s.paymentPackageService.ResolveForPurchase(
			ctx, userID, skuCode, model.PaymentProductCodeContactVoucher, 0,
		)
		if err != nil {
			return err
		}
		voucherNum := pkg.BenefitConfig.ContactVouchers
		giftNum := pkg.BenefitConfig.GiftContactVouchers
		order, item = buildSKUOrder(generatePaymentOrderNo(s.Service, "CV"), userID, model.ProductTypeContactVoucher, pkg)
		item.TitleSnapshot = pkg.Package.Name
		item.ContactVoucherNum = voucherNum + giftNum
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

func (s *orderService) CreateRefreshOrder(
	ctx context.Context,
	userID, jobID int64,
	skuCode string,
) (*model.Order, *model.OrderItem, error) {
	job, err := s.jobRepository.GetByID(ctx, jobID)
	if err != nil {
		return nil, nil, err
	}
	if job.UserID != userID {
		return nil, nil, ErrForbidden
	}
	var order *model.Order
	var item *model.OrderItem
	err = s.tm.Transaction(ctx, func(ctx context.Context) error {
		pkg, err := s.paymentPackageService.ResolveForPurchase(
			ctx, userID, skuCode, model.PaymentProductCodePaidRefresh, job.BizType,
		)
		if err != nil {
			return err
		}
		order, item = buildSKUOrder(generatePaymentOrderNo(s.Service, "REF"), userID, model.ProductTypeRefresh, pkg)
		item.TitleSnapshot = pkg.Package.Name
		item.TargetType = model.OrderTargetJob
		item.TargetID = jobID
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

func buildSKUOrder(
	orderNo string,
	userID int64,
	productType model.ProductType,
	pkg *PaymentPackageAggregate,
) (*model.Order, *model.OrderItem) {
	now := time.Now()
	price := float64(pkg.Package.PriceCents) / 100
	benefitSnapshot, _ := json.Marshal(pkg.BenefitConfig)
	order := &model.Order{
		OrderNo:     orderNo,
		UserID:      userID,
		AmountTotal: model.NewDecimalFromCents(pkg.Package.PriceCents),
		AmountPaid:  model.NewDecimalFromFloat64(0),
		Currency:    "CNY",
		Status:      model.OrderStatusPending,
		CreateAt:    now,
		UpdateAt:    now,
	}
	item := &model.OrderItem{
		ProductType:        productType,
		ProductID:          pkg.Product.ID,
		SKUID:              pkg.Package.ID,
		SKUCode:            pkg.Package.SKUCode,
		SKUVersion:         pkg.Package.Version,
		TitleSnapshot:      pkg.Package.Name,
		UnitPriceSnapshot:  price,
		PriceCentsSnapshot: pkg.Package.PriceCents,
		BenefitSnapshot:    string(benefitSnapshot),
		CreateAt:           now,
		UpdateAt:           now,
	}
	return order, item
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
				if item.ContactVoucherNum > 0 {
					if err := s.applyContactVoucher(ctx, order.UserID, item, "购买置顶套餐赠送联系券"); err != nil {
						return err
					}
				}
				if err := s.applyFirstTopStatus(ctx, order.UserID); err != nil {
					return err
				}
			case model.ProductTypeContactVoucher:
				if err := s.applyContactVoucher(ctx, order.UserID, item, "购买联系券"); err != nil {
					return err
				}
			case model.ProductTypeRefresh:
				if err := s.applyRefresh(ctx, item); err != nil {
					return err
				}
			case model.ProductTypePublishRent:
				// 招租发布支付成功：将 job 从待支付翻转为 Active，刷新时间更新为当前
				if err := s.jobRepository.ActivatePendingRent(ctx, item.TargetID); err != nil {
					return err
				}
			}
		}
		// 任意付费订单支付成功都触发"首次消费"判定；
		// applyNewCustomerStatus 内部以 user.new_customer_status 字段幂等，重复调用安全。
		if err := s.applyNewCustomerStatus(ctx, order.UserID); err != nil {
			return err
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

func (s *orderService) applyContactVoucher(ctx context.Context, userID int64, item *model.OrderItem, remark string) error {
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
		Remark:    remark,
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
	job.PaidRefreshTime = &now
	return s.jobRepository.Update(ctx, job)
}

func (s *orderService) applyFirstTopStatus(ctx context.Context, userID int64) error {
	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.FirstTopStatus == 0 {
		user.FirstTopStatus = 1
		user.UpdateAt = time.Now()
		return s.userRepository.Update(ctx, user)
	}
	return nil
}

func (s *orderService) applyNewCustomerStatus(ctx context.Context, userID int64) error {
	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.NewCustomerStatus == 0 {
		user.NewCustomerStatus = 1
		user.UpdateAt = time.Now()
		if err := s.userRepository.Update(ctx, user); err != nil {
			return err
		}
		return rewardInviter(ctx, userID, 5, "邀请好友首次消费奖励", s.userRepository, s.contactVoucherHistoryRepository)
	}
	return nil
}

func (s *orderService) ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.OrderWithItem, int64, error) {
	return s.orderRepository.ListByUser(ctx, userID, pageNum, pageSize)
}

func (s *orderService) GetByOrderNo(ctx context.Context, userID int64, orderNo string) (*model.Order, error) {
	order, err := s.orderRepository.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrForbidden
	}
	return order, nil
}

func (s *orderService) generateOrderNo(prefix string) string {
	return generatePaymentOrderNo(s.Service, prefix)
}

func generatePaymentOrderNo(service *Service, prefix string) string {
	id, err := service.sid.GenUint64()
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
