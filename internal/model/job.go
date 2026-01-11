package model

import "time"

type Job struct {
	ID              int64      `gorm:"primaryKey;column:id"`
	UserID          int64      `gorm:"column:user_id"`
	Positions       string     `gorm:"column:positions"`
	CompanyName     string     `gorm:"column:company_name"`
	Longitude       float64    `gorm:"column:longitude"`
	Latitude        float64    `gorm:"column:latitude"`
	Address         string     `gorm:"column:address"`
	Contact         string     `gorm:"column:contact"`
	Description     string     `gorm:"column:description"`
	PhotoURLs       string     `gorm:"column:photo_urls"`
	Status          int        `gorm:"column:status"`
	FailReason      string     `gorm:"column:fail_reason"`
	FirstAreaID     int        `gorm:"column:first_area_id"`
	FirstAreaDes    string     `gorm:"column:first_area_des"`
	SecondAreaID    int        `gorm:"column:second_area_id"`
	SecondAreaDes   string     `gorm:"column:second_area_des"`
	ThirdAreaID     int        `gorm:"column:third_area_id"`
	ThirdAreaDes    string     `gorm:"column:third_area_des"`
	FourAreaID      int        `gorm:"column:four_area_id"`
	FourAreaDes     string     `gorm:"column:four_area_des"`
	SalaryMin       int        `gorm:"column:salary_min"`
	SalaryMax       int        `gorm:"column:salary_max"`
	BasicProtection string     `gorm:"column:basic_protection"`
	SalaryBenefits  string     `gorm:"column:salary_benefits"`
	AttendanceLeave string     `gorm:"column:attendance_leave"`
	CreateAt        time.Time  `gorm:"column:create_at"`
	UpdateAt        time.Time  `gorm:"column:update_at"`
	RefreshTime     int64      `gorm:"column:refresh_time"`
	IsTop           int        `gorm:"column:is_top"`
	IsBuyTop        int        `gorm:"column:is_buy_top"`
	TopHour         int        `gorm:"column:top_hour"`
	TopStartTime    *time.Time `gorm:"column:top_start_time"`
	TopEndTime      *time.Time `gorm:"column:top_end_time"`
}

func (m *Job) TableName() string {
	return "job"
}
