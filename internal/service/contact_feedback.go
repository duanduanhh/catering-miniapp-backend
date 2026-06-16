package service

import (
	"context"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type ContactFeedbackService interface {
	Submit(ctx context.Context, userID int64, input ContactFeedbackSubmitInput) error
}

func NewContactFeedbackService(
	service *Service,
	contactFeedbackRepo repository.ContactFeedbackRepository,
	contactHistoryRepo repository.ContactHistoryRepository,
) ContactFeedbackService {
	return &contactFeedbackService{
		Service:             service,
		contactFeedbackRepo: contactFeedbackRepo,
		contactHistoryRepo:  contactHistoryRepo,
	}
}

type contactFeedbackService struct {
	*Service
	contactFeedbackRepo repository.ContactFeedbackRepository
	contactHistoryRepo  repository.ContactHistoryRepository
}

type ContactFeedbackSubmitInput struct {
	JobID       int64
	BizType     int
	Reason      int
	Description string
}

func (s *contactFeedbackService) Submit(ctx context.Context, userID int64, input ContactFeedbackSubmitInput) error {
	// 反查最近一次拨打记录的 ID 作为外键。找不到就用 0（兼容历史 / 异常路径）。
	historyID, err := s.contactHistoryRepo.GetLatestIDByUserAndJob(ctx, userID, input.JobID, input.BizType)
	if err != nil {
		historyID = 0
	}
	feedback := &model.ContactFeedback{
		UserID:           userID,
		ContactHistoryID: historyID,
		JobID:            input.JobID,
		BizType:          input.BizType,
		Reason:           input.Reason,
		Description:      input.Description,
		Status:           model.ContactFeedbackStatusPending,
		CreateAt:         time.Now(),
	}
	return s.contactFeedbackRepo.Create(ctx, feedback)
}
