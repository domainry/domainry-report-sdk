package modulehost

import (
	"context"

	model "github.com/domainry/domainry-report-sdk/model"
)

// AnalysisTableVersion describes a complete, currently authorized structured
// file table. DataVersion binds the immutable original, parser/schema version,
// sheet/range, ordered projected cells and row count. Search passages and
// preview pages cannot implement this contract. DefinitionVersion must equal
// the catalog dataset's Version; Rows includes all rows before analysis filters.
type AnalysisTableVersion struct {
	DatasetKey        string
	DefinitionVersion string
	DataVersion       string
	// ContentSHA256 is the lowercase SHA-256 of a canonical stream: encode
	// the sorted requested field array, then each row's aligned nullable string
	// array, using Go encoding/json.Encoder defaults (HTML escaping, UTF-8,
	// compact arrays and one LF per array). Even zero-column rows encode [].
	// This checksum lets Report verify every delivered cell against the owner's
	// current complete snapshot, not merely trust a count or an EOF flag.
	ContentSHA256 string
	Rows          int64
	Complete      bool
}

// AnalysisTables is an optional data-owner port, independent of Object SQL.
// Both methods must freshly authorize the dataset and every requested field.
// Stream must deliver every authorized row once, synchronously and in stable
// source order; it must stop on consumer/context error and return the matching
// complete version only after EOF and a final authorization/version check.
// Sources that cannot attest completeness must return an error, never snippets.
// Neither storage handles nor caller-selected permissions cross this port.
type AnalysisTables interface {
	ReadReportAnalysisTableVersion(context.Context, string, []string, model.ReportSubject) (AnalysisTableVersion, error)
	StreamReportAnalysisTable(context.Context, AnalysisTableVersion, []string, model.ReportSubject, func(model.AnalysisTableRow) error) (AnalysisTableVersion, error)
}

// AnalysisTableSource owns discovery and access for file tables only. It has
// no SQL dependency; a product adapter may consume a Knowledge/data-service
// public contract without either owner importing the other's implementation.
type AnalysisTableSource interface {
	AnalysisSources
	AnalysisTables
}

// AnalysisTableHost optionally supplies a file-table source at composition.
// Existing ApplicationHost and ObjectSQLExecutor implementations are unchanged.
type AnalysisTableHost interface {
	ReportAnalysisTables() AnalysisTableSource
}
