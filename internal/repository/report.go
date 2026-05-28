package repository

import (
	"context"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type ReportRepository interface {
	Create(ctx context.Context, report *model.Report) error
}

func NewReportRepository(
	repository *Repository,
) ReportRepository {
	return &reportRepository{
		Repository: repository,
	}
}

type reportRepository struct {
	*Repository
}

func (r *reportRepository) Create(ctx context.Context, report *model.Report) error {
	return r.DB(ctx).Create(report).Error
}
