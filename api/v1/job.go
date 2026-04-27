package v1

import "github.com/go-nunu/nunu-layout-advanced/internal/model"

type JobCreateRequest struct {
	BizType           int      `json:"biz_type"`              // 1=招聘（默认）2=求职，不传默认 1
	Positions         string   `json:"positions" binding:"required"`
	CompanyName       string   `json:"company_name"`  // 招聘必填，求职留空
	Longitude         float64  `json:"longitude"`     // 招聘必填，求职留空
	Latitude          float64  `json:"latitude"`      // 招聘必填，求职留空
	Address           string   `json:"address"`       // 招聘必填，求职留空
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
}

type JobCreateResponseData struct {
	JobID int64 `json:"job_id"`
}

type JobCreateResponse struct {
	Response
	Data JobCreateResponseData
}

type JobUpdateRequest struct {
	ID              int64    `json:"id" binding:"required"`
	Positions       *string  `json:"positions"`
	Longitude       *float64 `json:"longitude"`
	Latitude        *float64 `json:"latitude"`
	Address         *string  `json:"address"`
	Contact         *string  `json:"contact"`
	Description     *string  `json:"description"`
	PhotoURLs       []string `json:"photo_urls"`
	FirstAreaID     *int     `json:"first_area_id"`
	SecondAreaID    *int     `json:"second_area_id"`
	ThirdAreaID     *int     `json:"third_area_id"`
	FourAreaID      *int     `json:"four_area_id"`
	FirstAreaDes    *string  `json:"first_area_des"`
	SecondAreaDes   *string  `json:"second_area_des"`
	ThirdAreaDes    *string  `json:"third_area_des"`
	FourAreaDes     *string  `json:"four_area_des"`
	SalaryMin       *int     `json:"salary_min"`
	SalaryMax       *int     `json:"salary_max"`
	BasicProtection []string `json:"basic_protection"`
	SalaryBenefits  []string `json:"salary_benefits"`
	AttendanceLeave []string `json:"attendance_leave"`
}

type JobTopRequest struct {
	JobID   int64   `json:"job_id" binding:"required"`
	TopHour int     `json:"top_hour" binding:"required"`
	Price   float64 `json:"price" binding:"required"`
}

type JobRefreshRequest struct {
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
	JobID int64 `json:"job_id" binding:"required"`
}

type JobCancelCollectRequest struct {
	JobID int64 `json:"job_id" binding:"required"`
}

type JobInfoRequest struct {
	JobID int64 `json:"job_id" binding:"required"`
}

type JobFilter struct {
	BizType         int      `json:"biz_type"`      // 0=不限，1=招聘，2=求职
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
	Contact           string          `json:"contact"`
	ContactPersonName string          `json:"contact_person_name"`
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
	IsCollected       int             `json:"is_collected"` // 是否收藏：0=未收藏，1=已收藏
	CloseReason       string          `json:"close_reason"` // 关闭原因
	CloseTime         string          `json:"close_time"`   // 关闭时间
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
	JobID           int64  `json:"job_id"`
	Positions       string `json:"positions"`
	SalaryMin       int    `json:"salary_min"`
	SalaryMax       int    `json:"salary_max"`
	FirstAreaDes    string `json:"first_area_des"`
	SecondAreaDes   string `json:"second_area_des"`
	ThirdAreaDes    string `json:"third_area_des"`
	Address         string `json:"address"`
	CreateAt        string `json:"create_at"`
	IsTop           int    `json:"is_top"`
	Status          int    `json:"status"`
	LastRefreshTime string `json:"last_refresh_time"`
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
	Type  int `json:"type"`
	// 置顶区展示上限，不传默认5条
	Limit int `json:"limit"`
}

type HomeTopResponseData struct {
	List []JobListItem `json:"list"`
}

type HomeFeedRequest struct {
	// 内容类型：0或不传=全部，1=招聘，2=求职，3=招租
	Type     int `json:"type"`
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
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
