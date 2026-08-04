package v1

type ContactVoucherBuyRequest struct {
	SKUCode string `json:"sku_code" binding:"required" example:"contact_voucher_5"`
}

// PaymentOrderResponseData 业务下单成功后的最小响应。
// 小程序需随后调用 /payment/virtual/prepare 获取 wx.requestVirtualPayment 参数。
type PaymentOrderResponseData struct {
	OrderID     int64  `json:"order_id"`
	OrderNo     string `json:"order_no"`
	AmountCents int64  `json:"amount_cents" example:"1800"`
}

// VirtualPaymentPrepareRequest 由小程序在 wx.requestVirtualPayment 前调用。
// login_code 必须来自本次支付前新获取的 wx.login()，服务端只用于获取 session_key 生成用户态签名。
type VirtualPaymentPrepareRequest struct {
	OrderNo   string `json:"order_no" binding:"required" example:"TOP202608010001"`
	LoginCode string `json:"login_code" binding:"required"`
}

// VirtualPaymentParams 可直接传入 wx.requestVirtualPayment。
// sign_data 必须以字符串原样传入，不能在客户端重新序列化或修改。
type VirtualPaymentParams struct {
	SignData  string `json:"signData"`
	PaySig    string `json:"paySig"`
	Signature string `json:"signature"`
	Mode      string `json:"mode" example:"short_series_goods"`
}

type VirtualPaymentPrepareResponseData struct {
	OrderNo        string               `json:"order_no"`
	AmountCents    int64                `json:"amount_cents" example:"390"`
	VirtualPayment VirtualPaymentParams `json:"virtual_payment"`
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
