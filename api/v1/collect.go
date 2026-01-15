package v1

type CollectMyRequest struct {
	BizType  int `json:"biz_type"`
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
}

type CollectMyResponseData struct {
	List  []JobMyItem `json:"list"`
	Total int64       `json:"total"`
}
