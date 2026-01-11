package service

import (
	"context"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"time"
)

type CostHistoryService interface {
	ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.CostHistory, int64, error)
	AdjustVoucher(ctx context.Context, userID int64, bizType int, changeNum int, remark string) (int, error)
	GetUserVoucherNum(ctx context.Context, userID int64) (int, error)
}
func NewCostHistoryService(
    service *Service,
    costHistoryRepository repository.CostHistoryRepository,
	userRepository repository.UserRepository,
) CostHistoryService {
	return &costHistoryService{
		Service:        service,
		costHistoryRepository: costHistoryRepository,
		userRepository: userRepository,
	}
}

type costHistoryService struct {
	*Service
	costHistoryRepository repository.CostHistoryRepository
	userRepository        repository.UserRepository
}

func (s *costHistoryService) ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.CostHistory, int64, error) {
	return s.costHistoryRepository.ListByUser(ctx, userID, pageNum, pageSize)
}

func (s *costHistoryService) GetUserVoucherNum(ctx context.Context, userID int64) (int, error) {
	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return user.ContactVoucherNum, nil
}

func (s *costHistoryService) AdjustVoucher(ctx context.Context, userID int64, bizType int, changeNum int, remark string) (int, error) {
	var nextNum int
	err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		user, err := s.userRepository.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		lastNum := user.ContactVoucherNum
		nextNum = lastNum + changeNum
		if nextNum < 0 {
			return ErrInsufficientVoucher
		}
		user.ContactVoucherNum = nextNum
		user.UpdateAt = time.Now()
		if err := s.userRepository.Update(ctx, user); err != nil {
			return err
		}
		history := &model.CostHistory{
			UserID:    userID,
			BizType:   bizType,
			ChangeNum: changeNum,
			LastNum:   lastNum,
			NextNum:   nextNum,
			Remark:    remark,
			CreateAt:  time.Now(),
		}
		return s.costHistoryRepository.Create(ctx, history)
	})
	return nextNum, err
}
