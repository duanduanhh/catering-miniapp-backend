package service

import (
	"context"
	"time"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type ReportService interface {
	Submit(ctx context.Context, userID int64, input ReportSubmitInput) error
}

func NewReportService(
	service *Service,
	reportRepo repository.ReportRepository,
) ReportService {
	return &reportService{
		Service:    service,
		reportRepo: reportRepo,
	}
}

type reportService struct {
	*Service
	reportRepo repository.ReportRepository
}

type ReportSubmitInput struct {
	JobID       int64
	BizType     int
	Reason      int
	Description string
}

func (s *reportService) Submit(ctx context.Context, userID int64, input ReportSubmitInput) error {
	if !v1.IsValidReportReason(input.BizType, input.Reason) {
		return v1.ErrBadRequest
	}
	report := &model.Report{
		UserID:      userID,
		JobID:       input.JobID,
		BizType:     input.BizType,
		Reason:      input.Reason,
		Description: input.Description,
		Status:      model.ReportStatusPending,
		CreateAt:    time.Now(),
	}
	return s.reportRepo.Create(ctx, report)
}
