package reportmodel

import (
	"encoding/json"
	"testing"

	identitysdk "github.com/domainry/domainry-identity-sdk"
)

func TestTrustedProcessUsesOnlyExactNonSerializedCapabilities(t *testing.T) {
	subject := ReportSubject{
		Principal:           identitysdk.Principal{Known: true, WorkspaceID: "workspace-1", UserID: "runtime-target-executor"},
		AccessScopeHash:     "scope-1",
		TrustedProcess:      true,
		ProcessCapabilities: []string{"report.snapshots.refresh", "opportunity.read"},
	}
	if err := subject.Validate(); err != nil {
		t.Fatal(err)
	}
	if !subject.HasPermission("report.snapshots.refresh") || !subject.HasAllPermissions([]string{"report.snapshots.refresh", "opportunity.read"}) {
		t.Fatalf("exact process capabilities were not honored: %#v", subject.ProcessCapabilities)
	}
	if subject.HasPermission("opportunity.export") || subject.HasAllPermissions(nil) {
		t.Fatal("trusted process authority widened beyond its exact capabilities")
	}
	expanded := subject.WithExactProcessCapabilities("customer.read", "*.read", "customer.read")
	if !expanded.HasPermission("customer.read") || expanded.HasPermission("customer.export") || len(expanded.ProcessCapabilities) != 3 {
		t.Fatalf("exact expansion=%#v", expanded.ProcessCapabilities)
	}
	human := subject
	human.TrustedProcess = false
	human.ProcessCapabilities = nil
	if human.WithExactProcessCapabilities("customer.read").HasPermission("customer.read") {
		t.Fatal("external subject received trusted process capabilities")
	}
	encoded, err := json.Marshal(subject)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) == "" || json.Valid(encoded) == false {
		t.Fatalf("invalid JSON: %q", encoded)
	}
	var decoded ReportSubject
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.TrustedProcess || len(decoded.ProcessCapabilities) != 0 {
		t.Fatal("trusted process authority crossed the JSON boundary")
	}
}
