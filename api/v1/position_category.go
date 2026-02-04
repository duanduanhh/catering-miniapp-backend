package v1

type PositionCategory struct {
	ID            int64                  `json:"id"`
	CategoryName  string                 `json:"category_name"`
	Subcategories []*PositionSubcategory `json:"subcategories"`
	SubTotal      int                    `json:"sub_total"`
}

type PositionSubcategory struct {
	ID              int64  `json:"id"`
	SubcategoryName string `json:"subcategory_name"`
}

type GetPositionCategoryResponseData struct {
	List  []PositionCategory `json:"list"`
	Total int                `json:"total"`
}
