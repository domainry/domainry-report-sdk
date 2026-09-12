package reportmodel

// AnalysisDataset is current, authorized metadata supplied by the data owner.
// A key identifies the entire governed dataset, never one search/result page.
// Unit is source-declared; an empty unit is explicitly unknown to consumers.
type AnalysisDataset struct {
	Key     string           `json:"key"`
	Name    string           `json:"name"`
	Kind    string           `json:"kind"`
	Version string           `json:"version"`
	Columns []AnalysisColumn `json:"columns"`
}

type AnalysisColumn struct {
	Key       string `json:"key"`
	Type      string `json:"type"`
	Unit      string `json:"unit"`
	Precision int    `json:"precision,omitempty"`
	Scale     int    `json:"scale,omitempty"`
}

type AnalysisCatalogRequest struct {
	DatasetKey string            `json:"dataset_key,omitempty"`
	Page       ReportPageRequest `json:"page,omitempty"`
}

type AnalysisCatalog struct {
	Datasets   []AnalysisDataset `json:"datasets"`
	NextCursor string            `json:"next_cursor,omitempty"`
	Truncated  bool              `json:"truncated"`
}

// A filter is exactly one leaf, all-group or any-group. Values are bounded,
// typed scalar parameters; no SQL, predicate text or caller scope is accepted.
type AnalysisFilter struct {
	Field    string           `json:"field,omitempty"`
	Operator string           `json:"operator,omitempty"`
	Values   []any            `json:"values,omitempty"`
	All      []AnalysisFilter `json:"all,omitempty"`
	Any      []AnalysisFilter `json:"any,omitempty"`
}

type AnalysisMeasure struct {
	Key      string `json:"key"`
	Function string `json:"function"`
	Field    string `json:"field,omitempty"`
	Distinct bool   `json:"distinct,omitempty"`
}

type AnalysisTimeBucket struct {
	Field    string `json:"field"`
	Grain    string `json:"grain"`
	TimeZone string `json:"time_zone"`
}

// Baseline and Current are independent filters over the same authorized
// dataset. They may overlap; their input counts must not be added as a union.
type AnalysisComparison struct {
	Baseline []AnalysisFilter `json:"baseline"`
	Current  []AnalysisFilter `json:"current"`
}

// Expressions use reference, decimal constant, or a fixed arithmetic operator.
// They are not evaluated as source code. References bind to known output keys.
type AnalysisExpression struct {
	Reference string               `json:"reference,omitempty"`
	Decimal   string               `json:"decimal,omitempty"`
	Operator  string               `json:"operator,omitempty"`
	Arguments []AnalysisExpression `json:"arguments,omitempty"`
}

type AnalysisCalculation struct {
	Key        string             `json:"key"`
	Expression AnalysisExpression `json:"expression"`
	Scale      int                `json:"scale"`
}

type AnalysisAnomalyRule struct {
	Key      string   `json:"key"`
	Column   string   `json:"column"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}

type AnalysisRequest struct {
	DatasetKey   string                `json:"dataset_key"`
	Mode         string                `json:"mode"`
	Filters      []AnalysisFilter      `json:"filters,omitempty"`
	GroupBy      []string              `json:"group_by,omitempty"`
	Measures     []AnalysisMeasure     `json:"measures,omitempty"`
	Time         *AnalysisTimeBucket   `json:"time,omitempty"`
	Comparison   *AnalysisComparison   `json:"comparison,omitempty"`
	Select       []string              `json:"select,omitempty"`
	Calculations []AnalysisCalculation `json:"calculations,omitempty"`
	AnomalyRules []AnalysisAnomalyRule `json:"anomaly_rules,omitempty"`
	MaxRows      int                   `json:"max_rows,omitempty"`
}

// All values preserve decimal strings; nil is SQL NULL, not an empty string.
// Undefined arithmetic remains nil and is explained by a cell issue.
type AnalysisRow struct {
	InputCounts   map[string]string   `json:"input_counts"`
	NonNullCounts map[string]string   `json:"non_null_counts"`
	Values        map[string]*string  `json:"values"`
	Anomalies     []string            `json:"anomalies"`
	Issues        []AnalysisCellIssue `json:"issues"`
}

type AnalysisCellIssue struct {
	Column string `json:"column"`
	Code   string `json:"code"`
}

type AnalysisMethod struct {
	Numerics string `json:"numerics,omitempty"`
	Column   string `json:"column"`
	Method   string `json:"method"`
	Nulls    string `json:"nulls"`
	Rounding string `json:"rounding,omitempty"`
}

type AnalysisSource struct {
	DatasetKey        string            `json:"dataset_key"`
	DefinitionVersion string            `json:"definition_version"`
	DataVersion       string            `json:"data_version"`
	QueriedAt         string            `json:"queried_at"`
	InputCounts       map[string]string `json:"input_counts"`
	Scope             string            `json:"scope"`
	Complete          bool              `json:"complete"`
	Proof             string            `json:"proof"`
}

type AnalysisResult struct {
	Spec    AnalysisRequest  `json:"spec"`
	Columns []AnalysisColumn `json:"columns"`
	Rows    []AnalysisRow    `json:"rows"`
	Methods []AnalysisMethod `json:"methods"`
	Source  AnalysisSource   `json:"source"`
}

type AnalysisResultAuthorization struct {
	Request AnalysisRequest `json:"request"`
	Result  AnalysisResult  `json:"result"`
}
