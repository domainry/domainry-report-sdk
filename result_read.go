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

// SharedResultReader authenticates an original producer's immutable proof,
// then checks the actual reader's current audience, fields and source scope.
// producer is proof provenance only; it never supplies execution authority to
// reader. The containing resource's explicit publication is checked by its owner.
type SharedResultReader interface {
	AuthorizeSharedCatalogRead(context.Context, model.ReportCatalogReadAuthorization, model.ReportAuthority, model.ReportAuthority) error
	AuthorizeSharedAnalysisCatalogRead(context.Context, model.AnalysisCatalogReadAuthorization, model.ReportAuthority, model.ReportAuthority) error
	AuthorizeSharedQueryResultRead(context.Context, model.ReportQueryResultAuthorization, model.ReportAuthority, model.ReportAuthority) error
	AuthorizeSharedAnalysisResultRead(context.Context, model.AnalysisResultAuthorization, model.ReportAuthority, model.ReportAuthority) error
}
