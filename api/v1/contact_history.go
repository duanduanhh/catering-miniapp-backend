package v1

type ContactHistoryListRequest struct {
	BizType  int `json:"biz_type"`
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
}

type ContactHistoryItem struct {
	ID              int64  `json:"id"`
	Positions       string `json:"positions"`
	Address         string `json:"address"`
	PurposeUserName string `json:"purpose_user_name"`
	CreateAt        string `json:"create_at"`
}

type ContactHistoryListResponseData struct {
	Contacts []ContactHistoryItem `json:"contacts"`
	Total    int64                `json:"total"`
}
