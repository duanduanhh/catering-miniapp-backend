package service

import (
	"context"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type CallbackHistoryService interface {
	Create(ctx context.Context, input CallbackHistoryCreateInput) error
}

type CallbackHistoryCreateInput struct {
	UserID           int64
	PurposeID        int64
	PurposeType      int
	PurposeUserID    int64
	PurposeUserName  string
	PurposeUserPhone string
}

func NewCallbackHistoryService(
	service *Service,
	callbackHistoryRepository repository.CallbackHistoryRepository,
) CallbackHistoryService {
	return &callbackHistoryService{
		Service:                   service,
		callbackHistoryRepository: callbackHistoryRepository,
	}
}

type callbackHistoryService struct {
	*Service
	callbackHistoryRepository repository.CallbackHistoryRepository
}

func (s *callbackHistoryService) Create(ctx context.Context, input CallbackHistoryCreateInput) error {
	history := &model.CallbackHistory{
		UserID:           input.UserID,
		PurposeID:        input.PurposeID,
		PurposeType:      input.PurposeType,
		PurposeUserID:    input.PurposeUserID,
		PurposeUserName:  input.PurposeUserName,
		PurposeUserPhone: input.PurposeUserPhone,
		CreateAt:         time.Now(),
	}
	return s.callbackHistoryRepository.Create(ctx, history)
}
