package model

import "time"

type User struct {
	ID                int64     `gorm:"primaryKey;column:id"`
	Avatar            string    `gorm:"column:avatar"`
	Name              string    `gorm:"column:name"`
	Sex               int       `gorm:"column:sex"`
	Age               int       `gorm:"column:age"`
	Birthday          string    `gorm:"column:birthday"`
	Phone             string    `gorm:"column:phone"`
	WechartCode       string    `gorm:"column:wechart_code"`
	WechatOpenID      string    `gorm:"column:wechat_open_id"`
	Token             string    `gorm:"column:token"`
	Password          string    `gorm:"column:password"`
	FirstAreaID       int       `gorm:"column:first_area_id"`
	SecondAreaID      int       `gorm:"column:second_area_id"`
	ThirdAreaID       int       `gorm:"column:third_area_id"`
	Address           string    `gorm:"column:address"`
	Longitude         float64   `gorm:"column:longitude"`
	Latitude          float64   `gorm:"column:latitude"`
	Type              int       `gorm:"column:type"`
	Status            int       `gorm:"column:status"`
	Integral          uint64    `gorm:"column:integral"`
	CollectNum        uint64    `gorm:"column:collect_num"`
	BuyNum            uint64    `gorm:"column:buy_num"`
	InviteID          int64     `gorm:"column:invite_id"`
	InviteNum         uint64    `gorm:"column:invite_num"`
	FirstRecharge     string    `gorm:"column:first_recharge"`
	TotalRecharge     float64   `gorm:"column:total_recharge"`
	DeviceModel       string    `gorm:"column:device_model"`
	IP                string    `gorm:"column:ip"`
	ContactVoucherNum int       `gorm:"column:contact_voucher_num"`
	CreateAt          time.Time `gorm:"column:create_at"`
	UpdateAt          time.Time `gorm:"column:update_at"`
}

func (u *User) TableName() string {
	return "user"
}
