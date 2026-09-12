package reportmodel

// AnalysisTableRow contains exactly the requested fields. A present nil cell
// is NULL; a missing field is an invalid source response. Numbers are decimal
// strings, dates are YYYY-MM-DD, and datetimes have explicit RFC3339 offsets.
// Formula evaluation and file parsing remain with the source service.
type AnalysisTableRow map[string]*string
