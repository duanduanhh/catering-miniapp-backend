package v1

type ContactHistoryListRequest struct {
	BizType  int `json:"biz_type"`
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
}

type ContactHistoryItem struct {
	ID               int64  `json:"id"`
	JobID            int64  `json:"job_id"`
	BizType          int    `json:"biz_type"`
	Positions        string `json:"positions"`
	Address          string `json:"address"`
	PurposeUserID    int64  `json:"purpose_user_id"`
	PurposeUserName  string `json:"purpose_user_name"`
	PurposeUserPhone string `json:"purpose_user_phone"`
	CreateAt         string `json:"create_at"`

	Avatar        string `json:"avatar"`          // 头像
	SalaryMin     int    `json:"salary_min"`      // 薪资下限
	SalaryMax     int    `json:"salary_max"`      // 薪资上限
	FirstAreaDes  string `json:"first_area_des"`  // 一级地区
	SecondAreaDes string `json:"second_area_des"` // 二级地区
	ThirdAreaDes  string `json:"third_area_des"`  // 三级地区
	JobStatus     int    `json:"job_status"`      // 岗位状态
	CompanyName   string `json:"company_name"`    // 企业名称
}

type ContactHistoryListResponseData struct {
	List  []ContactHistoryItem `json:"list"`
	Total int64                `json:"total"`
}

// ContactHistoryDeleteRequest 删除联系记录的请求
type ContactHistoryDeleteRequest struct {
	PurposeID int64 `json:"purpose_id" binding:"required"` // 职位ID
}
