package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

var ErrInvalidAdminJobLocation = errors.New("invalid admin job location")

type AdminJobService interface {
	List(ctx context.Context, query repository.AdminJobListQuery) ([]AdminJobItem, int64, error)
	Update(ctx context.Context, input AdminJobUpdateInput) error
	Disable(ctx context.Context, jobID int64) error
	Enable(ctx context.Context, jobID int64) error
	Delete(ctx context.Context, jobID int64) error
}

type AdminJobUpdateInput struct {
	JobID         int64
	Description   *string
	WorkContent   *string
	Address       *string
	AddressDetail *string
	Latitude      *float64
	Longitude     *float64
}

func NewAdminJobService(
	service *Service,
	jobRepository repository.JobRepository,
	rentDetailRepository repository.RentDetailRepository,
) AdminJobService {
	return &adminJobService{
		Service:              service,
		jobRepository:        jobRepository,
		rentDetailRepository: rentDetailRepository,
	}
}

type adminJobService struct {
	*Service
	jobRepository        repository.JobRepository
	rentDetailRepository repository.RentDetailRepository
}

type AdminJobItem struct {
	JobID             int64
	BizType           int
	Positions         string
	CompanyName       string
	ContactPersonName string
	Contact           string
	Address           string
	AddressDetail     string
	Latitude          float64
	Longitude         float64
	SalaryMin         int
	SalaryMax         int
	Status            int
	UserID            int64
	UserName          string
	UserPhone         string
	CreateAt          time.Time
	UpdateAt          time.Time
	FirstAreaDes      string
	SecondAreaDes     string
	ThirdAreaDes      string
	Description       string
	WorkContent       string
	RecruitNum        int
	PhotoURLs         string
	CloseReason       string
	CloseTime         *time.Time
	RentDetail        *model.RentDetail
}

func (s *adminJobService) List(ctx context.Context, query repository.AdminJobListQuery) ([]AdminJobItem, int64, error) {
	jobs, total, err := s.jobRepository.AdminList(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	rentJobIDs := make([]int64, 0)
	for _, job := range jobs {
		if job.BizType == 3 {
			rentJobIDs = append(rentJobIDs, job.ID)
		}
	}
	rentDetails, err := s.rentDetailRepository.GetByJobIDs(ctx, rentJobIDs)
	if err != nil {
		return nil, 0, err
	}
	items := make([]AdminJobItem, 0, len(jobs))
	for _, job := range jobs {
		items = append(items, AdminJobItem{
			JobID:             job.ID,
			BizType:           job.BizType,
			Positions:         job.Positions,
			CompanyName:       job.CompanyName,
			ContactPersonName: job.ContactPersonName,
			Contact:           job.Contact,
			Address:           job.Address,
			AddressDetail:     job.AddressDetail,
			Latitude:          job.Latitude,
			Longitude:         job.Longitude,
			SalaryMin:         job.SalaryMin,
			SalaryMax:         job.SalaryMax,
			Status:            int(job.Status),
			UserID:            job.UserID,
			UserName:          job.UserName,
			UserPhone:         job.UserPhone,
			CreateAt:          job.CreateAt,
			UpdateAt:          job.UpdateAt,
			FirstAreaDes:      job.FirstAreaDes,
			SecondAreaDes:     job.SecondAreaDes,
			ThirdAreaDes:      job.ThirdAreaDes,
			Description:       job.Description,
			WorkContent:       job.WorkContent,
			RecruitNum:        job.RecruitNum,
			PhotoURLs:         job.PhotoURLs,
			CloseReason:       job.CloseReason,
			CloseTime:         job.CloseTime,
			RentDetail:        rentDetails[job.ID],
		})
	}
	return items, total, nil
}

func (s *adminJobService) Update(ctx context.Context, input AdminJobUpdateInput) error {
	job, err := s.jobRepository.GetByID(ctx, input.JobID)
	if err != nil {
		return err
	}
	if input.Description != nil {
		job.Description = *input.Description
	}
	if input.WorkContent != nil {
		job.WorkContent = *input.WorkContent
	}
	if input.Address != nil {
		job.Address = *input.Address
	}
	if input.AddressDetail != nil {
		job.AddressDetail = *input.AddressDetail
	}
	if input.Latitude != nil {
		if math.IsNaN(*input.Latitude) || math.IsInf(*input.Latitude, 0) || *input.Latitude < -90 || *input.Latitude > 90 {
			return fmt.Errorf("%w: 纬度必须在 -90 到 90 之间", ErrInvalidAdminJobLocation)
		}
		job.Latitude = *input.Latitude
	}
	if input.Longitude != nil {
		if math.IsNaN(*input.Longitude) || math.IsInf(*input.Longitude, 0) || *input.Longitude < -180 || *input.Longitude > 180 {
			return fmt.Errorf("%w: 经度必须在 -180 到 180 之间", ErrInvalidAdminJobLocation)
		}
		job.Longitude = *input.Longitude
	}
	return s.jobRepository.Update(ctx, job)
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
