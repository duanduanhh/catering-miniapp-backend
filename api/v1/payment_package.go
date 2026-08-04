package v1

// PaymentBenefitConfig 是 SKU 的权益配置。主权益仅填写所属产品对应的字段；岗位置顶可额外填写赠送联系券。
type PaymentBenefitConfig struct {
	ContactVouchers     int `json:"contact_vouchers,omitempty"`
	TopHours            int `json:"top_hours,omitempty"`
	RefreshTimes        int `json:"refresh_times,omitempty"`
	RentPublishTimes    int `json:"rent_publish_times,omitempty"`
	GiftContactVouchers int `json:"gift_contact_vouchers,omitempty"`
}

// PaymentSaleRule 是管理后台配置的 SKU 购买规则。
// audience：all=全部用户，platform_new=平台新用户，product_new=当前产品首购用户。
// max_purchase_per_user：同一用户可购买的累计次数，0 表示不限购。
type PaymentSaleRule struct {
	Audience           string `json:"audience" example:"all"`
	MaxPurchasePerUser int    `json:"max_purchase_per_user" example:"0"`
}

type AdminPaymentPackageListRequest struct {
	ProductID *int64 `json:"product_id"`
	Status    *int   `json:"status"`
	Keyword   string `json:"keyword"`
	PageNum   int    `json:"page_num"`
	PageSize  int    `json:"page_size"`
}

type PaymentPackageItem struct {
	ID          int64  `json:"id,omitempty"`
	ProductID   int64  `json:"product_id"`
	ProductCode string `json:"product_code" enums:"job_top,contact_voucher,paid_refresh,rent_publish" example:"paid_refresh"`
	ProductName string `json:"product_name" example:"付费刷新"`
	// SKU 唯一编码；创建支付订单时应传此字段，不得传客户端价格。
	SKUCode string `json:"sku_code" example:"paid_refresh_standard"`
	// 微信虚拟支付后台已发布的道具 ID；仅管理后台使用，不向小程序购买页返回。
	VirtualProductID string `json:"virtual_product_id" example:"wx_paid_refresh_1"`
	// 规格模式：1=单规格，2=多规格。
	SelectionMode int    `json:"selection_mode" enums:"1,2" example:"1"`
	Name          string `json:"name" example:"标准刷新"`
	Subtitle      string `json:"subtitle" example:"立即刷新信息排名"`
	Badge         string `json:"badge" example:""`
	// 售价，单位为分；例如199表示1.99元。
	PriceCents int64 `json:"price_cents" example:"199"`
	// 划线价，单位为分；0表示不展示划线价。
	OriginalPriceCents int64                `json:"original_price_cents" example:"0"`
	BenefitConfig      PaymentBenefitConfig `json:"benefit_config"`
	// 仅管理后台返回及配置。小程序端已根据此规则筛选可购买 SKU，不返回该字段。
	SaleRule  PaymentSaleRule `json:"sale_rule"`
	Status    int             `json:"status,omitempty"`
	Sort      int             `json:"sort"`
	Version   int             `json:"version,omitempty"`
	CreatedBy string          `json:"created_by,omitempty"`
	UpdatedBy string          `json:"updated_by,omitempty"`
	CreateAt  string          `json:"create_at,omitempty"`
	UpdateAt  string          `json:"update_at,omitempty"`
}

type AdminPaymentPackageListResponseData struct {
	List  []PaymentPackageItem `json:"list"`
	Total int64                `json:"total"`
}

type AdminPaymentPackageIDRequest struct {
	ID      int64  `json:"id" binding:"required"`
	Version int    `json:"version"`
	Reason  string `json:"reason"`
}

type AdminPaymentPackageDetailRequest struct {
	ID int64 `json:"id" binding:"required"`
}

type AdminPaymentPackageCreateRequest struct {
	// 收费业务 ID；创建后不可修改。
	ProductID int64 `json:"product_id" binding:"required"`
	// SKU 唯一编码；创建后不可修改。
	SKUCode          string `json:"sku_code" binding:"required"`
	VirtualProductID string `json:"virtual_product_id"`
	Name             string `json:"name" binding:"required"`
	Subtitle         string `json:"subtitle"`
	Badge            string `json:"badge"`
	// 售价，单位为分，必须大于 0。
	PriceCents int64 `json:"price_cents" binding:"required"`
	// 划线价，单位为分；0 表示不展示。
	OriginalPriceCents int64                `json:"original_price_cents"`
	Sort               int                  `json:"sort"`
	BenefitConfig      PaymentBenefitConfig `json:"benefit_config" binding:"required"`
	SaleRule           PaymentSaleRule      `json:"sale_rule"`
}

