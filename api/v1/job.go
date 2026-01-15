package v1

import "github.com/go-nunu/nunu-layout-advanced/internal/model"

type JobCreateRequest struct {
	Positions         string   `json:"positions" binding:"required"`
	CompanyName       string   `json:"company_name" binding:"required"`
	Longitude         float64  `json:"longitude" binding:"required"`
	Latitude          float64  `json:"latitude" binding:"required"`
	Address           string   `json:"address" binding:"required"`
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
	Positions       string   `json:"positions"`
	City            string   `json:"city"`
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
	Positions         string          `json:"positions"`
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
	CreateAt          string          `json:"create_at"`
	UpdateAt          string          `json:"update_at"`
	IsTop             int             `json:"is_top"`
	TopStartTime      string          `json:"top_start_time"`
	TopEndTime        string          `json:"top_end_time"`
	LastRefreshTime   string          `json:"last_refresh_time,omitempty"`
}

type JobListResponseData struct {
	Jobs  []JobListItem `json:"jobs"`
	Total int64         `json:"total"`
}

type JobListResponse struct {
	Response
	Data JobListResponseData
}

type JobMyRequest struct {
	BizType  int `json:"biz_type"`
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
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
	LastRefreshTime string `json:"last_refresh_time"`
}

type JobMyResponseData struct {
	List  []JobMyItem `json:"list"`
	Total int64       `json:"total"`
}
