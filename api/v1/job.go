package v1

import "github.com/go-nunu/nunu-layout-advanced/internal/model"

type JobCreateRequest struct {
	BizType           int      `json:"biz_type"`
	Positions         string   `json:"positions" binding:"required"`
	CompanyName       string   `json:"company_name"`
	Longitude         float64  `json:"longitude"`
	Latitude          float64  `json:"latitude"`
	Address           string   `json:"address"`
	AddressDetail     string   `json:"address_detail"`
	Contact           string   `json:"contact" binding:"required"`
	ContactPersonName string   `json:"contact_person_name" binding:"required"`
	Description       string   `json:"description" binding:"required"`
	PhotoURLs         []string `json:"photo_urls"`
	FirstAreaID       int      `json:"first_area_id"`
	FirstAreaDes      string   `json:"first_area_des" binding:"required"`
	SecondAreaID      int      `json:"second_area_id"`
	SecondAreaDes     string   `json:"second_area_des" binding:"required"`
	ThirdAreaID       int      `json:"third_area_id"`
	ThirdAreaDes      string   `json:"third_area_des"`
	FourAreaID        int      `json:"four_area_id"`
	FourAreaDes       string   `json:"four_area_des"`
	SalaryMin         int      `json:"salary_min" binding:"required"`
	SalaryMax         int      `json:"salary_max" binding:"required"`
	BasicProtection   []string `json:"basic_protection"`
	SalaryBenefits    []string `json:"salary_benefits"`
	AttendanceLeave   []string `json:"attendance_leave"`
	EnterpriseID      int64    `json:"enterprise_id"`
	RecruitNum        int      `json:"recruit_num"`
	WorkContent       string   `json:"work_content"`
}

type JobCreateResponseData struct {
	JobID int64 `json:"job_id"`
}

type JobCreateResponse struct {
	Response
	Data JobCreateResponseData
}

type JobUpdateRequest struct {
	ID                int64          `json:"id" binding:"required"`
	Positions         *string        `json:"positions"`
	CompanyName       *string        `json:"company_name"`
	ContactPersonName *string        `json:"contact_person_name"`
	Longitude         *float64       `json:"longitude"`
	Latitude          *float64       `json:"latitude"`
	Address           *string        `json:"address"`
	AddressDetail     *string        `json:"address_detail"`
	Contact           *string        `json:"contact"`
	Description       *string        `json:"description"`
	PhotoURLs         []string       `json:"photo_urls"`
	FirstAreaID       *int           `json:"first_area_id"`
	SecondAreaID      *int           `json:"second_area_id"`
	ThirdAreaID       *int           `json:"third_area_id"`
	FourAreaID        *int           `json:"four_area_id"`
	FirstAreaDes      *string        `json:"first_area_des"`
	SecondAreaDes     *string        `json:"second_area_des"`
	ThirdAreaDes      *string        `json:"third_area_des"`
	FourAreaDes       *string        `json:"four_area_des"`
	SalaryMin         *int           `json:"salary_min"`
	SalaryMax         *int           `json:"salary_max"`
	BasicProtection   []string       `json:"basic_protection"`
	SalaryBenefits    []string       `json:"salary_benefits"`
	AttendanceLeave   []string       `json:"attendance_leave"`
	EnterpriseID      *int64         `json:"enterprise_id"`
	RecruitNum        *int           `json:"recruit_num"`
	WorkContent       *string        `json:"work_content"`
	RentDetail        *RentDetailDTO `json:"rent_detail"`
}

type JobTopRequest struct {
	JobID             int64   `json:"job_id" binding:"required"`
	TopHour           int     `json:"top_hour" binding:"required"`
	Price             float64 `json:"price" binding:"required"`
	ContactVoucherNum int     `json:"contact_voucher_num"` // 加赠联系券数量，0=不加赠
}

type JobRefreshRequest struct {
	JobID int64 `json:"job_id" binding:"required"`
}

