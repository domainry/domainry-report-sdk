package reportsdk

import (
	"testing"

	"github.com/domainry/domainry-foundation/apperror"
)

func TestErrorClassificationContract(t *testing.T) {
	tests := []struct {
		status int
		kind   apperror.ErrorKind
		code   string
	}{
		{400, apperror.KindBadRequest, "report.query_invalid"},
		{403, apperror.KindForbidden, "report.permission_denied"},
		{409, apperror.KindConflict, "report.snapshot_fence_rejected"},
		{503, apperror.KindUnavailable, "report.database_unavailable"},
	}
	for _, test := range tests {
		err := &Error{StatusCode: test.status, Code: test.code, Params: map[string]string{"field": "amount"}}
		if apperror.KindOf(err) != test.kind || apperror.CodeOf(err) != test.code || apperror.ParamsOf(err)["field"] != "amount" {
			t.Fatalf("status=%d kind=%q code=%q params=%#v", test.status, apperror.KindOf(err), apperror.CodeOf(err), apperror.ParamsOf(err))
		}
	}
}
