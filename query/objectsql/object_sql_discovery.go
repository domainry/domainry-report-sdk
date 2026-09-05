package objectsql

import (
	"fmt"
	"strings"

	reportmodel "github.com/domainry/domainry-report-sdk/model"
	"vitess.io/vitess/go/vt/sqlparser"
)

// DiscoverReportObjectSQLSources returns the stable Object keys named by the
// single supported FROM/JOIN tree. It does not bind fields or authorize the
// Objects; hosts use the result only to load the exact metadata that the full
// compiler subsequently validates.
func DiscoverReportObjectSQLSources(sql string) ([]string, error) {
	if len(sql) == 0 || len(sql) > reportObjectSQLMaximumLength {
		return nil, objectSQLPlanError("backend.report.object_sql_length_invalid", "object_sql_v1.sql", nil)
	}
	statement, err := sqlparser.NewTestParser().Parse(sql)
	if err != nil {
		return nil, objectSQLPlanError("backend.report.object_sql_parse_failed", "object_sql_v1.sql", map[string]string{"reason": err.Error()})
	}
	selected, ok := statement.(*sqlparser.Select)
	if !ok || len(selected.From) != 1 {
		return nil, objectSQLPlanError("backend.report.object_sql_source_invalid", "object_sql_v1.sql.from", map[string]string{"reason": "exactly one joined SELECT source tree is required"})
	}
	result := []string{}
	seen := map[string]bool{}
	var collect func(sqlparser.TableExpr) error
	collect = func(expression sqlparser.TableExpr) error {
		switch value := expression.(type) {
		case *sqlparser.AliasedTableExpr:
			name, valid := value.Expr.(sqlparser.TableName)
			if !valid || !name.Qualifier.IsEmpty() || strings.TrimSpace(value.As.String()) == "" || len(value.Partitions) > 0 || value.Hints != nil || len(value.Columns) > 0 {
				return objectSQLPlanError("backend.report.object_sql_source_invalid", "object_sql_v1.sql.from", nil)
			}
			key := strings.TrimSpace(name.Name.String())
			if key == "" {
				return objectSQLPlanError("backend.report.object_sql_source_invalid", "object_sql_v1.sql.from", nil)
			}
			if !seen[key] {
				seen[key] = true
				result = append(result, key)
			}
			return nil
		case *sqlparser.JoinTableExpr:
			joinType := strings.ToLower(strings.TrimSpace(value.Join.ToString()))
			if joinType != "join" && joinType != "inner join" && joinType != "left join" {
				return objectSQLPlanError("backend.report.object_sql_join_forbidden", "object_sql_v1.sql.from", map[string]string{"join": joinType})
			}
			if value.Condition == nil || value.Condition.On == nil || len(value.Condition.Using) > 0 {
				return objectSQLPlanError("backend.report.object_sql_join_invalid", "object_sql_v1.sql.from", nil)
			}
			if err := collect(value.LeftExpr); err != nil {
				return err
			}
			if _, valid := value.RightExpr.(*sqlparser.AliasedTableExpr); !valid {
				return objectSQLPlanError("backend.report.object_sql_join_invalid", "object_sql_v1.sql.from", map[string]string{"node": fmt.Sprintf("%T", value.RightExpr)})
			}
			return collect(value.RightExpr)
		default:
			return objectSQLPlanError("backend.report.object_sql_source_invalid", "object_sql_v1.sql.from", map[string]string{"node": fmt.Sprintf("%T", expression)})
		}
	}
	if err := collect(selected.From[0]); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, &reportmodel.ReportObjectSQLPlanError{Code: "backend.report.source_objects_required", Path: "object_sql_v1.sql.from"}
	}
	return result, nil
}
