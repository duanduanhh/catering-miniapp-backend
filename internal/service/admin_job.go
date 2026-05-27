package service

import (
	"context"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type AdminJobService interface {
	List(ctx context.Context, query repository.AdminJobListQuery) ([]AdminJobItem, int64, error)
	Disable(ctx context.Context, jobID int64) error
	Enable(ctx context.Context, jobID int64) error
	Delete(ctx context.Context, jobID int64) error
}

func NewAdminJobService(
	service *Service,
	jobRepository repository.JobRepository,
) AdminJobService {
	return &adminJobService{
		Service:       service,
		jobRepository: jobRepository,
	}
}

type adminJobService struct {
	*Service
	jobRepository repository.JobRepository
}

type AdminJobItem struct {
	JobID         int64
	BizType       int
	Positions     string
	CompanyName   string
	Address       string
	SalaryMin     int
	SalaryMax     int
	Status        int
	UserID        int64
	UserName      string
	UserPhone     string
	CreateAt      time.Time
	UpdateAt      time.Time
	FirstAreaDes  string
	SecondAreaDes string
	ThirdAreaDes  string
	Description   string
	PhotoURLs     string
}

func (s *adminJobService) List(ctx context.Context, query repository.AdminJobListQuery) ([]AdminJobItem, int64, error) {
	jobs, total, err := s.jobRepository.AdminList(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	items := make([]AdminJobItem, 0, len(jobs))
	for _, job := range jobs {
		items = append(items, AdminJobItem{
			JobID:         job.ID,
			BizType:       job.BizType,
			Positions:     job.Positions,
			CompanyName:   job.CompanyName,
			Address:       job.Address,
			SalaryMin:     job.SalaryMin,
			SalaryMax:     job.SalaryMax,
			Status:        int(job.Status),
			UserID:        job.UserID,
			UserName:      job.UserName,
			UserPhone:     job.UserPhone,
			CreateAt:      job.CreateAt,
			UpdateAt:      job.UpdateAt,
			FirstAreaDes:  job.FirstAreaDes,
			SecondAreaDes: job.SecondAreaDes,
			ThirdAreaDes:  job.ThirdAreaDes,
			Description:   job.Description,
			PhotoURLs:     job.PhotoURLs,
		})
	}
	return items, total, nil
}

func (s *adminJobService) Disable(ctx context.Context, jobID int64) error {
	return s.jobRepository.AdminUpdateStatus(ctx, jobID, model.JobStatusAdminDisabled)
}

func (s *adminJobService) Enable(ctx context.Context, jobID int64) error {
	return s.jobRepository.AdminUpdateStatus(ctx, jobID, model.JobStatusActive)
}

func (s *adminJobService) Delete(ctx context.Context, jobID int64) error {
	return s.jobRepository.AdminUpdateStatus(ctx, jobID, model.JobStatusDeleted)
}
