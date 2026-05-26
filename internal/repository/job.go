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
	ListByUser(ctx context.Context, userID int64, bizType int, status []int, pageNum, pageSize int) ([]*model.Job, int64, error)
	ListByIDs(ctx context.Context, ids []int64) ([]*model.Job, error)
	CountByUser(ctx context.Context, userID int64, bizType int, status model.JobStatus) (int64, error)
	ListTop(ctx context.Context, bizType, limit int) ([]*model.Job, error)
	ListFeed(ctx context.Context, bizType, pageNum, pageSize int) ([]*model.Job, int64, error)
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
	BizType         int // 0=不限，1=招聘，2=求职
	QueryType       int
	Positions       string
	FirstAreaID     int
	SecondAreaID    int
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
	if err := r.DB(ctx).Table("job").
		Select("job.*, user.avatar, COALESCE(enterprise.name, '') AS enterprise_name").
		Joins("LEFT JOIN user ON job.user_id = user.id").
		Joins("LEFT JOIN enterprise ON job.enterprise_id = enterprise.id AND enterprise.status = ?", model.EnterpriseStatusVerified).
		Where("job.id = ?", id).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *jobRepository) List(ctx context.Context, query JobListQuery) ([]*model.Job, int64, error) {
	var (
		jobs  []*model.Job
		total int64
	)
	db := r.DB(ctx).Table("job").
		Select("job.*, user.avatar, COALESCE(enterprise.name, '') AS enterprise_name").
		Joins("LEFT JOIN user ON job.user_id = user.id").
		Joins("LEFT JOIN enterprise ON job.enterprise_id = enterprise.id AND enterprise.status = ?", model.EnterpriseStatusVerified).
		Where("job.status = ?", model.JobStatusActive)

	if query.BizType > 0 {
		db = db.Where("job.biz_type = ?", query.BizType)
	}
	if query.FirstAreaID > 0 {
		db = db.Where("job.first_area_id = ?", query.FirstAreaID)
	}
	if query.SecondAreaID > 0 {
		db = db.Where("job.second_area_id = ?", query.SecondAreaID)
	}
	if query.SalaryMin > 0 {
		db = db.Where("salary_max >= ?", query.SalaryMin)
	}
	if query.SalaryMax > 0 {
		db = db.Where("salary_min <= ?", query.SalaryMax)
	}
	if query.Positions != "" {
		db = db.Where("positions LIKE ?", "%"+query.Positions+"%")
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
			db = db.Order(fmt.Sprintf("((job.longitude-%f)*(job.longitude-%f)+(job.latitude-%f)*(job.latitude-%f)) ASC",
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

func (r *jobRepository) ListByUser(ctx context.Context, userID int64, bizType int, status []int, pageNum, pageSize int) ([]*model.Job, int64, error) {
	var (
		jobs  []*model.Job
		total int64
	)
	db := r.DB(ctx).Model(&model.Job{}).Where("user_id = ? AND status != ?", userID, model.JobStatusDeleted)
	if bizType > 0 {
		db = db.Where("biz_type = ?", bizType)
	}
	// 当 status 有值时，按数组查询
	if len(status) > 0 {
		db = db.Where("status IN ?", status)
	}
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
	if err := r.DB(ctx).
		Select("job.*, user.avatar, COALESCE(enterprise.name, '') AS enterprise_name").
		Joins("LEFT JOIN user ON job.user_id = user.id").
		Joins("LEFT JOIN enterprise ON job.enterprise_id = enterprise.id AND enterprise.status = ?", model.EnterpriseStatusVerified).
		Where("job.id IN ?", ids).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *jobRepository) CountByUser(ctx context.Context, userID int64, bizType int, status model.JobStatus) (int64, error) {
	var total int64
	db := r.DB(ctx).Model(&model.Job{}).Where("user_id = ? AND status = ?", userID, status)
	if bizType > 0 {
		db = db.Where("biz_type = ?", bizType)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *jobRepository) ListTop(ctx context.Context, bizType, limit int) ([]*model.Job, error) {
	var jobs []*model.Job
	if limit <= 0 {
		limit = 5
	}
	db := r.DB(ctx).Table("job").
		Select("job.*, user.avatar, COALESCE(enterprise.name, '') AS enterprise_name").
		Joins("LEFT JOIN user ON job.user_id = user.id").
		Joins("LEFT JOIN enterprise ON job.enterprise_id = enterprise.id AND enterprise.status = ?", model.EnterpriseStatusVerified).
		Where("job.status = ? AND job.top_start_time IS NOT NULL AND job.top_end_time IS NOT NULL AND job.top_start_time <= NOW() AND job.top_end_time >= NOW()", model.JobStatusActive)
	if bizType > 0 {
		db = db.Where("job.biz_type = ?", bizType)
	}
	err := db.Order("job.top_end_time DESC").Limit(limit).Find(&jobs).Error
	return jobs, err
}

func (r *jobRepository) ListFeed(ctx context.Context, bizType, pageNum, pageSize int) ([]*model.Job, int64, error) {
	var (
		jobs  []*model.Job
		total int64
	)
	db := r.DB(ctx).Table("job").
		Select("job.*, user.avatar, COALESCE(enterprise.name, '') AS enterprise_name").
		Joins("LEFT JOIN user ON job.user_id = user.id").
		Joins("LEFT JOIN enterprise ON job.enterprise_id = enterprise.id AND enterprise.status = ?", model.EnterpriseStatusVerified).
		Where("job.status = ? AND NOT (job.top_start_time IS NOT NULL AND job.top_end_time IS NOT NULL AND job.top_start_time <= NOW() AND job.top_end_time >= NOW())", model.JobStatusActive)
	if bizType > 0 {
		db = db.Where("job.biz_type = ?", bizType)
	}

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
	err := db.Order("COALESCE(job.refresh_time, job.create_at) DESC").
		Offset(offset).Limit(pageSize).Find(&jobs).Error
	return jobs, total, err
}