type JobShareRefreshRequest struct {
	JobID int64 `json:"job_id" binding:"required"`
}

type JobCloseRequest struct {
	JobID       int64  `json:"job_id" binding:"required"`
	CloseReason string `json:"close_reason"`
}

type JobReopenRequest struct {
	JobID int64 `json:"job_id" binding:"required"`
}

type JobDeleteRequest struct {
	JobID int64 `json:"job_id" binding:"required"`
}

type JobCollectRequest struct {
	JobID   int64 `json:"job_id" binding:"required"`
	BizType int   `json:"biz_type" binding:"required"`
}

type JobCancelCollectRequest struct {
	JobID   int64 `json:"job_id" binding:"required"`
	BizType int   `json:"biz_type" binding:"required"`
}

type JobInfoRequest struct {
	JobID int64 `json:"job_id" binding:"required"`
}

type JobFilter struct {
	BizType         int      `json:"biz_type"` // 0=不限，1=招聘，2=求职，3=招租
	Positions       string   `json:"positions"`
	FirstAreaID     int      `json:"first_area_id"`
	SecondAreaID    int      `json:"second_area_id"`
	SalaryMin       int      `json:"salary_min"`
	SalaryMax       int      `json:"salary_max"`
	BasicProtection []string `json:"basic_protection"`
	SalaryBenefits  []string `json:"salary_benefits"`
	AttendanceLeave []string `json:"attendance_leave"`
	Longitude       float64  `json:"longitude"`
	Latitude        float64  `json:"latitude"`
	// 招租专属筛选（仅当 BizType=3 时生效）
	AreaSizeRange   int `json:"area_size_range"`   // 0=不限 1=<15 2=[15,30) 3=[30,50) 4=[50,100) 5=[100,200) 6=>=200
	TransferFeeFlag int `json:"transfer_fee_flag"` // 0=不限 1=有转让费 2=无转让费
}

type JobListRequest struct {
	RequestID string    `json:"request_id"`
	QueryType int       `json:"query_type"`
	Filter    JobFilter `json:"filter"`
	PageNum   int       `json:"page_num"`
	PageSize  int       `json:"page_size"`
}

type JobListItem struct {
	ID                int64           `json:"id"`
	UserID            int64           `json:"user_id"`
	BizType           int             `json:"biz_type"`
	Positions         string          `json:"positions"`
	CompanyName       string          `json:"company_name"`
	Longitude         float64         `json:"longitude"`
	Latitude          float64         `json:"latitude"`
	Address           string          `json:"address"`
	AddressDetail     string          `json:"address_detail"`
	ContactPersonName string          `json:"contact_person_name"`
	Contact           string          `json:"contact"`
	Description       string          `json:"description"`
	PhotoURLs         []string        `json:"photo_urls"`
	Status            model.JobStatus `json:"status"`
	FirstAreaID       int             `json:"first_area_id"`
	FirstAreaDes      string          `json:"first_area_des"`
	SecondAreaID      int             `json:"second_area_id"`
	SecondAreaDes     string          `json:"second_area_des"`
	ThirdAreaID       int             `json:"third_area_id"`
	ThirdAreaDes      string          `json:"third_area_des"`
	FourAreaID        int             `json:"four_area_id"`
	FourAreaDes       string          `json:"four_area_des"`
	SalaryMin         int             `json:"salary_min"`
	SalaryMax         int             `json:"salary_max"`
	BasicProtection   []string        `json:"basic_protection"`
	SalaryBenefits    []string        `json:"salary_benefits"`
	AttendanceLeave   []string        `json:"attendance_leave"`
	Avatar            string          `json:"avatar"`
	CreateAt          string          `json:"create_at"`
	UpdateAt          string          `json:"update_at"`
	IsTop             int             `json:"is_top"`
	TopStartTime      string          `json:"top_start_time"`
	TopEndTime        string          `json:"top_end_time"`
	LastRefreshTime   string          `json:"last_refresh_time,omitempty"`
	IsCollected       int             `json:"is_collected"`
	IsContacted       int             `json:"is_contacted"`
	CloseReason       string          `json:"close_reason"`
	CloseTime         string          `json:"close_time"`
	EnterpriseID      int64           `json:"enterprise_id"`
	EnterpriseName    string          `json:"enterprise_name"`
	RecruitNum        int             `json:"recruit_num"`
	WorkContent       string          `json:"work_content"`
	RentDetail        *RentDetailDTO  `json:"rent_detail,omitempty"` // 招租(biz_type=3)扩展字段
}

