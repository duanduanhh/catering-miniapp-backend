package v1

// 通用分页/时间区间字段，admin 列表接口共用。
type AdminTimeRangeRequest struct {
	StartTime string `json:"start_time"` // YYYY-MM-DD HH:mm:ss，含
	EndTime   string `json:"end_time"`
}

// ===== 用户列表 =====

type AdminUserListRequest struct {
	UserID    int64  `json:"user_id"`
	Keyword   string `json:"keyword"`
	Status    *int   `json:"status"`
	Type      *int   `json:"type"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	PageNum   int    `json:"page_num"`
	PageSize  int    `json:"page_size"`
}

type AdminUserItem struct {
	UserID            int64   `json:"user_id"`
	Avatar            string  `json:"avatar"`
	Name              string  `json:"name"`
	Phone             string  `json:"phone"`
	UserCode          string  `json:"user_code"`
	Sex               int     `json:"sex"`
	Age               int     `json:"age"`
	Type              int     `json:"type"`
	Status            int     `json:"status"`
	ContactVoucherNum int     `json:"contact_voucher_num"`
	CollectNum        uint64  `json:"collect_num"`
	BuyNum            uint64  `json:"buy_num"`
	InviteNum         uint64  `json:"invite_num"`
	InviterID         int64   `json:"inviter_id"`
	TotalRecharge     float64 `json:"total_recharge"`
	Address           string  `json:"address"`
	CreateAt          string  `json:"create_at"`
	UpdateAt          string  `json:"update_at"`
}

type AdminUserListResponseData struct {
	List  []AdminUserItem `json:"list"`
	Total int64           `json:"total"`
}

// ===== 企业列表 =====

type AdminEnterpriseListRequest struct {
	EnterpriseID int64  `json:"enterprise_id"`
	Keyword      string `json:"keyword"` // name / social_credit_code
	Status       *int   `json:"status"`
	UserID       int64  `json:"user_id"`
	PageNum      int    `json:"page_num"`
	PageSize     int    `json:"page_size"`
}

type AdminEnterpriseItem struct {
	ID                  int64  `json:"id"`
	UserID              int64  `json:"user_id"`
	UserName            string `json:"user_name"`
	UserPhone           string `json:"user_phone"`
	Name                string `json:"name"`
	SocialCreditCode    string `json:"social_credit_code"`
	LegalRepresentative string `json:"legal_representative"`
	Address             string `json:"address"`
	BusinessScope       string `json:"business_scope"`
	LicenseURL          string `json:"license_url"`
	IsDefault           int    `json:"is_default"`
	Status              int    `json:"status"`
	CreateAt            string `json:"create_at"`
	UpdateAt            string `json:"update_at"`
}

type AdminEnterpriseListResponseData struct {
	List  []AdminEnterpriseItem `json:"list"`
	Total int64                 `json:"total"`
}

// ===== 意见反馈列表 =====

type AdminFeedbackListRequest struct {
	FeedbackID int64  `json:"feedback_id"`
	Type       *int   `json:"type"`
	UserID     int64  `json:"user_id"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	PageNum    int    `json:"page_num"`
	PageSize   int    `json:"page_size"`
}

type AdminFeedbackItem struct {
	ID        int64    `json:"id"`
	UserID    int64    `json:"user_id"`
	UserName  string   `json:"user_name"`
	UserPhone string   `json:"user_phone"`
	Type      int      `json:"type"`
	TypeName  string   `json:"type_name"`
	Content   string   `json:"content"`
	PhotoURLs []string `json:"photo_urls"`
	CreateAt  string   `json:"create_at"`
}

type AdminFeedbackListResponseData struct {
	List  []AdminFeedbackItem `json:"list"`
	Total int64               `json:"total"`
}

// ===== 联系记录列表 =====

type AdminContactHistoryListRequest struct {
	ID            int64  `json:"id"`
	UserID        int64  `json:"user_id"`
	PurposeUserID int64  `json:"purpose_user_id"`
	JobID         int64  `json:"job_id"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	PageNum       int    `json:"page_num"`
	PageSize      int    `json:"page_size"`
}

type AdminContactHistoryItem struct {
	ID                 int64  `json:"id"`
	UserID             int64  `json:"user_id"`
	UserName           string `json:"user_name"`
	UserPhone          string `json:"user_phone"`
	PurposeID          int64  `json:"purpose_id"` // 即 job_id
	PurposeType        int    `json:"purpose_type"`
	PurposeUserID      int64  `json:"purpose_user_id"`
	PurposeUserName    string `json:"purpose_user_name"`
	PurposeUserPhone   string `json:"purpose_user_phone"`
	UserDeleted        int    `json:"user_deleted"`
	PurposeUserDeleted int    `json:"purpose_user_deleted"`
	CreateAt           string `json:"create_at"`
	// 关联反馈（取该联系记录最新一条）
	FeedbackID          int64  `json:"feedback_id"` // 0 表示无反馈
	FeedbackReason      int    `json:"feedback_reason"`
	FeedbackReasonLabel string `json:"feedback_reason_label"`
	FeedbackStatus      int    `json:"feedback_status"`    // 1 待处理 2 已核实 3 已驳回
	FeedbackCreateAt    string `json:"feedback_create_at"` // 空字符串表示无
}

type AdminContactHistoryListResponseData struct {
	List  []AdminContactHistoryItem `json:"list"`
	Total int64                     `json:"total"`
}

// ===== 举报列表 =====

type AdminReportListRequest struct {
	ReportID  int64  `json:"report_id"`
	Status    *int   `json:"status"`
	Reason    *int   `json:"reason"`
	BizType   *int   `json:"biz_type"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	PageNum   int    `json:"page_num"`
	PageSize  int    `json:"page_size"`
}

type AdminReportItem struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	UserName    string `json:"user_name"`
	UserPhone   string `json:"user_phone"`
	JobID       int64  `json:"job_id"`
	BizType     int    `json:"biz_type"`
	Reason      int    `json:"reason"`
	ReasonLabel string `json:"reason_label"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	CreateAt    string `json:"create_at"`
}

type AdminReportListResponseData struct {
	List  []AdminReportItem `json:"list"`
	Total int64             `json:"total"`
}
