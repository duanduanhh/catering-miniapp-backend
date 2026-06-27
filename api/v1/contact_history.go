package v1

type ContactHistoryListRequest struct {
	BizType  int `json:"biz_type"`
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
}

type ContactHistoryItem struct {
	ID       int64  `json:"id"`
	CreateAt string `json:"create_at"`

	// 发起方信息
	UserID     int64  `json:"user_id"`     // 发起方用户ID
	UserName   string `json:"user_name"`   // 发起方用户名
	UserPhone  string `json:"user_phone"`  // 发起方手机号
	UserAvatar string `json:"user_avatar"` // 发起方头像

	// 被联系方信息
	PurposeUserID    int64  `json:"purpose_user_id"`
	PurposeUserName  string `json:"purpose_user_name"`
	PurposeUserPhone string `json:"purpose_user_phone"`
	Avatar           string `json:"avatar"` // 被联系方头像

	// 岗位信息
	JobID         int64  `json:"job_id"`
	BizType       int    `json:"biz_type"`
	Positions     string `json:"positions"`
	Address       string `json:"address"`
	SalaryMin     int    `json:"salary_min"`
	SalaryMax     int    `json:"salary_max"`
	FirstAreaDes  string `json:"first_area_des"`
	SecondAreaDes string `json:"second_area_des"`
	ThirdAreaDes  string `json:"third_area_des"`
	JobStatus     int    `json:"job_status"`
	CompanyName   string `json:"company_name"`
}

type ContactHistoryListResponseData struct {
	List  []ContactHistoryItem `json:"list"`
	Total int64                `json:"total"`
}

// ContactHistoryDeleteRequest 删除联系记录的请求
type ContactHistoryDeleteRequest struct {
	PurposeID int64 `json:"purpose_id" binding:"required"` // 职位ID
}
