package model

import "time"

type JobStatus int

const (
	JobStatusActive        JobStatus = 1
	JobStatusUserClosed    JobStatus = 2
	JobStatusAdminDisabled JobStatus = 3
	JobStatusDeleted       JobStatus = 4
)

type Job struct {
	ID                int64      `gorm:"primaryKey;column:id"`
	UserID            int64      `gorm:"column:user_id"`
	Positions         string     `gorm:"column:positions"`
	CompanyName       string     `gorm:"column:company_name"`
	Longitude         float64    `gorm:"column:longitude"`
	Latitude          float64    `gorm:"column:latitude"`
	Address           string     `gorm:"column:address"`
	ContactPersonName string     `gorm:"column:contact_person_name"`
	Contact           string     `gorm:"column:contact"`
	Description       string     `gorm:"column:description"`
	PhotoURLs         string     `gorm:"column:photo_urls"`
	Status            JobStatus  `gorm:"column:status"`
	FirstAreaID       int        `gorm:"column:first_area_id"`
	FirstAreaDes      string     `gorm:"column:first_area_des"`
	SecondAreaID      int        `gorm:"column:second_area_id"`
	SecondAreaDes     string     `gorm:"column:second_area_des"`
	ThirdAreaID       int        `gorm:"column:third_area_id"`
	ThirdAreaDes      string     `gorm:"column:third_area_des"`
	FourAreaID        int        `gorm:"column:four_area_id"`
	FourAreaDes       string     `gorm:"column:four_area_des"`
	SalaryMin         int        `gorm:"column:salary_min"`
	SalaryMax         int        `gorm:"column:salary_max"`
	BasicProtection   string     `gorm:"column:basic_protection"`
	SalaryBenefits    string     `gorm:"column:salary_benefits"`
	AttendanceLeave   string     `gorm:"column:attendance_leave"`
	CreateAt          time.Time  `gorm:"column:create_at"`
	UpdateAt          time.Time  `gorm:"column:update_at"`
	RefreshTime       *time.Time `gorm:"column:refresh_time"`
	TopStartTime      *time.Time `gorm:"column:top_start_time"`
	TopEndTime        *time.Time `gorm:"column:top_end_time"`
}

func (m *Job) TableName() string {
	return "job"
}
