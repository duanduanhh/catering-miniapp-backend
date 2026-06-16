package v1

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminLoginResponseData struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"` // Unix 秒
	Username  string `json:"username"`
}
