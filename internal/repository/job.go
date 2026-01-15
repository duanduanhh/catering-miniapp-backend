package repository

import (
	"context"
	"fmt"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type JobRepository interface {
	Create(ctx context.Context, job *model.Job) error
	Update(ctx context.Context, job *model.Job) error
	GetByID(ctx context.Context, id int64) (*model.Job, error)
	List(ctx context.Context, query JobListQuery) ([]*model.Job, int64, error)
	ListByUser(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.Job, int64, error)
	ListByIDs(ctx context.Context, ids []int64) ([]*model.Job, error)
	CountByUser(ctx context.Context, userID int64, status model.JobStatus) (int64, error)
}

func NewJobRepository(
	repository *Repository,
) JobRepository {
	return &jobRepository{
		Repository: repository,
	}
}

type jobRepository struct {
	*Repository
}

type JobListQuery struct {
	QueryType       int
	Positions       string
	City            string
	SalaryMin       int
	SalaryMax       int
	BasicProtection []string
	SalaryBenefits  []string
	AttendanceLeave []string
	Longitude       float64
	Latitude        float64
	PageNum         int
	PageSize        int
}

func (r *jobRepository) Create(ctx context.Context, job *model.Job) error {
	return r.DB(ctx).Create(job).Error
}

func (r *jobRepository) Update(ctx context.Context, job *model.Job) error {
	return r.DB(ctx).Save(job).Error
}

func (r *jobRepository) GetByID(ctx context.Context, id int64) (*model.Job, error) {
	var job model.Job
	if err := r.DB(ctx).Where("id = ?", id).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *jobRepository) List(ctx context.Context, query JobListQuery) ([]*model.Job, int64, error) {
	var (
		jobs  []*model.Job
		total int64
	)
	db := r.DB(ctx).Model(&model.Job{}).Where("status = ?", model.JobStatusActive)

	if query.Positions != "" {
		db = db.Where("positions LIKE ?", "%"+query.Positions+"%")
	}
	if query.City != "" {
		db = db.Where("second_area_des LIKE ?", "%"+query.City+"%")
	}
	if query.SalaryMin > 0 {
		db = db.Where("salary_max >= ?", query.SalaryMin)
	}
	if query.SalaryMax > 0 {
		db = db.Where("salary_min <= ?", query.SalaryMax)
	}
	for _, item := range query.BasicProtection {
		db = db.Where("basic_protection LIKE ?", "%"+item+"%")
	}
	for _, item := range query.SalaryBenefits {
		db = db.Where("salary_benefits LIKE ?", "%"+item+"%")
	}
	for _, item := range query.AttendanceLeave {
		db = db.Where("attendance_leave LIKE ?", "%"+item+"%")
	}

	switch query.QueryType {
	case 1:
		orderClause := "CASE WHEN top_start_time IS NOT NULL AND top_end_time IS NOT NULL AND top_start_time <= NOW() AND top_end_time >= NOW() THEN 1 ELSE 0 END DESC"
		db = db.Order(orderClause).
			Order("refresh_time DESC")
	case 2:
		if query.Longitude != 0 || query.Latitude != 0 {
			db = db.Order(fmt.Sprintf("((longitude-%f)*(longitude-%f)+(latitude-%f)*(latitude-%f)) ASC",
				query.Longitude, query.Longitude, query.Latitude, query.Latitude))
		}
	case 3:
		db = db.Order("create_at DESC")
	default:
		db = db.Order("create_at DESC")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	offset := (query.PageNum - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (r *jobRepository) ListByUser(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.Job, int64, error) {
	var (
		jobs  []*model.Job
		total int64
	)
	db := r.DB(ctx).Model(&model.Job{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (pageNum - 1) * pageSize
	if err := db.Order("create_at DESC").Offset(offset).Limit(pageSize).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (r *jobRepository) ListByIDs(ctx context.Context, ids []int64) ([]*model.Job, error) {
	if len(ids) == 0 {
		return []*model.Job{}, nil
	}
	var jobs []*model.Job
	if err := r.DB(ctx).Where("id IN ?", ids).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *jobRepository) CountByUser(ctx context.Context, userID int64, status model.JobStatus) (int64, error) {
	var total int64
	if err := r.DB(ctx).Model(&model.Job{}).
		Where("user_id = ? AND status = ?", userID, status).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
