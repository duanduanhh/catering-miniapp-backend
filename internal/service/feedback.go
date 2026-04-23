package service

import (
	"context"
	"strings"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type FeedbackService interface {
	Submit(ctx context.Context, userID int64, input FeedbackSubmitInput) error
	ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.Feedback, int64, error)
}

func NewFeedbackService(
	service *Service,
	feedbackRepo repository.FeedbackRepository,
) FeedbackService {
	return &feedbackService{
		Service:      service,
		feedbackRepo: feedbackRepo,
	}
}

type feedbackService struct {
	*Service
	feedbackRepo repository.FeedbackRepository
}

type FeedbackSubmitInput struct {
	Type      model.FeedbackType
	Content   string
	PhotoURLs []string
}

func (s *feedbackService) Submit(ctx context.Context, userID int64, input FeedbackSubmitInput) error {
	feedback := &model.Feedback{
		UserID:    userID,
		Type:      input.Type,
		Content:   input.Content,
		PhotoURLs: strings.Join(input.PhotoURLs, ","),
		CreateAt:  time.Now(),
	}
	return s.feedbackRepo.Create(ctx, feedback)
}

func (s *feedbackService) ListByUser(ctx context.Context, userID int64, pageNum, pageSize int) ([]*model.Feedback, int64, error) {
	return s.feedbackRepo.ListByUser(ctx, userID, pageNum, pageSize)
}