type JobListResponseData struct {
	Total int64         `json:"total"`
	Jobs  []JobListItem `json:"jobs"`
}

type JobListResponse struct {
	Response
	Data JobListResponseData
}

type JobMyRequest struct {
	BizType  int   `json:"biz_type"`
	PageNum  int   `json:"page_num"`
	PageSize int   `json:"page_size"`
	Status   []int `json:"status"` // 状态筛选：空数组或不传=查所有非删除，传数组按数组查
}

type JobMyItem struct {
	JobID             int64          `json:"job_id"`
	BizType           int            `json:"biz_type"`
	Positions         string         `json:"positions"`
	SalaryMin         int            `json:"salary_min"`
	SalaryMax         int            `json:"salary_max"`
	FirstAreaDes      string         `json:"first_area_des"`
	SecondAreaDes     string         `json:"second_area_des"`
	ThirdAreaDes      string         `json:"third_area_des"`
	Address           string         `json:"address"`
	AddressDetail     string         `json:"address_detail"`
	CompanyName       string         `json:"company_name"`
	CreateAt          string         `json:"create_at"`
	IsTop             int            `json:"is_top"`
	Status            int            `json:"status"`
	LastRefreshTime   string         `json:"last_refresh_time"`
	ContactPersonName string         `json:"contact_person_name"`
	Contact           string         `json:"contact"`
	Avatar            string         `json:"avatar"`
	BasicProtection   []string       `json:"basic_protection"`
	SalaryBenefits    []string       `json:"salary_benefits"`
	AttendanceLeave   []string       `json:"attendance_leave"`
	UserID            int64          `json:"user_id"`
	Description       string         `json:"description"`
	WorkContent       string         `json:"work_content"`
	RentDetail        *RentDetailDTO `json:"rent_detail,omitempty"` // 招租(biz_type=3)扩展字段
}

type JobMyResponseData struct {
	List  []JobMyItem `json:"list"`
	Total int64       `json:"total"`
}

// 内容类型常量
const (
	BizTypeAll     = 0 // 全部
	BizTypeRecruit = 1 // 招聘
	BizTypeResume  = 2 // 求职
	BizTypeRent    = 3 // 招租
)

type HomeTopRequest struct {
	// 内容类型：0或不传=全部，1=招聘，2=求职，3=招租
	Type int `json:"type"`
	// 置顶区展示上限，不传默认5条
	Limit        int `json:"limit"`
	FirstAreaID  int `json:"first_area_id"`
	SecondAreaID int `json:"second_area_id"`
}

type HomeTopResponseData struct {
	List []JobListItem `json:"list"`
}

type HomeFeedRequest struct {
	// 内容类型：0或不传=全部，1=招聘，2=求职，3=招租
	Type         int `json:"type"`
	PageNum      int `json:"page_num"`
	PageSize     int `json:"page_size"`
	FirstAreaID  int `json:"first_area_id"`
	SecondAreaID int `json:"second_area_id"`
}

type HomeFeedResponseData struct {
	List  []JobListItem `json:"list"`
	Total int64         `json:"total"`
}

type JobCloseReasonItem struct {
	Type    int      `json:"type"`    // 关闭类型：1=招聘，2=求职，3=招租
	Reasons []string `json:"reasons"` // 关闭原因列表
}

type JobCloseReasonResponseData struct {
	JobCloseReasonItem
}

type JobCloseReasonResponse struct {
	Response
	Data JobCloseReasonResponseData
}
