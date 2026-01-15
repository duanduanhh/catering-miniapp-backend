package service

import (
	"context"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type ContactVoucherHistoryService interface {
	ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.ContactVoucherHistory, int64, error)
	AdjustVoucher(ctx context.Context, userID int64, bizType model.ContactVoucherHistoryBizType, changeNum int, remark string) (int, error)
	GetUserVoucherNum(ctx context.Context, userID int64) (int, error)
}

func NewContactVoucherHistoryService(
	service *Service,
	contactVoucherHistoryRepository repository.ContactVoucherHistoryRepository,
	userRepository repository.UserRepository,
) ContactVoucherHistoryService {
	return &contactVoucherHistoryService{
		Service:                         service,
		contactVoucherHistoryRepository: contactVoucherHistoryRepository,
		userRepository:                  userRepository,
	}
}

type contactVoucherHistoryService struct {
	*Service
	contactVoucherHistoryRepository repository.ContactVoucherHistoryRepository
	userRepository                  repository.UserRepository
}

func (s *contactVoucherHistoryService) ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.ContactVoucherHistory, int64, error) {
	return s.contactVoucherHistoryRepository.ListByUser(ctx, userID, pageNum, pageSize)
}

func (s *contactVoucherHistoryService) GetUserVoucherNum(ctx context.Context, userID int64) (int, error) {
	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return user.ContactVoucherNum, nil
}

func (s *contactVoucherHistoryService) AdjustVoucher(ctx context.Context, userID int64, bizType model.ContactVoucherHistoryBizType, changeNum int, remark string) (int, error) {
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
		history := &model.ContactVoucherHistory{
			UserID:    userID,
			BizType:   bizType,
			ChangeNum: changeNum,
			LastNum:   lastNum,
			NextNum:   nextNum,
			Remark:    remark,
			CreateAt:  time.Now(),
		}
		return s.contactVoucherHistoryRepository.Create(ctx, history)
	})
	return nextNum, err
}
