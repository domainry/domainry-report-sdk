package reportsdk

// Report Action keys are stable executable identities. HTTP paths are
// bindings and may change without changing the Action identity.
const (
	ActionReportSummaryGet       = "report.summary.get"
	ActionReportQueryExecute     = "report.query.execute"
	ActionReportResultsRead      = "report.results.read"
	ActionReportSnapshotsRefresh = "report.snapshots.refresh"
	ActionReportExportsPrepare   = "report.exports.prepare"
)

const CapabilityReportBusiness = "report.business"
