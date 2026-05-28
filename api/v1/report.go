package v1

type ReasonOption struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}

type ReportReasonsResponseData struct {
	Reasons []ReasonOption `json:"reasons"`
}

var ReportReasons = []ReasonOption{
	{Value: 1, Label: "诈骗/收费"},
	{Value: 2, Label: "薪资虚假"},
	{Value: 3, Label: "代招冒充直招"},
	{Value: 4, Label: "无法联系到招聘方"},
	{Value: 5, Label: "其他"},
}

type ReportSubmitRequest struct {
	JobID       int64  `json:"job_id" binding:"required"`
	BizType     int    `json:"biz_type" binding:"required,min=1,max=3"`
	Reason      int    `json:"reason" binding:"required,min=1,max=5"`
	Description string `json:"description" binding:"max=500"`
}
