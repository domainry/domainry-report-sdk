package modulehost

import (
	"context"

	model "github.com/domainry/domainry-report-sdk/model"
)

// ResultReadScopeReader attests the exact currently authorized source
// projection after resolving every source and field of the supplied report.
// The opaque stable hash must bind workspace, sources, projected fields and
// every effective row/relationship/organization/business-profile predicate.
// It must change if that projection changes, including on an empty dataset.
// An unrelated execution permission/revision change alone must not change it.
// It never grants query execution and must not execute the query. Missing
// support leaves independent reading unavailable, rather than dropping RLS.
type ResultReadScopeReader interface {
	ReadReportResultScope(context.Context, model.ReportSchema, model.ReportSubject) (string, error)
}

// SharedResultScopeAuthorizer proves that the actual reader can read every
// source row/relationship, field and organization covered by the producer's
// current projection. A broader reader may qualify. No query is executed.
// Unsupported or unprovable predicates must fail closed.
type SharedResultScopeAuthorizer interface {
	AuthorizeSharedReportResultScope(context.Context, model.ReportSchema, model.ReportSubject, model.ReportSubject) error
}
