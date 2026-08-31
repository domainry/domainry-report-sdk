package reportsdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/domainry/domainry-report-sdk/modulehost"
	"github.com/domainry/domainry-report-sdk/persistence"
)

const (
	ProtocolVersionV1 = "domainry-report-protocol-v1"
	ProtocolVersionV2 = "domainry-report-protocol-v2"
)

type DeploymentMode string

const (
	DeploymentModeModule DeploymentMode = "module"
	DeploymentModeSaaS   DeploymentMode = "saas"
)

type ApplicationRef struct {
	RuntimeID string `json:"runtime_id"`
}

func (r ApplicationRef) Validate() error {
	if strings.TrimSpace(r.RuntimeID) == "" {
		return fmt.Errorf("Report runtime identity is required")
	}
	return nil
}

type Descriptor struct {
	ProtocolVersion string
	Mode            DeploymentMode
	Capabilities    []string
}

func (d Descriptor) Validate() error {
	if d.ProtocolVersion != ProtocolVersionV2 {
		return fmt.Errorf("unsupported Report protocol %q", d.ProtocolVersion)
	}
	if d.Mode != DeploymentModeModule && d.Mode != DeploymentModeSaaS {
		return fmt.Errorf("invalid Report deployment mode %q", d.Mode)
	}
	required := map[string]bool{"definitions.sync": false, "snapshots.manage": false}
	for _, capability := range d.Capabilities {
		if _, ok := required[strings.TrimSpace(capability)]; ok {
			required[strings.TrimSpace(capability)] = true
		}
	}
	for capability, present := range required {
		if !present {
			return fmt.Errorf("Report capability %q is required", capability)
		}
	}
	return nil
}

type Factory interface {
	Open(context.Context, ApplicationRef, modulehost.Host) (Binding, error)
}

type Binding interface {
	Descriptor() Descriptor
	Definitions() persistence.DefinitionRepository
	Snapshots() persistence.SnapshotRepository
	Close(context.Context) error
}
