package repository

import (
	"context"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type ContactFeedbackRepository interface {
	Create(ctx context.Context, feedback *model.ContactFeedback) error
}

func NewContactFeedbackRepository(
	repository *Repository,
) ContactFeedbackRepository {
	return &contactFeedbackRepository{
		Repository: repository,
	}
}

type contactFeedbackRepository struct {
	*Repository
}

func (r *contactFeedbackRepository) Create(ctx context.Context, feedback *model.ContactFeedback) error {
	return r.DB(ctx).Create(feedback).Error
}
