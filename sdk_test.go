package reportsdk

import "testing"

func TestDescriptorRequiresProtocolV2Capabilities(t *testing.T) {
	valid := Descriptor{ProtocolVersion: ProtocolVersionV2, Mode: DeploymentModeModule, Capabilities: []string{"definitions.sync", "snapshots.manage"}}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, descriptor := range map[string]Descriptor{
		"legacy":     {ProtocolVersion: ProtocolVersionV1, Mode: DeploymentModeModule, Capabilities: valid.Capabilities},
		"mode":       {ProtocolVersion: ProtocolVersionV2, Mode: "", Capabilities: valid.Capabilities},
		"capability": {ProtocolVersion: ProtocolVersionV2, Mode: DeploymentModeSaaS, Capabilities: []string{"definitions.sync"}},
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
