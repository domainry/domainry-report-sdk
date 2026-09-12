package reportsdk

import (
	"context"

	model "github.com/domainry/domainry-report-sdk/model"
)

// Analyses is an optional query binding capability. Each operation resolves
// current Report query permission and source/field/data authorization. Run
// executes full-dataset host aggregation, never client-side result pages.
// Obtain it from ApplicationBinding.Queries(); existing bindings stay valid.
type Analyses interface {
	AnalysisCatalog(context.Context, model.AnalysisCatalogRequest, model.ReportAuthority) (model.AnalysisCatalog, error)
	RunAnalysis(context.Context, model.AnalysisRequest, model.ReportAuthority) (model.AnalysisResult, error)
	AuthorizeAnalysisResult(context.Context, model.AnalysisResultAuthorization, model.ReportAuthority) error
}
