package v1

type WechatPayRequest struct {
	OrderID int64   `json:"order_id"`
	OrderNo string  `json:"order_no"`
	Price   float64 `json:"price"`
}

type WechatPayNotifyRequest struct {
	OrderNo    string  `json:"order_no" binding:"required"`
	Amount     float64 `json:"amount"`
	PayChannel string  `json:"pay_channel"`
	PayTradeNo string  `json:"pay_trade_no"`
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
