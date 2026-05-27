package v1

type AdminJobListRequest struct {
	BizType  int    `json:"biz_type"`
	Status   []int  `json:"status"`
	Keyword  string `json:"keyword"`
	PageNum  int    `json:"page_num"`
	PageSize int    `json:"page_size"`
}

type AdminJobItem struct {
	JobID         int64    `json:"job_id"`
	BizType       int      `json:"biz_type"`
	Positions     string   `json:"positions"`
	CompanyName   string   `json:"company_name"`
	Address       string   `json:"address"`
	SalaryMin     int      `json:"salary_min"`
	SalaryMax     int      `json:"salary_max"`
	Status        int      `json:"status"`
	UserID        int64    `json:"user_id"`
	UserName      string   `json:"user_name"`
	UserPhone     string   `json:"user_phone"`
	CreateAt      string   `json:"create_at"`
	UpdateAt      string   `json:"update_at"`
	FirstAreaDes  string   `json:"first_area_des"`
	SecondAreaDes string   `json:"second_area_des"`
	ThirdAreaDes  string   `json:"third_area_des"`
	Description   string   `json:"description"`
	PhotoURLs     []string `json:"photo_urls"`
}

type AdminJobListResponseData struct {
	List  []AdminJobItem `json:"list"`
	Total int64          `json:"total"`
}

type AdminJobIDRequest struct {
	JobID int64 `json:"job_id" binding:"required"`
}
