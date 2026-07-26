package v1

type AdminJobListRequest struct {
	JobID    int64  `json:"job_id"`
	UserID   int64  `json:"user_id"`
	BizType  int    `json:"biz_type"`
	Status   []int  `json:"status"`
	Keyword  string `json:"keyword"`
	PageNum  int    `json:"page_num"`
	PageSize int    `json:"page_size"`
}

type AdminJobItem struct {
	JobID             int64          `json:"job_id"`
	BizType           int            `json:"biz_type"`
	Positions         string         `json:"positions"`
	CompanyName       string         `json:"company_name"`
	ContactPersonName string         `json:"contact_person_name"`
	Contact           string         `json:"contact"`
	Address           string         `json:"address"`
	AddressDetail     string         `json:"address_detail"`
	SalaryMin         int            `json:"salary_min"`
	SalaryMax         int            `json:"salary_max"`
	Status            int            `json:"status"`
	UserID            int64          `json:"user_id"`
	UserName          string         `json:"user_name"`
	UserPhone         string         `json:"user_phone"`
	CreateAt          string         `json:"create_at"`
	UpdateAt          string         `json:"update_at"`
	FirstAreaDes      string         `json:"first_area_des"`
	SecondAreaDes     string         `json:"second_area_des"`
	ThirdAreaDes      string         `json:"third_area_des"`
	Description       string         `json:"description"`
	WorkContent       string         `json:"work_content"`
	RecruitNum        int            `json:"recruit_num"`
	PhotoURLs         []string       `json:"photo_urls"`
	CloseReason       string         `json:"close_reason"`
	CloseTime         string         `json:"close_time"`
	RentDetail        *RentDetailDTO `json:"rent_detail,omitempty"`
}

type AdminJobListResponseData struct {
	List  []AdminJobItem `json:"list"`
	Total int64          `json:"total"`
}

type AdminJobIDRequest struct {
	JobID int64 `json:"job_id" binding:"required"`
}

type AdminJobUpdateRequest struct {
	JobID       int64   `json:"job_id" binding:"required"`
	Description *string `json:"description"`
	WorkContent *string `json:"work_content"`
}
