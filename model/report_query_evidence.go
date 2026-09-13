package reportmodel

// ReportCatalogRequest discovers only reports executable by the current caller.
// ReportKey selects one published definition; it is never an SQL expression.
type ReportCatalogRequest struct {
	ReportKey string            `json:"report_key,omitempty"`
	Page      ReportPageRequest `json:"page,omitempty"`
}

const (
	ReportCatalogDefaultSize = 20
	ReportCatalogMaximumSize = 50
)

// ReportCatalogEntry contains the authorized public contract, never authored
// SQL, physical sources, or the host's row/field authorization predicates.
type ReportCatalogEntry struct {
	Key               string                     `json:"key"`
	Name              string                     `json:"name,omitempty"`
	DefinitionVersion string                     `json:"definition_version"`
	Parameters        []ReportObjectSQLParameter `json:"parameters"`
	ResultSchema      []ReportResultColumnSchema `json:"result_schema"`
	RowLimit          int                        `json:"row_limit"`
}

type ReportCatalog struct {
	Reports    []ReportCatalogEntry `json:"reports"`
	NextCursor string               `json:"next_cursor,omitempty"`
	Truncated  bool                 `json:"truncated"`
	ReadProof  string               `json:"read_proof,omitempty"`
}

type ReportCatalogReadAuthorization struct {
	Request ReportCatalogRequest `json:"request"`
	Result  ReportCatalog        `json:"result"`
}

// ReportQuerySource binds a real query result to its definition, authorized
// source revision and caller. Complete means the entire published report was
// returned in this response, within RowLimit; it never means all source records.
// A final continuation page is not a complete report on its own.
type ReportQuerySource struct {
	ReportKey         string `json:"report_key"`
	DefinitionVersion string `json:"definition_version"`
	DataVersion       string `json:"data_version"`
	QueriedAt         string `json:"queried_at"`
	RowLimit          int    `json:"row_limit"`
	Complete          bool   `json:"complete"`
	Proof             string `json:"proof"`
	ReadProof         string `json:"read_proof,omitempty"`
}

type ReportQueryResult struct {
	Summary ReportSummary     `json:"summary"`
	Source  ReportQuerySource `json:"source"`
}

type ReportQueryResultAuthorization struct {
	Query  ReportObjectSQLRequest `json:"query"`
	Result ReportQueryResult      `json:"result"`
}
