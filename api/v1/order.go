package v1

type WechatPayRequest struct {
	OrderID int64   `json:"order_id"`
	OrderNo string  `json:"order_no"`
	Price   float64 `json:"price"`
}

type ContactVoucherBuyRequest struct {
	Price             float64 `json:"price" binding:"required"`
	ContactVoucherNum int     `json:"contact_voucher_num" binding:"required"`
}

type ContactVoucherBuyResponseData struct {
	OrderID           int64   `json:"order_id"`
	OrderNo           string  `json:"order_no"`
	BuyerUserID       int64   `json:"buyer_user_id"`
	ContactVoucherNum int     `json:"contact_voucher_num"`
	Price             float64 `json:"price"`
	CreatedAt         string  `json:"created_at"`
}

type JobTopResponseData struct {
	OrderID   int64   `json:"order_id"`
	OrderNo   string  `json:"order_no"`
	RuleID    int64   `json:"rule_id"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	CreatedAt string  `json:"created_at"`
}

type ContactVoucherCostRequest struct {
	PurposeID        *int64  `json:"purpose_id"`
	PurposeType      *int    `json:"purpose_type"`
	PurposeUserID    *int64  `json:"purpose_user_id"`
	PurposeUserPhone *string `json:"purpose_user_phone"`
}

type ContactVoucherMyResponseData struct {
	ContactVoucherNum int                     `json:"contact_voucher_num"`
	List              []ContactVoucherMyItem  `json:"list"`
	ListTotal         int64                   `json:"list_total"`
}

type ContactVoucherMyItem struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	ChangeNum int   `json:"change_num"`
	CreateAt string `json:"create_at"`
}
