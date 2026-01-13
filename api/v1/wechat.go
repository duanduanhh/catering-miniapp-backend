package v1

type WechatLoginRequest struct {
	LoginCode string `json:"login_code" binding:"required"`
}

type WechatLoginResponseData struct {
	UserInfo WechatLoginUserInfo `json:"user_info"`
}

type WechatLoginUserInfo struct {
	ID int64 `json:"id"`
}

type WechatRegisterRequest struct {
	PhoneCode string `json:"phone_code" binding:"required"`
	LoginCode string `json:"loginCode" binding:"required"`
	InviterID int64  `json:"inviter_id"`
}
