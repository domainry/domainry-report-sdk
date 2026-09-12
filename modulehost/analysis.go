package modulehost

import (
	"context"

	model "github.com/domainry/domainry-report-sdk/model"
)

// AnalysisSources optionally extends ObjectSQLExecutor with data-owner
// discovery. Only complete datasets and fields actually usable by this subject
// may be returned. Metadata contains no physical SQL, storage or credentials.
// ObjectSQLExecutor still authorizes every compiled plan before executing it;
// a catalog entry is never an execution grant or an executed result.
type AnalysisSources interface {
	ReportAnalysisSources(context.Context, model.ReportSubject) ([]model.AnalysisDataset, error)
}

// AnalysisSourceVersionReader supplies a stable content/revision fingerprint
// of every authorized input field used by the query. A row count plus latest
// timestamp is insufficient: same-time updates must invalidate old analysis.
// Empty datasets still have a non-empty SourceVersions entry. The fingerprint
// is computed at the data owner; raw source records never cross this port.
type AnalysisSourceVersionReader interface {
	ReadReportAnalysisSourceVersion(context.Context, model.ReportSchema, model.ReportSubject) (model.ReportSnapshotSourceVersion, error)
}
