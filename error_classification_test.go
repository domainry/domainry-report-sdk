package reportsdk

import (
	"testing"

	"github.com/domainry/domainry-foundation/apperror"
)

func TestErrorClassificationContract(t *testing.T) {
	tests := []struct {
		class ErrorClass
		kind  apperror.ErrorKind
		code  string
	}{
		{ErrorClassBadRequest, apperror.KindBadRequest, "report.query_invalid"},
		{ErrorClassForbidden, apperror.KindForbidden, "report.permission_denied"},
		{ErrorClassConflict, apperror.KindConflict, "report.snapshot_fence_rejected"},
		{ErrorClassUnavailable, apperror.KindUnavailable, "report.database_unavailable"},
	}
	for _, test := range tests {
		err := &Error{Class: test.class, Code: test.code, Params: map[string]string{"field": "amount"}}
		if apperror.KindOf(err) != test.kind || apperror.CodeOf(err) != test.code || apperror.ParamsOf(err)["field"] != "amount" {
			t.Fatalf("class=%q kind=%q code=%q params=%#v", test.class, apperror.KindOf(err), apperror.CodeOf(err), apperror.ParamsOf(err))
		}
	}
}
