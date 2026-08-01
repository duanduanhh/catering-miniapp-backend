package model

import "time"

type PaymentPackageStatus int

const (
	PaymentPackageStatusDraft       PaymentPackageStatus = 1
	PaymentPackageStatusPublished   PaymentPackageStatus = 2
	PaymentPackageStatusUnpublished PaymentPackageStatus = 3
)

// PaymentBenefitConfig 是一个 SKU 的完整发放权益。
// 当前四类商品只需要一个主权益，并可选赠送联系券，因此无需拆分权益明细表。
type PaymentBenefitConfig struct {
	ContactVouchers     int `json:"contact_vouchers,omitempty"`
	TopHours            int `json:"top_hours,omitempty"`
	RefreshTimes        int `json:"refresh_times,omitempty"`
	RentPublishTimes    int `json:"rent_publish_times,omitempty"`
	GiftContactVouchers int `json:"gift_contact_vouchers,omitempty"`
}

// PaymentSaleRule 是 SKU 的购买资格和限购规则。
// Audience: all=全部用户，platform_new=平台新用户，product_new=当前商品首购用户。
type PaymentSaleRule struct {
	Audience           string `json:"audience"`
	MaxPurchasePerUser int    `json:"max_purchase_per_user"`
}

type PaymentPackageAction int

const (
	PaymentPackageActionCreate    PaymentPackageAction = 1
	PaymentPackageActionUpdate    PaymentPackageAction = 2
	PaymentPackageActionPublish   PaymentPackageAction = 3
	PaymentPackageActionUnpublish PaymentPackageAction = 4
	PaymentPackageActionDelete    PaymentPackageAction = 5
)

type PaymentPackage struct {
	ID                 int64                `gorm:"primaryKey;column:id"`
	ProductID          int64                `gorm:"column:product_id;not null;default:0;index:idx_payment_sku_product_id"`
	SKUCode            string               `gorm:"column:sku_code;size:64;uniqueIndex:uk_payment_sku_code"`
	Name               string               `gorm:"column:name;size:100"`
	Subtitle           string               `gorm:"column:subtitle;size:200"`
	Badge              string               `gorm:"column:badge;size:50"`
	PriceCents         int64                `gorm:"column:price_cents"`
	OriginalPriceCents int64                `gorm:"column:original_price_cents"`
	BenefitConfigJSON  string               `gorm:"column:benefit_config;type:json"`
	SaleRuleJSON       string               `gorm:"column:sale_rule;type:json"`
	Status             PaymentPackageStatus `gorm:"column:status;index:idx_payment_sku_list,priority:2"`
	Sort               int                  `gorm:"column:sort;index:idx_payment_sku_list,priority:3"`
	Version            int                  `gorm:"column:version"`
	CreatedBy          string               `gorm:"column:created_by;size:64"`
	UpdatedBy          string               `gorm:"column:updated_by;size:64"`
	CreateAt           time.Time            `gorm:"column:create_at"`
	UpdateAt           time.Time            `gorm:"column:update_at"`
	DeletedAt          *time.Time           `gorm:"column:deleted_at;index"`
}

func (m *PaymentPackage) TableName() string {
	return "payment_sku"
}

type PaymentPackageChangeLog struct {
	ID             int64                `gorm:"primaryKey;column:id"`
	SKUID          int64                `gorm:"column:sku_id;index:idx_payment_sku_log,priority:1"`
	SKUVersion     int                  `gorm:"column:sku_version"`
	Action         PaymentPackageAction `gorm:"column:action"`
	BeforeSnapshot string               `gorm:"column:before_snapshot;type:json"`
	AfterSnapshot  string               `gorm:"column:after_snapshot;type:json"`
	ChangeReason   string               `gorm:"column:change_reason;size:500"`
	Operator       string               `gorm:"column:operator;size:64"`
	CreateAt       time.Time            `gorm:"column:create_at;index:idx_payment_sku_log,priority:2"`
}

func (m *PaymentPackageChangeLog) TableName() string {
	return "payment_sku_change_log"
}
