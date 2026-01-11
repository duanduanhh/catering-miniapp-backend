package v1

type UserInfoResponseData struct {
	UserID int64  `json:"user_id"`
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
	Sex    int    `json:"sex"`
	Phone  string `json:"phone"`
}

type UserInfoResponse struct {
	Response
	Data UserInfoResponseData
}

type UpdateUserInfoRequest struct {
	Avatar *string `json:"avatar"`
	Name   *string `json:"name"`
	Sex    *int    `json:"sex"`
	Phone  *string `json:"phone"`
}

type UpdateUserGeoRequest struct {
	FirstAreaID  *int     `json:"first_area_id"`
	SecondAreaID *int     `json:"second_area_id"`
	ThirdAreaID  *int     `json:"third_area_id"`
	Address      *string  `json:"address"`
	Longitude    *float64 `json:"longitude"`
	Latitude     *float64 `json:"latitude"`
}
