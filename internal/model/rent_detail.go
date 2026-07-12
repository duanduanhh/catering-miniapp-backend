package model

import "time"

// TransferFeeType 转让费类型
type TransferFeeType int

const (
	TransferFeeNone       TransferFeeType = 0 // 无转让费
	TransferFeeFixed      TransferFeeType = 1 // 固定金额
	TransferFeeNegotiable TransferFeeType = 2 // 面议
)

// RentDetail 招租业务的扩展字段，与 job 表通过 job_id 一对一关联。
// 仅当 job.biz_type = 3（招租）时创建并使用。
type RentDetail struct {
	JobID             int64           `gorm:"primaryKey;column:job_id"`
	MonthlyRent       int             `gorm:"column:monthly_rent"`        // 店面租金（元/月），0~999999
	AreaSize          int             `gorm:"column:area_size"`           // 店面面积（㎡），0~999999
	TransferFeeType   TransferFeeType `gorm:"column:transfer_fee_type"`   // 0=无 1=固定金额 2=面议
	TransferFeeAmount int             `gorm:"column:transfer_fee_amount"` // 转让费金额（元）；面议或无=0
	TransferDesc      string          `gorm:"column:transfer_desc;size:1000"`
	CreateAt          time.Time       `gorm:"column:create_at"`
	UpdateAt          time.Time       `gorm:"column:update_at"`
}

func (m *RentDetail) TableName() string {
	return "rent_detail"
}
