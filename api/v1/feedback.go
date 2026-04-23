package v1

import "github.com/go-nunu/nunu-layout-advanced/internal/model"

// FeedbackType 枚举
const (
	FeedbackTypeProductSuggestion = 1 // 产品建议
	FeedbackTypeFunctionIssue     = 2 // 功能问题
	FeedbackTypeContentCorrection = 3 // 内容修正
	FeedbackTypeOther             = 4 // 其他
)

var feedbackTypeNames = map[int]string{
	FeedbackTypeProductSuggestion: "产品建议",
	FeedbackTypeFunctionIssue:     "功能问题",
	FeedbackTypeContentCorrection: "内容修正",
	FeedbackTypeOther:             "其他",
}

func FeedbackTypeName(t model.FeedbackType) string {
	if name, ok := feedbackTypeNames[int(t)]; ok {
		return name
	}
	return "其他"
}

type FeedbackSubmitRequest struct {
	// 反馈类型: 1=产品建议 2=功能问题 3=内容修正 4=其他
	Type      int      `json:"type" binding:"required,min=1,max=4" enums:"1,2,3,4" example:"1"`
	Content   string   `json:"content" binding:"required,max=500"`
	PhotoURLs []string `json:"photo_urls"`
}

type FeedbackListRequest struct {
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
}

type FeedbackListItem struct {
	ID       int64    `json:"id"`
	// 反馈类型: 1=产品建议 2=功能问题 3=内容修正 4=其他
	Type     int      `json:"type" enums:"1,2,3,4" example:"1"`
	TypeName string   `json:"type_name" example:"产品建议"`
	Content  string   `json:"content"`
	PhotoURLs []string `json:"photo_urls"`
	CreateAt string   `json:"create_at"`
}

type FeedbackListResponseData struct {
	List  []FeedbackListItem `json:"list"`
	Total int64              `json:"total"`
}
