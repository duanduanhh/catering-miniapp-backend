package v1

// 招租店铺面积区间枚举（用于 /jobs/list 招租筛选）
const (
	AreaSizeRangeAll     = 0 // 不限
	AreaSizeRangeUnder15 = 1 // <15㎡
	AreaSizeRange15To30  = 2 // [15,30)㎡
	AreaSizeRange30To50  = 3 // [30,50)㎡
	AreaSizeRange50To100 = 4 // [50,100)㎡
	AreaSizeRange100T200 = 5 // [100,200)㎡
	AreaSizeRangeOver200 = 6 // >=200㎡
)

// 转让费筛选枚举
const (
	TransferFeeFlagAll = 0 // 不限
	TransferFeeFlagHas = 1 // 有转让费（固定或面议）
	TransferFeeFlagNo  = 2 // 无转让费
)

// RentDetailDTO 招租扩展字段（列表/详情 attach）
type RentDetailDTO struct {
	MonthlyRent int `json:"monthly_rent"` // 月租(元)
	AreaSize    int `json:"area_size"`    // 店铺面积(平方米)
	// 0=无转让费 1=固定金额 2=面议
	TransferFeeType   int    `json:"transfer_fee_type" enums:"0,1,2" example:"1"`
	TransferFeeAmount int    `json:"transfer_fee_amount"` // 转让费金额(元)
	TransferDesc      string `json:"transfer_desc"`
}

// RentPrePublishRequest 招租发布（付费）请求
type RentPrePublishRequest struct {
	SKUCode           string   `json:"sku_code" binding:"required" example:"rent_publish_1"`
	Positions         string   `json:"positions" binding:"required"` // 招租标题
	Address           string   `json:"address" binding:"required"`
	AddressDetail     string   `json:"address_detail"`
	Longitude         float64  `json:"longitude" binding:"required"`
	Latitude          float64  `json:"latitude" binding:"required"`
	Contact           string   `json:"contact" binding:"required"`
	ContactPersonName string   `json:"contact_person_name" binding:"required"`
	Description       string   `json:"description"`
	PhotoURLs         []string `json:"photo_urls" binding:"required,min=1,max=4"`
	FirstAreaID       int      `json:"first_area_id"`
	FirstAreaDes      string   `json:"first_area_des" binding:"required"`
	SecondAreaID      int      `json:"second_area_id"`
	SecondAreaDes     string   `json:"second_area_des" binding:"required"`
	ThirdAreaID       int      `json:"third_area_id"`
	ThirdAreaDes      string   `json:"third_area_des"`
	FourAreaID        int      `json:"four_area_id"`
	FourAreaDes       string   `json:"four_area_des"`
	// 招租扩展字段
	MonthlyRent       int    `json:"monthly_rent" binding:"required"`         // 月租(元)
	AreaSize          int    `json:"area_size" binding:"required"`            // 店铺面积(㎡)
	TransferFeeType   int    `json:"transfer_fee_type" binding:"oneof=0 1 2"` // 0=无 1=固定 2=面议
	TransferFeeAmount int    `json:"transfer_fee_amount"`                     // 仅当 TransferFeeType=1 时必填
	TransferDesc      string `json:"transfer_desc"`
}

// RentPrePublishResponseData 招租预发布响应
type RentPrePublishResponseData struct {
	JobID       int64  `json:"job_id"`
	OrderID     int64  `json:"order_id"`
	OrderNo     string `json:"order_no"`
	AmountCents int64  `json:"amount_cents" example:"1800"`
}

type RentPrePublishResponse struct {
	Response
	Data RentPrePublishResponseData
}
