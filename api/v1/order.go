package v1

// WechatPayNotifyRequest 微信支付回调请求
// 微信支付回调是 JSON 请求体，需要直接接收原始 Body 进行验签和解密
type WechatPayNotifyRequest struct {
	// RawBody 用于接收原始请求体，在 middleware 中填充
	RawBody []byte `json:"-"`
}

type ContactVoucherBuyRequest struct {
	Price             float64 `json:"price" binding:"required"`
	ContactVoucherNum int     `json:"contact_voucher_num" binding:"required"`
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
	JobID int64   `json:"job_id" binding:"required"`
	Price float64 `json:"price" binding:"required"`
}

type ContactVoucherCostRequest struct {
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
