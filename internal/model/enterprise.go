package model

import "time"

type EnterpriseStatus int

const (
	EnterpriseStatusPending  EnterpriseStatus = 1 // 待审核
	EnterpriseStatusVerified EnterpriseStatus = 2 // 已认证
	EnterpriseStatusRejected EnterpriseStatus = 3 // 审核驳回
	EnterpriseStatusDeleted  EnterpriseStatus = 4 // 已删除
)

type Enterprise struct {
	ID                  int64            `gorm:"primaryKey;column:id"`
	UserID              int64            `gorm:"column:user_id"`
	Name                string           `gorm:"column:name"`
	SocialCreditCode    string           `gorm:"column:social_credit_code"`
	LegalRepresentative string           `gorm:"column:legal_representative"`
	Address             string           `gorm:"column:address"`
	EstablishedDate     string           `gorm:"column:established_date"`
	BusinessPeriod      string           `gorm:"column:business_period"`
	RegisteredCapital   string           `gorm:"column:registered_capital"`
	BusinessScope       string           `gorm:"column:business_scope;type:text"`
	LicenseURL          string           `gorm:"column:license_url"`
	IsDefault           int              `gorm:"column:is_default;default:0"`
	Status              EnterpriseStatus `gorm:"column:status;default:2"`
	CreateAt            time.Time        `gorm:"column:create_at"`
	UpdateAt            time.Time        `gorm:"column:update_at"`
}

func (m *Enterprise) TableName() string {
	return "enterprise"
}
