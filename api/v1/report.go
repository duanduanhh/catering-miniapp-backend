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

// RentReportReasons 是招租（biz_type=3）专用的举报原因。
var RentReportReasons = []ReasonOption{
	{Value: 1, Label: "诈骗/收费"},
	{Value: 2, Label: "虚假信息/虚假宣传"},
	{Value: 3, Label: "无法联系到房东"},
	{Value: 4, Label: "其他"},
}

// ReportReasonsByBizType 根据业务类型返回可选原因。未指定业务类型时保留原有岗位举报原因，兼容旧客户端。
func ReportReasonsByBizType(bizType int) []ReasonOption {
	if bizType == BizTypeRent {
		return RentReportReasons
	}
	return ReportReasons
}

func IsValidReportReason(bizType, reason int) bool {
	for _, option := range ReportReasonsByBizType(bizType) {
		if option.Value == reason {
			return true
		}
	}
	return false
}

type ReportReasonsRequest struct {
	BizType int `form:"biz_type" binding:"omitempty,min=1,max=3"`
}

type ReportSubmitRequest struct {
	JobID       int64  `json:"job_id" binding:"required"`
	BizType     int    `json:"biz_type" binding:"required,min=1,max=3"`
	Reason      int    `json:"reason" binding:"required,min=1,max=5"`
	Description string `json:"description" binding:"max=500"`
}
