package repository

import (
	"context"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type ReportRepository interface {
	Create(ctx context.Context, report *model.Report) error
	AdminList(ctx context.Context, query AdminReportListQuery) ([]*AdminReportRow, int64, error)
}

// AdminReportListQuery 管理后台举报列表筛选条件
type AdminReportListQuery struct {
	ReportID  int64
	Status    *int
	Reason    *int
	BizType   *int
	StartTime string
	EndTime   string
	PageNum   int
	PageSize  int
}

// AdminReportRow 举报 + JOIN user 出的举报人姓名/手机号
type AdminReportRow struct {
	model.Report
	UserName  string `gorm:"column:user_name"`
	UserPhone string `gorm:"column:user_phone"`
}

func NewReportRepository(
	repository *Repository,
) ReportRepository {
	return &reportRepository{
		Repository: repository,
	}
}

type reportRepository struct {
	*Repository
}

func (r *reportRepository) Create(ctx context.Context, report *model.Report) error {
	return r.DB(ctx).Create(report).Error
}

// AdminList 管理后台举报列表，JOIN user 拿举报人信息。
func (r *reportRepository) AdminList(ctx context.Context, query AdminReportListQuery) ([]*AdminReportRow, int64, error) {
	var (
		rows  []*AdminReportRow
		total int64
	)
	db := r.DB(ctx).
		Table("report AS rp").
		Select("rp.*, u.name AS user_name, u.phone AS user_phone").
		Joins("LEFT JOIN user u ON rp.user_id = u.id")

	if query.ReportID != 0 {
		db = db.Where("rp.id = ?", query.ReportID)
	}
	if query.Status != nil {
		db = db.Where("rp.status = ?", *query.Status)
	}
	if query.Reason != nil {
		db = db.Where("rp.reason = ?", *query.Reason)
	}
	if query.BizType != nil {
		db = db.Where("rp.biz_type = ?", *query.BizType)
	}
	if query.StartTime != "" {
		db = db.Where("rp.create_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("rp.create_at <= ?", query.EndTime)
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
	if err := db.Order("rp.create_at DESC").Offset(offset).Limit(query.PageSize).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
