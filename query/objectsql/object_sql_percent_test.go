package objectsql

import (
	"testing"

	reportmodel "github.com/domainry/domainry-report-sdk/model"
	reportquery "github.com/domainry/domainry-report-sdk/query"
)

func TestObjectSQLPreservesPercentScaleThroughPeriodProduct(t *testing.T) {
	objects := map[string]reportquery.Object{"nomination": {Key: "nomination", Fields: []reportquery.Field{
		{Key: "fees", Type: "integer"}, {Key: "rate", Type: "percent", Precision: 7, Scale: 4},
	}}}
	for _, product := range []string{
		"SUM(s.fees) * MAX(s.rate)", "MAX(s.rate) * SUM(s.fees)",
		"SUM(s.fees) * (CASE WHEN SUM(s.fees) >= 50000 THEN MAX(s.rate) ELSE 0 END)",
	} {
		t.Run(product, func(t *testing.T) {
			plan, err := CompileReportObjectSQL(reportmodel.ReportObjectSQLSchema{SQL: "SELECT " + product + " AS scaled_amount, FLOOR(" + product + ") AS whole_amount FROM nomination s LIMIT 1"}, objects)
			if err != nil {
				t.Fatal(err)
			}
			productResult, floorResult := plan.ResultSchema[0], plan.ResultSchema[1]
			if productResult.Type != "decimal" || productResult.Precision != 26 || productResult.Scale != 4 || floorResult.Type != "integer" || floorResult.Precision != 0 || floorResult.Scale != 0 {
				t.Fatalf("product/floor lost exact scale: %+v", plan.ResultSchema)
			}
		})
	}
}
