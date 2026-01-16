package service

import (
	"context"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type ContactHistoryService interface {
	Create(ctx context.Context, input ContactHistoryCreateInput) (*model.ContactHistory, error)
	ListOut(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]ContactHistoryItem, int64, error)
	ListIn(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]ContactHistoryItem, int64, error)
}

func NewContactHistoryService(
	service *Service,
	contactHistoryRepository repository.ContactHistoryRepository,
	jobRepository repository.JobRepository,
	userRepository repository.UserRepository,
) ContactHistoryService {
	return &contactHistoryService{
		Service:                  service,
		contactHistoryRepository: contactHistoryRepository,
		jobRepository:            jobRepository,
		userRepository:           userRepository,
	}
}

type contactHistoryService struct {
	*Service
	contactHistoryRepository repository.ContactHistoryRepository
	jobRepository            repository.JobRepository
	userRepository           repository.UserRepository
}

type ContactHistoryCreateInput struct {
	UserID           int64
	PurposeID        int64
	PurposeType      int
	PurposeUserID    int64
	PurposeUserName  string
	PurposeUserPhone string
}

type ContactHistoryItem struct {
	ID               int64
	Positions        string
	Address          string
	PurposeUserID    int64
	PurposeUserName  string
	PurposeUserPhone string
	CreateAt         time.Time
}

func (s *contactHistoryService) Create(ctx context.Context, input ContactHistoryCreateInput) (*model.ContactHistory, error) {
	history := &model.ContactHistory{
		UserID:           input.UserID,
		PurposeID:        input.PurposeID,
		PurposeType:      input.PurposeType,
		PurposeUserID:    input.PurposeUserID,
		PurposeUserName:  input.PurposeUserName,
		PurposeUserPhone: input.PurposeUserPhone,
		CreateAt:         time.Now(),
		UpdateAt:         time.Now(),
	}
	if err := s.contactHistoryRepository.Create(ctx, history); err != nil {
		return nil, err
	}
	return history, nil
}

func (s *contactHistoryService) ListOut(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]ContactHistoryItem, int64, error) {
	histories, total, err := s.contactHistoryRepository.ListOut(ctx, userID, bizType, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return s.buildHistoryItems(ctx, histories), total, nil
}

func (s *contactHistoryService) ListIn(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]ContactHistoryItem, int64, error) {
	histories, total, err := s.contactHistoryRepository.ListIn(ctx, userID, bizType, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return s.buildHistoryItems(ctx, histories), total, nil
}

func (s *contactHistoryService) buildHistoryItems(ctx context.Context, histories []*model.ContactHistory) []ContactHistoryItem {
	jobIDs := make([]int64, 0, len(histories))
	userIDs := make([]int64, 0, len(histories))
	for _, item := range histories {
		jobIDs = append(jobIDs, item.PurposeID)
		userIDs = append(userIDs, item.PurposeUserID)
	}
	jobs, _ := s.jobRepository.ListByIDs(ctx, jobIDs)

	jobMap := make(map[int64]*model.Job, len(jobs))
	for _, job := range jobs {
		jobMap[job.ID] = job
	}

	items := make([]ContactHistoryItem, 0, len(histories))
	for _, history := range histories {
		item := ContactHistoryItem{
			ID:               history.ID,
			CreateAt:         history.CreateAt,
			PurposeUserID:    history.PurposeID,
			PurposeUserName:  history.PurposeUserName,
			PurposeUserPhone: history.PurposeUserPhone,
		}
		if job, ok := jobMap[history.PurposeID]; ok {
			item.Positions = job.Positions
			item.Address = job.Address
		}
		items = append(items, item)
	}
	return items
}
