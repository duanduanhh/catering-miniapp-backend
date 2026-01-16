package v1

type ContactHistoryListRequest struct {
	BizType  int `json:"biz_type"`
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
}

type ContactHistoryItem struct {
	ID               int64  `json:"id"`
	Positions        string `json:"positions"`
	Address          string `json:"address"`
	PurposeUserID    int64  `json:"purpose_user_id"`
	PurposeUserName  string `json:"purpose_user_name"`
	PurposeUserPhone string `json:"purpose_user_phone"`
	CreateAt         string `json:"create_at"`
}

type ContactHistoryListResponseData struct {
	List  []ContactHistoryItem `json:"list"`
	Total int64                `json:"total"`
}
