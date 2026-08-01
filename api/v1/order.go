package v1

// WechatPayNotifyRequest 微信支付回调请求
// 微信支付回调是 JSON 请求体，需要直接接收原始 Body 进行验签和解密
type WechatPayNotifyRequest struct {
	// RawBody 用于接收原始请求体，在 middleware 中填充
	RawBody []byte `json:"-"`
}

type ContactVoucherBuyRequest struct {
	SKUCode string `json:"sku_code" binding:"required" example:"contact_voucher_5"`
}

type PayParams struct {
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

type PayOrderResponseData struct {
	OrderID   int64     `json:"order_id"`
	OrderNo   string    `json:"order_no"`
	Amount    float64   `json:"amount"`
	PayParams PayParams `json:"pay_params"`
}

type JobRefreshPayRequest struct {
	JobID   int64  `json:"job_id" binding:"required"`
	SKUCode string `json:"sku_code" binding:"required" example:"paid_refresh_1"`
}

type ContactVoucherCostRequest struct {
	PurposeID        *int64  `json:"purpose_id"`
	PurposeType      *int    `json:"purpose_type"`
	PurposeUserID    *int64  `json:"purpose_user_id"`
	PurposeUserName  *string `json:"purpose_user_name"`
	PurposeUserPhone *string `json:"purpose_user_phone"`
}

type ContactVoucherCallbackCostRequest struct {
	PurposeID        *int64  `json:"purpose_id"`
	PurposeType      *int    `json:"purpose_type"`
	PurposeUserID    *int64  `json:"purpose_user_id"`
	PurposeUserName  *string `json:"purpose_user_name"`
	PurposeUserPhone *string `json:"purpose_user_phone"`
}

type ContactVoucherRecordsResponseData struct {
	ContactVoucherNum int                         `json:"contact_voucher_num"`
	List              []ContactVoucherRecordsItem `json:"list"`
	ListTotal         int64                       `json:"list_total"`
}

type ContactVoucherRecordType string

const (
	ContactVoucherRecordBuy  ContactVoucherRecordType = "buy"
	ContactVoucherRecordCost ContactVoucherRecordType = "cost"
)

type ContactVoucherRecordsItem struct {
	ID        int64                    `json:"id"`
	Type      ContactVoucherRecordType `json:"type"`
	Title     string                   `json:"title"`
	ChangeNum int                      `json:"change_num"`
	CreateAt  string                   `json:"create_at"`
}

type OrderStatusRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
}

type OrderStatusResponseData struct {
	OrderNo string `json:"order_no"`
	// 订单状态: 1=待支付 2=已支付 3=已取消 4=已退款
	Status int `json:"status" enums:"1,2,3,4" example:"2"`
}

type UserOrderListRequest struct {
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
}

type UserOrderListItem struct {
	OrderID int64  `json:"order_id"`
	OrderNo string `json:"order_no"`
	// 商品类型: 1=置顶 2=联系券 3=刷新 4=招租发布
	ProductType int     `json:"product_type" enums:"1,2,3,4" example:"1"`
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	PaidAt      string  `json:"paid_at"`
	CreateAt    string  `json:"create_at"`
}

type UserOrderListResponseData struct {
	List  []UserOrderListItem `json:"list"`
	Total int64               `json:"total"`
}
