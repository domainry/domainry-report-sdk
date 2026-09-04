package reportmodel

import "time"

// ReportSourceField and ReportSourceObject are the minimum metadata Report
// needs to compile calculations and Object SQL. Host authorization must omit
// disabled or unreadable fields before returning these values.
type ReportSourceField struct {
	Key       string `json:"key"`
	Type      string `json:"type"`
	Precision int    `json:"precision,omitempty"`
	Scale     int32  `json:"scale,omitempty"`
}

type ReportSourceObject struct {
	Key    string              `json:"key"`
	Fields []ReportSourceField `json:"fields"`
}

type ReportSourceRecord struct {
	ID        string         `json:"id"`
	Data      map[string]any `json:"data"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

type ReportObjectSQLExecutionRequest struct {
	Report       ReportSchema        `json:"report"`
	Plan         ReportObjectSQLPlan `json:"plan"`
	Parameters   map[string]any      `json:"parameters,omitempty"`
	Subject      ReportSubject       `json:"subject"`
	Timeout      time.Duration       `json:"-"`
	PageCursor   string              `json:"page_cursor,omitempty"`
	PagePosition int                 `json:"page_position,omitempty"`
	PageSize     int                 `json:"page_size,omitempty"`
}

type ReportObjectSQLExecutionResult struct {
	Rows       []map[string]string `json:"rows"`
	HasMore    bool                `json:"has_more"`
	Total      int                 `json:"total"`
	TotalKnown bool                `json:"total_known"`
	NextCursor string              `json:"next_cursor,omitempty"`
}
