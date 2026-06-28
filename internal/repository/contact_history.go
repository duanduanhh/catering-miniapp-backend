package repository

import (
	"context"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type ContactHistoryRepository interface {
	Create(ctx context.Context, history *model.ContactHistory) error
	ListOut(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.ContactHistory, int64, error)
	ListIn(ctx context.Context, purposeUserID int64, bizType int, pageNum, pageSize int) ([]*model.ContactHistory, int64, error)
	DeleteOut(ctx context.Context, userID, purposeID int64) error
	DeleteIn(ctx context.Context, purposeUserID, purposeID int64) error
	ExistsByUserAndJob(ctx context.Context, userID, jobID int64, purposeType int) (bool, error)
	GetLatestIDByUserAndJob(ctx context.Context, userID, jobID int64, purposeType int) (int64, error)
	AdminList(ctx context.Context, query AdminContactHistoryListQuery) ([]*AdminContactHistoryRow, int64, error)
}

func NewContactHistoryRepository(
	repository *Repository,
) ContactHistoryRepository {
	return &contactHistoryRepository{
		Repository: repository,
	}
}

type contactHistoryRepository struct {
	*Repository
}

// AdminContactHistoryListQuery 管理后台联系记录筛选条件
type AdminContactHistoryListQuery struct {
	ID            int64
	UserID        int64
	PurposeUserID int64
	JobID         int64
	StartTime     string // YYYY-MM-DD HH:mm:ss，含
	EndTime       string // YYYY-MM-DD HH:mm:ss，不含或含均可（这里用 <=）
	PageNum       int
	PageSize      int
}

// AdminContactHistoryRow 管理后台一条联系记录 + JOIN 出的发起人信息 + 反馈聚合
// 字段以 GORM column 显式映射，便于跨多表 JOIN 直接 Scan。
type AdminContactHistoryRow struct {
	ID                 int64     `gorm:"column:id"`
	UserID             int64     `gorm:"column:user_id"`
	UserName           string    `gorm:"column:user_name"`
	UserPhone          string    `gorm:"column:user_phone"`
	PurposeID          int64     `gorm:"column:purpose_id"`
	PurposeType        int       `gorm:"column:purpose_type"`
	PurposeUserID      int64     `gorm:"column:purpose_user_id"`
	PurposeUserName    string    `gorm:"column:purpose_user_name"`
	PurposeUserPhone   string    `gorm:"column:purpose_user_phone"`
	UserDeleted        int       `gorm:"column:user_deleted"`
	PurposeUserDeleted int       `gorm:"column:purpose_user_deleted"`
	CreateAt           time.Time `gorm:"column:create_at"`
	FeedbackID         int64     `gorm:"column:feedback_id"`
	FeedbackReason     int       `gorm:"column:feedback_reason"`
	FeedbackStatus     int       `gorm:"column:feedback_status"`
	FeedbackCreateAt   *time.Time `gorm:"column:feedback_create_at"`
}

func (r *contactHistoryRepository) Create(ctx context.Context, history *model.ContactHistory) error {
	return r.DB(ctx).Create(history).Error
}

func (r *contactHistoryRepository) ListOut(ctx context.Context, userID int64, bizType int, pageNum, pageSize int) ([]*model.ContactHistory, int64, error) {
	var (
		histories []*model.ContactHistory
		total     int64
	)
	db := r.DB(ctx).Model(&model.ContactHistory{}).Where("user_id = ? AND user_deleted = 0", userID)
	if bizType > 0 {
		db = db.Where("purpose_type = ?", bizType)
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
	if err := db.Order("create_at DESC").Offset(offset).Limit(pageSize).Find(&histories).Error; err != nil {
		return nil, 0, err
	}
	return histories, total, nil
}

func (r *contactHistoryRepository) ListIn(ctx context.Context, purposeUserID int64, bizType int, pageNum, pageSize int) ([]*model.ContactHistory, int64, error) {
	var (
		histories []*model.ContactHistory
		total     int64
	)
	db := r.DB(ctx).Model(&model.ContactHistory{}).Where("purpose_user_id = ? AND purpose_user_deleted = 0", purposeUserID)
	if bizType > 0 {
		db = db.Where("purpose_type = ?", bizType)
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
	if err := db.Order("create_at DESC").Offset(offset).Limit(pageSize).Find(&histories).Error; err != nil {
		return nil, 0, err
	}
	return histories, total, nil
}

func (r *contactHistoryRepository) ExistsByUserAndJob(ctx context.Context, userID, jobID int64, purposeType int) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.ContactHistory{}).
		Where("user_id = ? AND purpose_id = ? AND purpose_type = ? AND user_deleted = 0", userID, jobID, purposeType).
		Count(&count).Error
	return count > 0, err
}

// DeleteOut 删除"我联系的"记录（软删除）
func (r *contactHistoryRepository) DeleteOut(ctx context.Context, userID, purposeID int64) error {
	return r.DB(ctx).Model(&model.ContactHistory{}).Where("user_id = ? AND purpose_id = ?", userID, purposeID).Update("user_deleted", 1).Error
}

// DeleteIn 删除"联系我的"记录（软删除）
func (r *contactHistoryRepository) DeleteIn(ctx context.Context, purposeUserID, purposeID int64) error {
	return r.DB(ctx).Model(&model.ContactHistory{}).Where("purpose_user_id = ? AND purpose_id = ?", purposeUserID, purposeID).Update("purpose_user_deleted", 1).Error
}

// GetLatestIDByUserAndJob 取该用户对该岗位最近一次拨打记录的 ID，用于 contact_feedback 提交时回填外键。
// 找不到时返回 0。
func (r *contactHistoryRepository) GetLatestIDByUserAndJob(ctx context.Context, userID, jobID int64, purposeType int) (int64, error) {
	var id int64
	err := r.DB(ctx).Model(&model.ContactHistory{}).
		Where("user_id = ? AND purpose_id = ? AND purpose_type = ?", userID, jobID, purposeType).
		Order("create_at DESC").
		Limit(1).
		Pluck("id", &id).Error
	return id, err
}

// AdminList 管理后台联系记录列表，JOIN 用户表拿发起人姓名手机号，
// LEFT JOIN contact_feedback 取最新一条反馈聚合。
func (r *contactHistoryRepository) AdminList(ctx context.Context, query AdminContactHistoryListQuery) ([]*AdminContactHistoryRow, int64, error) {
	var (
		rows  []*AdminContactHistoryRow
		total int64
	)
	// 子查询：取每条 contact_history 对应的最新一条 contact_feedback
	feedbackSub := `(
		SELECT cf.contact_history_id,
		       cf.id          AS feedback_id,
		       cf.reason      AS feedback_reason,
		       cf.status      AS feedback_status,
		       cf.create_at   AS feedback_create_at
		FROM contact_feedback cf
		INNER JOIN (
			SELECT contact_history_id, MAX(create_at) AS max_create
			FROM contact_feedback
			WHERE contact_history_id > 0
			GROUP BY contact_history_id
		) latest ON latest.contact_history_id = cf.contact_history_id AND latest.max_create = cf.create_at
	)`

	db := r.DB(ctx).
		Table("contact_history AS ch").
		Select(`ch.id,
		        ch.user_id, u.name AS user_name, u.phone AS user_phone,
		        ch.purpose_id, ch.purpose_type,
		        ch.purpose_user_id, ch.purpose_user_name, ch.purpose_user_phone,
		        ch.user_deleted, ch.purpose_user_deleted, ch.create_at,
		        COALESCE(fb.feedback_id, 0)        AS feedback_id,
		        COALESCE(fb.feedback_reason, 0)    AS feedback_reason,
		        COALESCE(fb.feedback_status, 0)    AS feedback_status,
		        fb.feedback_create_at`).
		Joins("LEFT JOIN user u ON ch.user_id = u.id").
		Joins("LEFT JOIN " + feedbackSub + " fb ON fb.contact_history_id = ch.id")

	if query.ID != 0 {
		db = db.Where("ch.id = ?", query.ID)
	}
	if query.UserID > 0 {
		db = db.Where("ch.user_id = ?", query.UserID)
	}
	if query.PurposeUserID > 0 {
		db = db.Where("ch.purpose_user_id = ?", query.PurposeUserID)
	}
	if query.JobID > 0 {
		db = db.Where("ch.purpose_id = ?", query.JobID)
	}
	if query.StartTime != "" {
		db = db.Where("ch.create_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("ch.create_at <= ?", query.EndTime)
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
	if err := db.Order("ch.create_at DESC").Offset(offset).Limit(query.PageSize).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
