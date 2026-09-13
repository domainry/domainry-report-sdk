package reportsdk

import (
	"context"

	model "github.com/domainry/domainry-report-sdk/model"
)

// ResultReader validates saved Report/Analysis results independently of query
// execution. The current caller needs ActionReportResultsRead and all source,
// audience and field permissions. Only a source-attested reading scope and an
// owner-issued ReadProof can authorize the result; no query is reexecuted.
// Legacy results without a ReadProof require the original execution-authorized
// APIs. Access to a containing delivery is the consumer's responsibility.
type ResultReader interface {
	AuthorizeCatalogRead(context.Context, model.ReportCatalogReadAuthorization, model.ReportAuthority) error
	AuthorizeAnalysisCatalogRead(context.Context, model.AnalysisCatalogReadAuthorization, model.ReportAuthority) error
	AuthorizeQueryResultRead(context.Context, model.ReportQueryResultAuthorization, model.ReportAuthority) error
	AuthorizeAnalysisResultRead(context.Context, model.AnalysisResultAuthorization, model.ReportAuthority) error
}
