package v1

type UserInfoResponseData struct {
	UserID                int64  `json:"user_id"`
	Avatar                string `json:"avatar"`
	Name                  string `json:"name"`
	Sex                   int    `json:"sex"`
	Phone                 string `json:"phone"`
	UserCode              string `json:"user_code"`
	ContactVoucherNum     int    `json:"contact_voucher_num"`
	FirstTopStatus        int    `json:"first_top_status"`
	NewCustomerStatus     int    `json:"new_customer_status"`
	ProfileCompleteStatus int    `json:"profile_complete_status"`
}

type UserInfoResponse struct {
	Response
	Data UserInfoResponseData
}

type UpdateUserInfoRequest struct {
	Avatar *string `json:"avatar"`
	Name   *string `json:"name"`
	Sex    *int    `json:"sex"`
	Phone  *string `json:"phone"`
}

type UpdateUserGeoRequest struct {
	FirstAreaID  *int     `json:"first_area_id"`
	SecondAreaID *int     `json:"second_area_id"`
	ThirdAreaID  *int     `json:"third_area_id"`
	Address      *string  `json:"address"`
	Longitude    *float64 `json:"longitude"`
	Latitude     *float64 `json:"latitude"`
}

type UserInviteListRequest struct {
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
}

type UserInviteItem struct {
	UserID        int64  `json:"user_id"`
	Avatar        string `json:"avatar"`
	Name          string `json:"name"`
	CreateAt      string `json:"create_at"`
	LoginStatus   int    `json:"login_status"`
	PublishStatus int    `json:"publish_status"`
	ConsumeStatus int    `json:"consume_status"`
	VoucherEarned int    `json:"voucher_earned"`
}

type UserInviteListResponseData struct {
	InviteTotal          int64            `json:"invite_total"`
	List                 []UserInviteItem `json:"list"`
	Total                int64            `json:"total"`
	LoginVoucherReward   int              `json:"login_voucher_reward"`
	PublishVoucherReward int              `json:"publish_voucher_reward"`
	ConsumeVoucherReward int              `json:"consume_voucher_reward"`
}
