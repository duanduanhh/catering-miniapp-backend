package v1

type UploadImageResponseData struct {
	URL         string `json:"url"`
	CheckMode   string `json:"check_mode,omitempty"`
	AuditStatus string `json:"audit_status,omitempty"`
	TraceID     string `json:"trace_id,omitempty"`
}
