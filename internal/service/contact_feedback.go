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
) ContactFeedbackService {
	return &contactFeedbackService{
		Service:             service,
		contactFeedbackRepo: contactFeedbackRepo,
	}
}

type contactFeedbackService struct {
	*Service
	contactFeedbackRepo repository.ContactFeedbackRepository
}

type ContactFeedbackSubmitInput struct {
	JobID       int64
	BizType     int
	Reason      int
	Description string
}

func (s *contactFeedbackService) Submit(ctx context.Context, userID int64, input ContactFeedbackSubmitInput) error {
	feedback := &model.ContactFeedback{
		UserID:      userID,
		JobID:       input.JobID,
		BizType:     input.BizType,
		Reason:      input.Reason,
		Description: input.Description,
		Status:      model.ContactFeedbackStatusPending,
		CreateAt:    time.Now(),
	}
	return s.contactFeedbackRepo.Create(ctx, feedback)
}
