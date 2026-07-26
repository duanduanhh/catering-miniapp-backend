package service

import (
	"context"
	"strings"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

// AdminListService 聚合所有管理后台只读列表查询，避免一个 service 一个文件膨胀。
type AdminListService interface {
	Users(ctx context.Context, query repository.AdminUserListQuery) (AdminUserListResult, error)
	Enterprises(ctx context.Context, query repository.AdminEnterpriseListQuery) (AdminEnterpriseListResult, error)
	Feedbacks(ctx context.Context, query repository.AdminFeedbackListQuery) (AdminFeedbackListResult, error)
	ContactHistories(ctx context.Context, query repository.AdminContactHistoryListQuery) (AdminContactHistoryListResult, error)
	Reports(ctx context.Context, query repository.AdminReportListQuery) (AdminReportListResult, error)
	Orders(ctx context.Context, query repository.AdminOrderListQuery) (AdminOrderListResult, error)
}

func NewAdminListService(
	service *Service,
	userRepo repository.UserRepository,
	enterpriseRepo repository.EnterpriseRepository,
	feedbackRepo repository.FeedbackRepository,
	contactHistoryRepo repository.ContactHistoryRepository,
	reportRepo repository.ReportRepository,
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
) AdminListService {
	return &adminListService{
		Service:            service,
		userRepo:           userRepo,
		enterpriseRepo:     enterpriseRepo,
		feedbackRepo:       feedbackRepo,
		contactHistoryRepo: contactHistoryRepo,
		reportRepo:         reportRepo,
		orderRepo:          orderRepo,
		orderItemRepo:      orderItemRepo,
	}
}

type adminListService struct {
	*Service
	userRepo           repository.UserRepository
	enterpriseRepo     repository.EnterpriseRepository
	feedbackRepo       repository.FeedbackRepository
	contactHistoryRepo repository.ContactHistoryRepository
	reportRepo         repository.ReportRepository
	orderRepo          repository.OrderRepository
	orderItemRepo      repository.OrderItemRepository
}

// 各列表的 service 层结果（model + 必要 JOIN 字段，时间保留 time.Time，由 handler 层格式化）

type AdminUserListResult struct {
	List  []*model.User
	Total int64
}

type AdminEnterpriseListResult struct {
	List  []*repository.AdminEnterpriseRow
	Total int64
}

type AdminFeedbackListResult struct {
	List  []*repository.AdminFeedbackRow
	Total int64
}

type AdminContactHistoryListResult struct {
	List  []*repository.AdminContactHistoryRow
	Total int64
}

type AdminReportListResult struct {
	List  []*repository.AdminReportRow
	Total int64
}

type AdminOrderListResult struct {
	List  []*repository.AdminOrderRow
	Items map[int64][]*model.OrderItem
	Total int64
}

func (s *adminListService) Users(ctx context.Context, query repository.AdminUserListQuery) (AdminUserListResult, error) {
	query.Keyword = strings.TrimSpace(query.Keyword)
	list, total, err := s.userRepo.AdminList(ctx, query)
	if err != nil {
		return AdminUserListResult{}, err
	}
	return AdminUserListResult{List: list, Total: total}, nil
}

func (s *adminListService) Enterprises(ctx context.Context, query repository.AdminEnterpriseListQuery) (AdminEnterpriseListResult, error) {
	query.Keyword = strings.TrimSpace(query.Keyword)
	list, total, err := s.enterpriseRepo.AdminList(ctx, query)
	if err != nil {
		return AdminEnterpriseListResult{}, err
	}
	return AdminEnterpriseListResult{List: list, Total: total}, nil
}

func (s *adminListService) Feedbacks(ctx context.Context, query repository.AdminFeedbackListQuery) (AdminFeedbackListResult, error) {
	list, total, err := s.feedbackRepo.AdminList(ctx, query)
	if err != nil {
		return AdminFeedbackListResult{}, err
	}
	return AdminFeedbackListResult{List: list, Total: total}, nil
}

func (s *adminListService) ContactHistories(ctx context.Context, query repository.AdminContactHistoryListQuery) (AdminContactHistoryListResult, error) {
	list, total, err := s.contactHistoryRepo.AdminList(ctx, query)
	if err != nil {
		return AdminContactHistoryListResult{}, err
	}
	return AdminContactHistoryListResult{List: list, Total: total}, nil
}

func (s *adminListService) Reports(ctx context.Context, query repository.AdminReportListQuery) (AdminReportListResult, error) {
	list, total, err := s.reportRepo.AdminList(ctx, query)
	if err != nil {
		return AdminReportListResult{}, err
	}
	return AdminReportListResult{List: list, Total: total}, nil
}

func (s *adminListService) Orders(ctx context.Context, query repository.AdminOrderListQuery) (AdminOrderListResult, error) {
	query.OrderNo = strings.TrimSpace(query.OrderNo)
	query.UserKeyword = strings.TrimSpace(query.UserKeyword)
	list, total, err := s.orderRepo.AdminList(ctx, query)
	if err != nil {
		return AdminOrderListResult{}, err
	}
	orderIDs := make([]int64, 0, len(list))
	for _, order := range list {
		orderIDs = append(orderIDs, order.ID)
	}
	items, err := s.orderItemRepo.ListByOrderIDs(ctx, orderIDs)
	if err != nil {
		return AdminOrderListResult{}, err
	}
	itemsByOrderID := make(map[int64][]*model.OrderItem, len(list))
	for _, item := range items {
		itemsByOrderID[item.OrderID] = append(itemsByOrderID[item.OrderID], item)
	}
	return AdminOrderListResult{List: list, Items: itemsByOrderID, Total: total}, nil
}