type AdminPaymentPackageUpdateRequest struct {
	ID               int64  `json:"id" binding:"required"`
	Version          int    `json:"version" binding:"required"`
	Name             string `json:"name" binding:"required"`
	Subtitle         string `json:"subtitle"`
	Badge            string `json:"badge"`
	VirtualProductID string `json:"virtual_product_id"`
	// 售价，单位为分，必须大于 0。
	PriceCents int64 `json:"price_cents" binding:"required"`
	// 划线价，单位为分；0 表示不展示。
	OriginalPriceCents int64                `json:"original_price_cents"`
	Sort               int                  `json:"sort"`
	BenefitConfig      PaymentBenefitConfig `json:"benefit_config" binding:"required"`
	SaleRule           PaymentSaleRule      `json:"sale_rule"`
	ChangeReason       string               `json:"change_reason"`
}

type AdminPaymentPackageCreateResponseData struct {
	ID int64 `json:"id"`
}

type AdminPaymentPackageHistoryRequest struct {
	SKUID    int64 `json:"sku_id" binding:"required"`
	PageNum  int   `json:"page_num"`
	PageSize int   `json:"page_size"`
}

type AdminPaymentPackageHistoryItem struct {
	ID             int64  `json:"id"`
	SKUID          int64  `json:"sku_id"`
	SKUVersion     int    `json:"sku_version"`
	Action         int    `json:"action"`
	BeforeSnapshot string `json:"before_snapshot"`
	AfterSnapshot  string `json:"after_snapshot"`
	ChangeReason   string `json:"change_reason"`
	Operator       string `json:"operator"`
	CreateAt       string `json:"create_at"`
}

type AdminPaymentPackageHistoryResponseData struct {
	List  []AdminPaymentPackageHistoryItem `json:"list"`
	Total int64                            `json:"total"`
}

type PaymentPackageListRequest struct {
	// 产品编码，前端应优先使用该字段：job_top=岗位置顶，contact_voucher=联系券，paid_refresh=付费刷新，rent_publish=招租发布。
	ProductCode string `json:"product_code" enums:"job_top,contact_voucher,paid_refresh,rent_publish" example:"paid_refresh"`
	// 当前业务类型：0=不限，1=招聘，2=求职，3=招租。岗位置顶传1或2；联系券传0；付费刷新传当前信息类型；招租发布传3。
	BizType int `json:"biz_type" enums:"0,1,2,3" example:"3"`
}

// PaymentSKU 是小程序购买页所需的最小 SKU 展示数据。购买资格由服务端筛选，不向客户端暴露营销规则。
type PaymentSKU struct {
	SKUCode            string               `json:"sku_code" example:"paid_refresh_1"`
	Name               string               `json:"name" example:"刷新1次"`
	Subtitle           string               `json:"subtitle" example:"立即刷新，提升曝光"`
	Badge              string               `json:"badge" example:"推荐"`
	PriceCents         int64                `json:"price_cents" example:"200"`
	OriginalPriceCents int64                `json:"original_price_cents" example:"400"`
	BenefitConfig      PaymentBenefitConfig `json:"benefit_config"`
}

type PaymentPackageListResponseData struct {
	ProductCode    string `json:"product_code" enums:"job_top,contact_voucher,paid_refresh,rent_publish" example:"paid_refresh"`
	ProductName    string `json:"product_name" example:"付费刷新"`
	PurchaseNotice string `json:"purchase_notice" example:"虚拟商品购买后自动到账"`
	// 规格模式：1=单规格，前端直接使用 skus[0]；2=多规格，前端展示 SKU 供用户选择。
	SelectionMode int `json:"selection_mode" enums:"1,2" example:"1"`
	// 当前已上架、适用于 biz_type 且符合当前用户购买资格的 SKU；空数组表示暂无可购买 SKU。
	SKUs []PaymentSKU `json:"skus"`
}

type PaymentProductItem struct {
	ID             int64  `json:"id"`
	ProductCode    string `json:"product_code"`
	Name           string `json:"name"`
	SelectionMode  int    `json:"selection_mode"`
	PurchaseNotice string `json:"purchase_notice"`
	PackageCount   int64  `json:"package_count"`
	CreateAt       string `json:"create_at"`
	UpdateAt       string `json:"update_at"`
}

type AdminPaymentProductUpdateRequest struct {
	ID             int64  `json:"id" binding:"required"`
	PurchaseNotice string `json:"purchase_notice"`
}

type PaymentProductListResponseData struct {
	List []PaymentProductItem `json:"list"`
}
