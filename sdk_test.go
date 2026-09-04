package reportsdk

import "testing"

func TestDescriptorRequiresProtocolV3Capabilities(t *testing.T) {
	valid := Descriptor{ProtocolVersion: ProtocolVersionV3, Mode: DeploymentModeModule, Capabilities: []string{
		CapabilityDefinitionsSync, CapabilityQueriesExecute, CapabilitySnapshotsManage, CapabilityExportsManage, CapabilityHTTPAdapter,
	}}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, descriptor := range map[string]Descriptor{
		"legacy":     {ProtocolVersion: ProtocolVersionV2, Mode: DeploymentModeModule, Capabilities: valid.Capabilities},
		"mode":       {ProtocolVersion: ProtocolVersionV3, Mode: "", Capabilities: valid.Capabilities},
		"capability": {ProtocolVersion: ProtocolVersionV3, Mode: DeploymentModeSaaS, Capabilities: []string{CapabilityDefinitionsSync}},
	} {
		t.Run(name, func(t *testing.T) {
			if descriptor.Validate() == nil {
				t.Fatal("invalid descriptor accepted")
			}
		})
	}
}

func TestApplicationRefRequiresRuntimeIdentity(t *testing.T) {
	if (ApplicationRef{}).Validate() == nil {
		t.Fatal("empty runtime identity accepted")
	}
	if err := (ApplicationRef{RuntimeID: "runtime-a"}).Validate(); err != nil {
		t.Fatal(err)
	}
}
