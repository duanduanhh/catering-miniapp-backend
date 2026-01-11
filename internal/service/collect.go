package service

import (
	"context"
	"errors"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"gorm.io/gorm"
	"time"
)

type CollectService interface {
	Collect(ctx context.Context, userID, contentID int64, bizType int) error
	Cancel(ctx context.Context, userID, contentID int64, bizType int) error
	ListByUser(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.Job, int64, error)
}
func NewCollectService(
    service *Service,
    collectRepository repository.CollectRepository,
	jobRepository repository.JobRepository,
) CollectService {
	return &collectService{
		Service:        service,
		collectRepository: collectRepository,
		jobRepository: jobRepository,
	}
}

type collectService struct {
	*Service
	collectRepository repository.CollectRepository
	jobRepository     repository.JobRepository
}

func (s *collectService) Collect(ctx context.Context, userID, contentID int64, bizType int) error {
	existing, err := s.collectRepository.Get(ctx, userID, contentID, bizType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			collect := &model.Collect{
				UserID:    userID,
				ContentID: contentID,
				Type:      bizType,
				Status:    1,
				CreateAt:  time.Now(),
				UpdateAt:  time.Now(),
			}
			return s.collectRepository.Create(ctx, collect)
		}
		return err
	}
	if existing.Status != 1 {
		existing.Status = 1
		existing.UpdateAt = time.Now()
		return s.collectRepository.Update(ctx, existing)
	}
	return nil
}

func (s *collectService) Cancel(ctx context.Context, userID, contentID int64, bizType int) error {
	existing, err := s.collectRepository.Get(ctx, userID, contentID, bizType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if existing.Status != 2 {
		existing.Status = 2
		existing.UpdateAt = time.Now()
		return s.collectRepository.Update(ctx, existing)
	}
	return nil
}

func (s *collectService) ListByUser(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.Job, int64, error) {
	collects, total, err := s.collectRepository.ListByUser(ctx, userID, bizType, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	var ids []int64
	for _, item := range collects {
		ids = append(ids, item.ContentID)
	}
	jobs, err := s.jobRepository.ListByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	jobMap := make(map[int64]*model.Job, len(jobs))
	for _, job := range jobs {
		jobMap[job.ID] = job
	}
	ordered := make([]*model.Job, 0, len(ids))
	for _, id := range ids {
		if job, ok := jobMap[id]; ok {
			ordered = append(ordered, job)
		}
	}
	return ordered, total, nil
}
