package reportmodel

// ReportExportScopeRequest is the closed, user-controlled part of a governed
// export. Dataset keys select only declared fields/dimensions/measures/analyses;
// Object SQL values bind only declared typed parameters. Neither form accepts
// SQL text or request-selected identifiers.
type ReportExportScopeRequest struct {
	Parameters        map[string]any              `json:"parameters,omitempty"`
	QueryKey          string                      `json:"query_key,omitempty"`
	AnalysisKey       string                      `json:"analysis_key,omitempty"`
	Filters           []ReportExportFilter        `json:"filters,omitempty"`
	DateRange         *ReportExportDateRange      `json:"date_range,omitempty"`
	TimeZone          string                      `json:"timezone,omitempty"`
	Tags              []string                    `json:"tags,omitempty"`
	TagMatch          string                      `json:"tag_match,omitempty"`
	FieldProjection   []string                    `json:"field_projection,omitempty"`
	Purpose           string                      `json:"purpose"`
	MetricDefinitions []ReportMetricDefinitionRef `json:"metric_definitions,omitempty"`
	Freshness         ReportExportFreshness       `json:"freshness"`
}

type ReportExportFilter struct {
	DimensionKey string   `json:"dimension_key"`
	Operator     string   `json:"operator"`
	Values       []string `json:"values,omitempty"`
}

type ReportExportDateRange struct {
	DimensionKey string `json:"dimension_key"`
	From         string `json:"from"`
	To           string `json:"to"`
}

type ReportMetricDefinitionRef struct {
	Key     string `json:"key"`
	Version string `json:"version"`
}

type ReportExportFreshness struct {
	Mode              string `json:"mode"`
	SnapshotID        string `json:"snapshot_id,omitempty"`
	MaximumLagSeconds int64  `json:"maximum_lag_seconds,omitempty"`
}

type ReportExportDefinition struct {
	Report  ReportSchema              `json:"report"`
	Control ReportExportControlSchema `json:"control"`
}

// ReportExportExecutionRequest is the closed worker-side input for resolving
// and reading a governed export. ReportKey/ObjectKey select owner definitions;
// Scope can only narrow those definitions and Page carries an opaque
// Report-owned cursor.
type ReportExportExecutionRequest struct {
	ReportKey string                   `json:"report_key"`
	ObjectKey string                   `json:"object_key"`
	Scope     ReportExportScopeRequest `json:"scope"`
	Page      ReportPageRequest        `json:"page,omitempty"`
}

// ReportExportExecution is the current, authorized owner projection used by a
// Data Exchange provider. Callers use the normalized scope and masking facts
// for artifact encoding; the executable scoped definition stays inside Report.
type ReportExportExecution struct {
	Definition       ReportExportDefinition   `json:"definition"`
	Scope            ReportExportScopeRequest `json:"scope"`
	MaskedDimensions map[string]bool          `json:"masked_dimensions,omitempty"`
}

// ReportExportJob is the Report-specific projection of a Data Exchange job.
// ID is the Data Exchange job identity used by all lifecycle routes.
type ReportExportJob struct {
	ID             string                   `json:"id"`
	AuditID        string                   `json:"audit_id"`
	ReportKey      string                   `json:"report_key"`
	ObjectKey      string                   `json:"object_key"`
	Status         string                   `json:"status"`
	PagesCompleted int                      `json:"pages_completed"`
	RowsExported   int                      `json:"rows_exported"`
	Total          int                      `json:"total"`
	Scope          ReportExportScopeRequest `json:"scope"`
	ArtifactID     string                   `json:"artifact_id,omitempty"`
	ContentSHA256  string                   `json:"content_sha256,omitempty"`
	ExpiresAt      string                   `json:"expires_at,omitempty"`
	ErrorCode      string                   `json:"error_code,omitempty"`
	CreatedAt      string                   `json:"created_at"`
	UpdatedAt      string                   `json:"updated_at"`
}
