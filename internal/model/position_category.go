package model

type PositionCategory struct {
	ID            int64                 `gorm:"primaryKey" json:"id"`
	CategoryName  string                `gorm:"size:64;not null" json:"category_name"`
	SortOrder     int                   `gorm:"not null;default:0" json:"-"`
	Status        int8                  `gorm:"not null;default:1" json:"-"`
	Subcategories []PositionSubcategory `gorm:"foreignKey:CategoryID" json:"subcategories"`
}

type PositionSubcategory struct {
	ID              int64  `gorm:"primaryKey" json:"id"`
	CategoryID      int64  `gorm:"not null" json:"-"`
	SubcategoryName string `gorm:"size:64;not null" json:"subcategory_name"`
	Description     string `gorm:"type:text" json:"description"`
	SortOrder       int    `gorm:"not null;default:0" json:"-"`
	Status          int8   `gorm:"not null;default:1" json:"-"`
}

func (PositionCategory) TableName() string {
	return "position_category"
}

func (PositionSubcategory) TableName() string {
	return "position_subcategory"
}
