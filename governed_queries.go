package reportsdk

import (
	"context"

	model "github.com/domainry/domainry-report-sdk/model"
)

// GovernedQueries is an optional extension of Queries for consumers that must
// discover published contracts and reauthorize persisted results. Obtain it
// from ApplicationBinding.Queries(); older bindings remain source compatible.
// These methods require ActionReportQueryExecute and the definition's current
// audience, source and field permissions. Query executes the published report
// in realtime, with the owner's existing stable pagination. No SQL is accepted.
type GovernedQueries interface {
	Queries
	Catalog(context.Context, model.ReportCatalogRequest, model.ReportAuthority) (model.ReportCatalog, error)
	Query(context.Context, model.ReportObjectSQLRequest, model.ReportAuthority) (model.ReportQueryResult, error)
	// AuthorizeQueryResult checks the original request, all returned content,
	// current caller, definition, field permissions and source revision without
	// executing the report again. Changed sources invalidate the saved result;
	// consumers must query again instead of treating validation as execution.
	AuthorizeQueryResult(context.Context, model.ReportQueryResultAuthorization, model.ReportAuthority) error
}
