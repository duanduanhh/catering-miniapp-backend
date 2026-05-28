package v1

type ContactFeedbackConfig struct {
	Title   string         `json:"title"`
	Hint    string         `json:"hint"`
	Reasons []ReasonOption `json:"reasons"`
}

var ContactFeedbackConfigs = map[int]ContactFeedbackConfig{
	1: {
		Title: "岗位信息反馈",
		Hint:  "感谢您帮助我们维护平台信息的真实性；若平台核实为无效信息，将退还您1张联系券。",
		Reasons: []ReasonOption{
			{Value: 1, Label: "联系电话无法接通（关机/忙音/拒接）"},
			{Value: 2, Label: "电话是空号或错误号码"},
			{Value: 3, Label: "对方表示岗位已招满 / 不再招聘"},
			{Value: 4, Label: "岗位描述与实际情况严重不符（如薪资、地址）"},
			{Value: 5, Label: "其他问题（请说明）"},
		},
	},
	2: {
		Title: "求职信息反馈",
		Hint:  "感谢您帮助我们维护平台信息的真实性；若平台核实为无效信息，将退还您1张联系券。",
		Reasons: []ReasonOption{
			{Value: 1, Label: "联系电话无法接通（关机/忙音/拒接）"},
			{Value: 2, Label: "电话是空号或错误号码"},
			{Value: 3, Label: "对方表示已找到工作/不再求职"},
			{Value: 4, Label: "求职意向与沟通内容严重不符（如薪资期望等）"},
			{Value: 5, Label: "其他问题（请说明）"},
		},
	},
	3: {
		Title: "招租信息反馈",
		Hint:  "感谢您帮助我们维护平台信息的真实性；若平台核实为无效信息，将退还您1张联系券。",
		Reasons: []ReasonOption{
			{Value: 1, Label: "联系电话无法接通（关机/忙音/拒接）"},
			{Value: 2, Label: "电话是空号或错误号码"},
			{Value: 3, Label: "对方表示已租出去/不再出租"},
			{Value: 4, Label: "房屋信息与实际情况严重不符（如面积、租金、位置等）"},
			{Value: 5, Label: "其他问题（请说明）"},
		},
	},
}

type ContactFeedbackSubmitRequest struct {
	JobID       int64  `json:"job_id" binding:"required"`
	BizType     int    `json:"biz_type" binding:"required,min=1,max=3"`
	Reason      int    `json:"reason" binding:"required,min=1,max=5"`
	Description string `json:"description" binding:"max=500"`
}
