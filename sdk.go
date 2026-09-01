package reportsdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/domainry/domainry-foundation/modulecapability"
	model "github.com/domainry/domainry-report-sdk/model"
	"github.com/domainry/domainry-report-sdk/modulehost"
	"github.com/domainry/domainry-report-sdk/persistence"
)

const (
	ProtocolVersionV1 = "domainry-report-protocol-v1"
	ProtocolVersionV2 = "domainry-report-protocol-v2"
	ProtocolVersionV3 = "domainry-report-protocol-v3"

	CapabilityDefinitionsSync = "definitions.sync"
	CapabilityQueriesExecute  = "queries.execute"
	CapabilitySnapshotsManage = "snapshots.manage"
	CapabilityExportsManage   = "exports.manage"
	CapabilityHTTPSurface     = "http.surface"
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
	if d.ProtocolVersion != ProtocolVersionV3 {
		return fmt.Errorf("unsupported Report protocol %q", d.ProtocolVersion)
	}
	if d.Mode != DeploymentModeModule && d.Mode != DeploymentModeSaaS {
		return fmt.Errorf("invalid Report deployment mode %q", d.Mode)
	}
	required := map[string]bool{
		CapabilityDefinitionsSync: false,
		CapabilityQueriesExecute:  false,
		CapabilitySnapshotsManage: false,
		CapabilityExportsManage:   false,
		CapabilityHTTPSurface:     false,
	}
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

// Error is stable across embedded and future remote Report bindings. Cause is
// local-only and must never be serialized.
type Error struct {
	StatusCode int               `json:"-"`
	Code       string            `json:"code"`
	Message    string            `json:"message,omitempty"`
	Retryable  bool              `json:"retryable,omitempty"`
	Params     map[string]string `json:"params,omitempty"`
	Cause      error             `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Message) != "" {
		return e.Code + ": " + e.Message
	}
	return e.Code
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type Queries interface {
	Summary(context.Context, model.ReportSummaryRequest, model.ReportAuthority) (model.ReportSummary, error)
	QueryObjectSQL(context.Context, model.ReportObjectSQLRequest, model.ReportAuthority) (model.ReportSummary, error)
}

type SnapshotCommands interface {
	Refresh(context.Context, model.ReportSnapshotRefreshRequest, model.ReportAuthority) (model.ReportSnapshot, error)
}

type Exports interface {
	Prepare(context.Context, model.ReportExportPrepareRequest, model.ReportAuthority) (model.ReportExportJob, error)
	ResolveExecution(context.Context, model.ReportExportExecutionRequest, model.ReportAuthority) (model.ReportExportExecution, error)
	ReadPage(context.Context, model.ReportExportExecutionRequest, model.ReportAuthority) (model.ReportSummary, error)
	SourceVersion(context.Context, model.ReportExportExecutionRequest, model.ReportAuthority) (model.ReportSnapshotSourceVersion, error)
}

// ApplicationBinding is the business-use-case boundary implemented by a full
// Report owner. Persistence repositories remain available only for host
// bootstrap and metadata restoration; product callers use these application
// services rather than reaching through to Report storage.
type ApplicationBinding interface {
	Binding
	Queries() Queries
	SnapshotCommands() SnapshotCommands
	Exports() Exports
}

// ApplicationHostBinder completes an embedded binding after the host has
// assembled its record, authorization and audit capabilities. Report opens
// persistence first so metadata restoration can run before Runtime business
// services exist; HTTP surfaces must not be published until this bind succeeds.
type ApplicationHostBinder interface {
	BindApplicationHost(modulehost.ApplicationHost) error
}

type Binding interface {
	modulecapability.Binding
	Descriptor() Descriptor
	Definitions() persistence.DefinitionRepository
	Snapshots() persistence.SnapshotRepository
	Close(context.Context) error
}
