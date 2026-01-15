package service

import (
	"context"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type JobService interface {
	Create(ctx context.Context, userID int64, input JobCreateInput) (*model.Job, error)
	Update(ctx context.Context, userID int64, input JobUpdateInput) error
	Refresh(ctx context.Context, userID, jobID int64) error
	Close(ctx context.Context, userID, jobID int64) error
	GetByID(ctx context.Context, jobID int64) (*model.Job, error)
	List(ctx context.Context, query repository.JobListQuery) ([]*model.Job, int64, error)
	ListByUser(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.Job, int64, error)
}

func NewJobService(
	service *Service,
	jobRepository repository.JobRepository,
) JobService {
	return &jobService{
		Service:       service,
		jobRepository: jobRepository,
	}
}

type jobService struct {
	*Service
	jobRepository repository.JobRepository
}

type JobCreateInput struct {
	Positions          string
	CompanyName        string
	Longitude          float64
	Latitude           float64
	Address            string
	Contact            string
	ContanctPersonName string
	Description        string
	PhotoURLs          string
	FirstAreaID        int
	FirstAreaDes       string
	SecondAreaID       int
	SecondAreaDes      string
	ThirdAreaID        int
	ThirdAreaDes       string
	FourAreaID         int
	FourAreaDes        string
	SalaryMin          int
	SalaryMax          int
	BasicProtection    string
	SalaryBenefits     string
	AttendanceLeave    string
}

type JobUpdateInput struct {
	ID              int64
	Positions       *string
	CompanyName     *string
	Longitude       *float64
	Latitude        *float64
	Address         *string
	Contact         *string
	Description     *string
	PhotoURLs       *string
	FirstAreaID     *int
	FirstAreaDes    *string
	SecondAreaID    *int
	SecondAreaDes   *string
	ThirdAreaID     *int
	ThirdAreaDes    *string
	FourAreaID      *int
	FourAreaDes     *string
	SalaryMin       *int
	SalaryMax       *int
	BasicProtection *string
	SalaryBenefits  *string
	AttendanceLeave *string
}

func (s *jobService) Create(ctx context.Context, userID int64, input JobCreateInput) (*model.Job, error) {
	total, err := s.jobRepository.CountByUser(ctx, userID, model.JobStatusActive)
	if err != nil {
		return nil, err
	}
	if total >= 5 {
		return nil, ErrJobLimitExceeded
	}
	now := time.Now()
	job := &model.Job{
		UserID:            userID,
		Positions:         input.Positions,
		CompanyName:       input.CompanyName,
		Longitude:         input.Longitude,
		Latitude:          input.Latitude,
		Address:           input.Address,
		Contact:           input.Contact,
		ContactPersonName: input.ContanctPersonName,
		Description:       input.Description,
		PhotoURLs:         input.PhotoURLs,
		Status:            model.JobStatusActive,
		FirstAreaID:       input.FirstAreaID,
		FirstAreaDes:      input.FirstAreaDes,
		SecondAreaID:      input.SecondAreaID,
		SecondAreaDes:     input.SecondAreaDes,
		ThirdAreaID:       input.ThirdAreaID,
		ThirdAreaDes:      input.ThirdAreaDes,
		FourAreaID:        input.FourAreaID,
		FourAreaDes:       input.FourAreaDes,
		SalaryMin:         input.SalaryMin,
		SalaryMax:         input.SalaryMax,
		BasicProtection:   input.BasicProtection,
		SalaryBenefits:    input.SalaryBenefits,
		AttendanceLeave:   input.AttendanceLeave,
		CreateAt:          now,
		UpdateAt:          now,
	}
	if err := s.jobRepository.Create(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *jobService) Update(ctx context.Context, userID int64, input JobUpdateInput) error {
	job, err := s.jobRepository.GetByID(ctx, input.ID)
	if err != nil {
		return err
	}
	if job.UserID != userID {
		return ErrForbidden
	}
	if input.Positions != nil {
		job.Positions = *input.Positions
	}
	if input.CompanyName != nil {
		job.CompanyName = *input.CompanyName
	}
	if input.Longitude != nil {
		job.Longitude = *input.Longitude
	}
	if input.Latitude != nil {
		job.Latitude = *input.Latitude
	}
	if input.Address != nil {
		job.Address = *input.Address
	}
	if input.Contact != nil {
		job.Contact = *input.Contact
	}
	if input.Description != nil {
		job.Description = *input.Description
	}
	if input.PhotoURLs != nil {
		job.PhotoURLs = *input.PhotoURLs
	}
	if input.FirstAreaID != nil {
		job.FirstAreaID = *input.FirstAreaID
	}
	if input.FirstAreaDes != nil {
		job.FirstAreaDes = *input.FirstAreaDes
	}
	if input.SecondAreaID != nil {
		job.SecondAreaID = *input.SecondAreaID
	}
	if input.SecondAreaDes != nil {
		job.SecondAreaDes = *input.SecondAreaDes
	}
	if input.ThirdAreaID != nil {
		job.ThirdAreaID = *input.ThirdAreaID
	}
	if input.ThirdAreaDes != nil {
		job.ThirdAreaDes = *input.ThirdAreaDes
	}
	if input.FourAreaID != nil {
		job.FourAreaID = *input.FourAreaID
	}
	if input.FourAreaDes != nil {
		job.FourAreaDes = *input.FourAreaDes
	}
	if input.SalaryMin != nil {
		job.SalaryMin = *input.SalaryMin
	}
	if input.SalaryMax != nil {
		job.SalaryMax = *input.SalaryMax
	}
	if input.BasicProtection != nil {
		job.BasicProtection = *input.BasicProtection
	}
	if input.SalaryBenefits != nil {
		job.SalaryBenefits = *input.SalaryBenefits
	}
	if input.AttendanceLeave != nil {
		job.AttendanceLeave = *input.AttendanceLeave
	}
	return s.jobRepository.Update(ctx, job)
}

func (s *jobService) Refresh(ctx context.Context, userID, jobID int64) error {
	job, err := s.jobRepository.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job.UserID != userID {
		return ErrForbidden
	}
	now := time.Now()
	job.RefreshTime = &now
	return s.jobRepository.Update(ctx, job)
}

func (s *jobService) Close(ctx context.Context, userID, jobID int64) error {
	job, err := s.jobRepository.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job.UserID != userID {
		return ErrForbidden
	}
	job.Status = model.JobStatusUserClosed
	return s.jobRepository.Update(ctx, job)
}

func (s *jobService) GetByID(ctx context.Context, jobID int64) (*model.Job, error) {
	return s.jobRepository.GetByID(ctx, jobID)
}

func (s *jobService) List(ctx context.Context, query repository.JobListQuery) ([]*model.Job, int64, error) {
	return s.jobRepository.List(ctx, query)
}

func (s *jobService) ListByUser(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.Job, int64, error) {
	return s.jobRepository.ListByUser(ctx, userID, bizType, pageNum, pageSize)
}
