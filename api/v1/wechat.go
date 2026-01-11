package v1

type WechatLoginRequest struct {
	Code      string `json:"code" binding:"required"`
	LoginCode string `json:"loginCode" binding:"required"`
	InviterID int64  `json:"inviter_id"`
}

type WechatLoginResponseData struct {
	UserInfo  WechatLoginUserInfo `json:"user_info"`
	ExpiresIn int64               `json:"expires_in"`
}

type WechatLoginUserInfo struct {
	ID int64 `json:"id"`
}
