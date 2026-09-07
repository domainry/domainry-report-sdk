package objectsql

import (
	"strings"
	"testing"

	reportmodel "github.com/domainry/domainry-report-sdk/model"
	reportquery "github.com/domainry/domainry-report-sdk/query"
)

func TestObjectSQLContainsBindsTextFieldAndLiteralParameter(t *testing.T) {
	objects := map[string]reportquery.Object{"voucher": {Key: "voucher", Fields: []reportquery.Field{{Key: "customer", Type: "text"}, {Key: "cast_names", Type: "long_text"}}}}
	schema := reportmodel.ReportObjectSQLSchema{
		SQL:        "SELECT COUNT(v.id) AS records FROM voucher v WHERE CONTAINS(v.customer, :keyword) AND NOT CONTAINS(v.cast_names, :excluded_cast) LIMIT 1",
		Parameters: []reportmodel.ReportObjectSQLParameter{{Key: "keyword", Type: "text"}, {Key: "excluded_cast", Type: "text"}},
	}
	canonical, plan, err := CanonicalReportObjectSQL(schema, objects)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Sources) != 1 || plan.Where == nil || plan.Where.Type != "boolean" || plan.Where.Operator != "and" ||
		plan.Where.Arguments[0].Name != "contains" || plan.Where.Arguments[0].Arguments[0].FieldKey != "customer" ||
		plan.Where.Arguments[0].Arguments[1].Name != "keyword" || len(canonical.ResultSchema) != 1 || canonical.ResultSchema[0].Type != "integer" {
		t.Fatalf("canonical=%+v plan=%+v", canonical, plan)
	}
	if !containsObjectSQLString(plan.Sources[0].Fields, "customer") || !containsObjectSQLString(plan.Sources[0].Fields, "cast_names") {
		t.Fatalf("contains fields omitted from secured source: %+v", plan.Sources)
	}
	for _, keyword := range []string{" 50%_off~猫 ", "' OR 1=1 --", "", "[a-z]"} {
		parameters, err := NormalizeParameters(canonical.Parameters, map[string]any{"keyword": keyword})
		if err != nil || parameters["keyword"] != keyword || parameters["excluded_cast"] != nil {
			t.Fatalf("literal=%q parameters=%+v error=%v", keyword, parameters, err)
		}
	}
	// The envelope ID is a permitted filter field, never a new projected business field.
	schema.SQL = "SELECT COUNT(v.id) AS records FROM voucher v WHERE CONTAINS(v.id, :keyword) LIMIT 1"
	if _, _, err := CanonicalReportObjectSQL(schema, objects); err != nil {
		t.Fatal(err)
	}
}

func TestObjectSQLContainsRejectsUnsafeAndNonTextShapes(t *testing.T) {
	objects := map[string]reportquery.Object{"voucher": {Key: "voucher", Fields: []reportquery.Field{
		{Key: "customer", Type: "text"}, {Key: "amount", Type: "integer"}, {Key: "status", Type: "select"}, {Key: "customer_id", Type: "relation"},
	}}}
	for _, expression := range []string{
		"CONTAINS(v.customer)", "CONTAINS(v.customer, :keyword, :keyword)", "CONTAINS(v.amount, :keyword)",
		"CONTAINS(v.status, :keyword)", "CONTAINS(v.customer_id, :keyword)", "CONTAINS(v.created_at, :keyword)",
		"CONTAINS(v.customer, v.customer)", "CONTAINS(:keyword, :keyword)", "CONTAINS(v.customer, :count)",
		"CONTAINS(v.customer, 'literal')", "CONTAINS(v.customer, :missing)", "CONTAINS(v.owner_org_id, :keyword)",
		"CONTAINS(COALESCE(v.customer, :keyword), :keyword)", "evil.CONTAINS(v.customer, :keyword)",
		"v.customer LIKE :keyword", "v.customer LIKE :keyword ESCAPE '~'",
	} {
		t.Run(expression, func(t *testing.T) {
			schema := reportmodel.ReportObjectSQLSchema{
				SQL:        "SELECT COUNT(v.id) AS records FROM voucher v WHERE " + expression + " LIMIT 1",
				Parameters: []reportmodel.ReportObjectSQLParameter{{Key: "keyword", Type: "text"}, {Key: "count", Type: "integer"}},
			}
			if _, _, err := CanonicalReportObjectSQL(schema, objects); err == nil {
				t.Fatalf("accepted %s", expression)
			}
		})
	}
	if _, err := NormalizeParameters([]reportmodel.ReportObjectSQLParameter{{Key: "keyword", Type: "text"}}, map[string]any{"keyword": 10}); err == nil || !strings.Contains(err.Error(), "parameter_type_invalid") {
		t.Fatalf("numeric substring accepted: %v", err)
	}
}
