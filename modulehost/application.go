package modulehost

import (
	"context"
	"time"

	notificationmodel "github.com/domainry/domainry-notification-sdk/contract"
	reportmodel "github.com/domainry/domainry-report-sdk/model"
	reportpersistence "github.com/domainry/domainry-report-sdk/persistence"
)

// SubjectResolver is the only boundary allowed to translate caller proof into
// Report authorization facts. Embedded hosts may use an authenticated request
// context; remote services validate the original access token.
type SubjectResolver interface {
	ResolveReportSubject(context.Context, reportmodel.ReportAuthority) (reportmodel.ReportSubject, error)
}

// DatasetReader returns authorized source tuples from a safe pushdown or
// authorized records grouped by alias. Report owns fallback joins, filters,
// cardinality validation, aggregation, analyses, comparison and result shape.
type DatasetReader interface {
	ReadReportDataset(context.Context, reportmodel.ReportDatasetReadRequest) (reportmodel.ReportDatasetReadResult, error)
}

// ObjectSQLExecutor borrows host metadata authorization and SQL execution but
// accepts only a Report-compiled plan, never authored SQL text chosen by a
// request.
type ObjectSQLExecutor interface {
	ResolveReportObjectSQLSources(context.Context, reportmodel.ReportSchema, reportmodel.ReportSubject) (map[string]reportmodel.ReportSourceObject, error)
	AuthorizeReportObjectSQLPlan(context.Context, reportmodel.ReportSchema, reportmodel.ReportObjectSQLPlan, reportmodel.ReportSubject) error
	ExecuteReportObjectSQL(context.Context, reportmodel.ReportObjectSQLExecutionRequest) (reportmodel.ReportObjectSQLExecutionResult, error)
}

type SourceVersionReader interface {
	ReadReportSourceVersion(context.Context, reportmodel.ReportSchema, reportmodel.ReportSubject) (reportmodel.ReportSnapshotSourceVersion, error)
}

type ExecutionAudit interface {
	AppendReportExecution(context.Context, reportmodel.ReportSchema, reportmodel.ReportSummary, reportmodel.ReportSubject) error
}

// ExportAuthorization supplies only current host-owned Record authorization
// facts. Report owns scope normalization, projection, masking semantics and
// Object SQL export policy.
type ExportAuthorization interface {
	AuthorizeReportExportSource(context.Context, string, reportmodel.ReportSubject) error
	AuthorizeReportExportField(context.Context, string, string, reportmodel.ReportSubject) (bool, error)
}

// SnapshotTerminalCommitter atomically commits the Report-owned terminal
// snapshot state with a Notification-owned event. The host supplies the shared
// transaction adapter; neither owner writes the other's tables directly.
type SnapshotTerminalCommitter interface {
	CompleteReportSnapshot(context.Context, reportpersistence.SnapshotCompleteRequest, notificationmodel.NotificationIntent) error
	FailReportSnapshot(context.Context, reportpersistence.SnapshotFailRequest, notificationmodel.NotificationIntent) error
}

// ExportGateway is the narrow host integration used to submit Report exports.
// Data Exchange owns every job lifecycle operation after submission.
type ExportGateway interface {
	PrepareReportExport(context.Context, reportmodel.ReportExportPrepareRequest, reportmodel.ReportSchema, reportmodel.ReportExportControlSchema, reportmodel.ReportSubject) (reportmodel.ReportExportJob, error)
}

// ApplicationHost is the complete embedded host boundary required by Report
// business use cases. Persistence-only tests may still implement Host, while a
// production module factory must reject a host that lacks these capabilities.
type ApplicationHost interface {
	Host
	ReportSubjects() SubjectResolver
	ReportDatasets() DatasetReader
	ReportObjectSQL() ObjectSQLExecutor
	ReportSourceVersions() SourceVersionReader
	ReportExecutionAudit() ExecutionAudit
	ReportExportAuthorization() ExportAuthorization
	ReportSnapshotTerminals() SnapshotTerminalCommitter
	ReportExports() ExportGateway
	ReportCursorSigningKey() []byte
	ReportClock() func() time.Time
}
