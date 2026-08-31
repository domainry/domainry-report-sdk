package reportmodel

type ReportSummaryRequest struct {
	ReportKey string            `json:"report_key"`
	Mode      string            `json:"mode,omitempty"`
	QueryKey  string            `json:"query_key,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Page      ReportPageRequest `json:"page,omitempty"`
}

type ReportObjectSQLRequest struct {
	ReportKey  string            `json:"report_key"`
	Parameters map[string]any    `json:"parameters"`
	Page       ReportPageRequest `json:"page,omitempty"`
}

type ReportSnapshotRefreshRequest struct {
	ReportKey      string `json:"report_key"`
	IdempotencyKey string `json:"-"`
}

type ReportExportPrepareRequest struct {
	ReportKey      string                   `json:"report_key"`
	ObjectKey      string                   `json:"object_key"`
	AuditID        string                   `json:"audit_id"`
	IdempotencyKey string                   `json:"-"`
	Scope          ReportExportScopeRequest `json:"scope"`
}
