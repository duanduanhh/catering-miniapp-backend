package repository

import (
	"context"
	"fmt"
	"time"

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
	ListTop(ctx context.Context, bizType, firstAreaID, secondAreaID, limit int) ([]*model.Job, error)
	ListFeed(ctx context.Context, bizType, firstAreaID, secondAreaID, pageNum, pageSize int) ([]*model.Job, int64, error)
	AdminList(ctx context.Context, query AdminJobListQuery) ([]*model.Job, int64, error)
	AdminUpdateStatus(ctx context.Context, jobID int64, status model.JobStatus) error
	HasPublishedByUserIDs(ctx context.Context, userIDs []int64) (map[int64]bool, error)
	ActivatePendingRent(ctx context.Context, jobID int64) error
	ListPendingRentBefore(ctx context.Context, before time.Time, limit int) ([]*model.Job, error)
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

type AdminJobListQuery struct {
	JobID    int64
	UserID   int64
	BizType  int
	Status   []int
	Keyword  string
	PageNum  int
	PageSize int
}

type JobListQuery struct {
	BizType         int // 0=不限，1=招聘，2=求职，3=招租
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
	// 招租专属筛选（仅当 BizType=3 时生效，通过 LEFT JOIN rent_detail 过滤）
	AreaSizeRange   int // 0=不限 1=<15 2=[15,30) 3=[30,50) 4=[50,100) 5=[100,200) 6=>=200
	TransferFeeFlag int // 0=不限 1=有转让费 2=无转让费
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
	// 招租专属筛选：仅当明确按 biz_type=3 查询时启用 JOIN，其他 biz_type 不影响。
	if query.BizType == 3 && (query.AreaSizeRange > 0 || query.TransferFeeFlag > 0) {
		db = db.Joins("LEFT JOIN rent_detail ON rent_detail.job_id = job.id")
		if lo, hi, ok := areaSizeRangeBounds(query.AreaSizeRange); ok {
			if hi > 0 {
				db = db.Where("rent_detail.area_size >= ? AND rent_detail.area_size < ?", lo, hi)
			} else {
				db = db.Where("rent_detail.area_size >= ?", lo)
			}
		}
		switch query.TransferFeeFlag {
		case 1:
			db = db.Where("rent_detail.transfer_fee_type > 0")
		case 2:
			db = db.Where("rent_detail.transfer_fee_type = 0")
		}
	}

	switch query.QueryType {
	case 1:
		orderClause := "CASE WHEN top_start_time IS NOT NULL AND top_end_time IS NOT NULL AND top_start_time <= NOW() AND top_end_time >= NOW() THEN 1 ELSE 0 END DESC"
		db = db.Order(orderClause).
			Order("GREATEST(COALESCE(refresh_time, job.create_at), COALESCE(paid_refresh_time, job.create_at)) DESC")
	case 2:
		if query.Longitude != 0 || query.Latitude != 0 {
			db = db.Order(fmt.Sprintf("((job.longitude-%f)*(job.longitude-%f)+(job.latitude-%f)*(job.latitude-%f)) ASC",
				query.Longitude, query.Longitude, query.Latitude, query.Latitude))
		}
	case 3:
		db = db.Order("job.create_at DESC")
	default:
		db = db.Order("job.create_at DESC")
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

func (r *jobRepository) ListTop(ctx context.Context, bizType, firstAreaID, secondAreaID, limit int) ([]*model.Job, error) {
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
	if firstAreaID > 0 {
		db = db.Where("job.first_area_id = ?", firstAreaID)
	}
	if secondAreaID > 0 {
		db = db.Where("job.second_area_id = ?", secondAreaID)
	}
	err := db.Order("GREATEST(COALESCE(job.refresh_time, job.create_at), COALESCE(job.paid_refresh_time, job.create_at)) DESC").Limit(limit).Find(&jobs).Error
	return jobs, err
}

func (r *jobRepository) ListFeed(ctx context.Context, bizType, firstAreaID, secondAreaID, pageNum, pageSize int) ([]*model.Job, int64, error) {
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
	if firstAreaID > 0 {
		db = db.Where("job.first_area_id = ?", firstAreaID)
	}
	if secondAreaID > 0 {
		db = db.Where("job.second_area_id = ?", secondAreaID)
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
	err := db.Order("GREATEST(COALESCE(job.refresh_time, job.create_at), COALESCE(job.paid_refresh_time, job.create_at)) DESC").
		Offset(offset).Limit(pageSize).Find(&jobs).Error
	return jobs, total, err
}

func (r *jobRepository) AdminList(ctx context.Context, query AdminJobListQuery) ([]*model.Job, int64, error) {
	var (
		jobs  []*model.Job
		total int64
	)
	db := r.DB(ctx).Table("job").
		Select("job.*, user.name AS user_name, user.phone AS user_phone").
		Joins("LEFT JOIN user ON job.user_id = user.id").
		Where("job.status != ?", model.JobStatusDeleted)
	if query.JobID != 0 {
		db = db.Where("job.id = ?", query.JobID)
	}
	if query.UserID != 0 {
		db = db.Where("job.user_id = ?", query.UserID)
	}
	if query.BizType > 0 {
		db = db.Where("job.biz_type = ?", query.BizType)
	}
	if len(query.Status) > 0 {
		db = db.Where("job.status IN ?", query.Status)
	}
	if query.Keyword != "" {
		db = db.Where("job.positions LIKE ? OR job.company_name LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	offset := (query.PageNum - 1) * query.PageSize
	if err := db.Order("job.create_at DESC").Offset(offset).Limit(query.PageSize).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (r *jobRepository) AdminUpdateStatus(ctx context.Context, jobID int64, status model.JobStatus) error {
	return r.DB(ctx).Model(&model.Job{}).Where("id = ?", jobID).Update("status", status).Error
}

func (r *jobRepository) HasPublishedByUserIDs(ctx context.Context, userIDs []int64) (map[int64]bool, error) {
	if len(userIDs) == 0 {
		return map[int64]bool{}, nil
	}
	var rows []struct {
		UserID int64 `gorm:"column:user_id"`
	}
	err := r.DB(ctx).Model(&model.Job{}).
		Select("user_id").
		Where("user_id IN ? AND status != ?", userIDs, model.JobStatusDeleted).
		Group("user_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]bool, len(rows))
	for _, row := range rows {
		result[row.UserID] = true
	}
	return result, nil
}

// areaSizeRangeBounds 将招租店铺面积区间枚举翻译为 [lo, hi) 上下界；
// hi=0 表示上界为 +∞（如 200㎡以上）。
func areaSizeRangeBounds(rangeEnum int) (int, int, bool) {
	switch rangeEnum {
	case 1:
		return 0, 15, true
	case 2:
		return 15, 30, true
	case 3:
		return 30, 50, true
	case 4:
		return 50, 100, true
	case 5:
		return 100, 200, true
	case 6:
		return 200, 0, true
	}
	return 0, 0, false
}
// WHERE 条件保证幂等：重复回调不会重置 refresh_time；status 已被其他动作修改的记录不受影响。
func (r *jobRepository) ActivatePendingRent(ctx context.Context, jobID int64) error {
	now := time.Now()
	return r.DB(ctx).Model(&model.Job{}).
		Where("id = ? AND status = ? AND biz_type = ?", jobID, model.JobStatusPendingPay, 3).
		Updates(map[string]interface{}{
			"status":       model.JobStatusActive,
			"refresh_time": now,
			"update_at":    now,
		}).Error
}

// ListPendingRentBefore 招租超时清理任务专用：查询指定时间之前仍处于待支付状态的招租 job。
func (r *jobRepository) ListPendingRentBefore(ctx context.Context, before time.Time, limit int) ([]*model.Job, error) {
	var jobs []*model.Job
	if limit <= 0 {
		limit = 100
	}
	err := r.DB(ctx).
		Where("biz_type = ? AND status = ? AND create_at < ?", 3, model.JobStatusPendingPay, before).
		Order("create_at ASC").
		Limit(limit).
		Find(&jobs).Error
	if err != nil {
		return nil, err
	}
	return jobs, nil
}
