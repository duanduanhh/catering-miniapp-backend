package v1

import "github.com/go-nunu/nunu-layout-advanced/internal/model"

// ---- OCR ----

type EnterpriseOCRRequest struct {
	LicenseURL string `json:"license_url" binding:"required" example:"https://oss.example.com/license/abc.jpg"`
}

type EnterpriseOCRResponseData struct {
	Name                string `json:"name" example:"通化市东昌区助刻科技信息服务站(个体工商户)"`
	SocialCreditCode    string `json:"social_credit_code" example:"92220502MADA7W5L9P"`
	LegalRepresentative string `json:"legal_representative" example:"张占"`
	Address             string `json:"address" example:"通化市东昌区柳泉路玉皇佳园28号楼2单元601号"`
	EstablishedDate     string `json:"established_date" example:"2024-01-24"`
	BusinessPeriod      string `json:"business_period" example:"长期"`
	RegisteredCapital   string `json:"registered_capital" example:"100万人民币"`
	BusinessScope       string `json:"business_scope" example:"信息技术咨询服务;网络技术服务"`
}

// ---- Create ----

type EnterpriseCreateRequest struct {
	Name                string `json:"name" binding:"required" example:"通化市东昌区助刻科技信息服务站"`
	SocialCreditCode    string `json:"social_credit_code" binding:"required" example:"92220502MADA7W5L9P"`
	LegalRepresentative string `json:"legal_representative" binding:"required" example:"张占"`
	Address             string `json:"address" binding:"required" example:"通化市东昌区柳泉路玉皇佳园28号楼2单元601号"`
	EstablishedDate     string `json:"established_date" example:"2024-01-24"`
	BusinessPeriod      string `json:"business_period" example:"长期"`
	RegisteredCapital   string `json:"registered_capital" example:"100万人民币"`
	BusinessScope       string `json:"business_scope" example:"信息技术咨询服务;网络技术服务"`
	LicenseURL          string `json:"license_url" binding:"required" example:"https://oss.example.com/license/abc.jpg"`
	IsDefault           int    `json:"is_default" example:"1" enums:"0,1"`
}

type EnterpriseCreateResponseData struct {
	EnterpriseID int64 `json:"enterprise_id" example:"1234567890"`
}

// ---- Update ----

type EnterpriseUpdateRequest struct {
	ID                  int64   `json:"id" binding:"required" example:"1234567890"`
	Name                *string `json:"name" example:"通化市东昌区助刻科技信息服务站"`
	LegalRepresentative *string `json:"legal_representative" example:"张占"`
	Address             *string `json:"address" example:"通化市东昌区柳泉路玉皇佳园28号楼2单元601号"`
	EstablishedDate     *string `json:"established_date" example:"2024-01-24"`
	BusinessPeriod      *string `json:"business_period" example:"长期"`
	RegisteredCapital   *string `json:"registered_capital" example:"100万人民币"`
	BusinessScope       *string `json:"business_scope" example:"信息技术咨询服务;网络技术服务"`
	LicenseURL          *string `json:"license_url" example:"https://oss.example.com/license/abc.jpg"`
	IsDefault           *int    `json:"is_default" example:"1" enums:"0,1"`
}

// ---- SetDefault ----

type EnterpriseSetDefaultRequest struct {
	ID int64 `json:"id" binding:"required" example:"1234567890"`
}

// ---- My list ----

type EnterpriseMyItem struct {
	ID                  int64                  `json:"id" example:"1234567890"`
	Name                string                 `json:"name" example:"通化市东昌区助刻科技信息服务站"`
	SocialCreditCode    string                 `json:"social_credit_code" example:"92220502MADA7W5L9P"`
	LegalRepresentative string                 `json:"legal_representative" example:"张占"`
	Status              model.EnterpriseStatus `json:"status" example:"2" enums:"1,2,3,4"`
	IsDefault           int                    `json:"is_default" example:"1" enums:"0,1"`
}

type EnterpriseMyResponseData struct {
	List []EnterpriseMyItem `json:"list"`
}

// ---- Select list (for job posting) ----

type EnterpriseSelectItem struct {
	ID        int64  `json:"id" example:"1234567890"`
	Name      string `json:"name" example:"通化市东昌区助刻科技信息服务站"`
	IsDefault int    `json:"is_default" example:"1" enums:"0,1"`
}

type EnterpriseSelectListResponseData struct {
	List []EnterpriseSelectItem `json:"list"`
}

// ---- Detail (public) ----

type EnterpriseDetailRequest struct {
	ID int64 `form:"id" binding:"required" example:"1234567890"`
}

type EnterpriseDetailResponseData struct {
	ID                  int64  `json:"id" example:"1234567890"`
	Name                string `json:"name" example:"通化市东昌区助刻科技信息服务站"`
	LegalRepresentative string `json:"legal_representative" example:"张占"`
	RegisteredCapital   string `json:"registered_capital" example:"100万人民币"`
	EstablishedDate     string `json:"established_date" example:"2024-01-24"`
	BusinessPeriod      string `json:"business_period" example:"长期"`
	SocialCreditCode    string `json:"social_credit_code" example:"92220502MADA7W5L9P"`
	Address             string `json:"address" example:"通化市东昌区柳泉路玉皇佳园28号楼2单元601号"`
	BusinessScope       string `json:"business_scope" example:"信息技术咨询服务;网络技术服务"`
}
