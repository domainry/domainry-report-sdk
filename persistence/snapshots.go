package persistence

import (
	"context"
	"encoding/json"
)

type Snapshot struct {
	ID              string            `json:"id"`
	WorkspaceID     string            `json:"workspace_id"`
	ReportKey       string            `json:"report_key"`
	AccessScopeHash string            `json:"-"`
	IdempotencyKey  string            `json:"idempotency_key"`
	Status          string            `json:"status"`
	Summary         json.RawMessage   `json:"summary,omitempty"`
	Watermark       string            `json:"watermark,omitempty"`
	SourceVersions  map[string]string `json:"source_versions,omitempty"`
	StartedAt       string            `json:"started_at"`
	RefreshedAt     string            `json:"refreshed_at,omitempty"`
	ErrorCode       string            `json:"error_code,omitempty"`
	LeaseOwner      string            `json:"-"`
	LeaseExpiresAt  string            `json:"-"`
	FencingToken    int64             `json:"-"`
}

type SnapshotBeginRequest struct {
	WorkspaceID, ReportKey, AccessScopeHash, IdempotencyKey string
	StartedAt, LeaseOwner, LeaseExpiresAt                   string
}

type SnapshotCompleteRequest struct {
	Snapshot       Snapshot
	ExpectedStatus string
	LeaseOwner     string
	FencingToken   int64
}

type SnapshotFailRequest struct {
	WorkspaceID, ID, ExpectedStatus, ErrorCode, LeaseOwner string
	FencingToken                                           int64
}

type SnapshotClaimDisposition string

const (
	SnapshotClaimAcquired SnapshotClaimDisposition = "acquired"
	SnapshotClaimRunning  SnapshotClaimDisposition = "running"
	SnapshotClaimReplay   SnapshotClaimDisposition = "replay"
)

type SnapshotClaim struct {
	Snapshot    Snapshot
	Disposition SnapshotClaimDisposition
}

func (c SnapshotClaim) Acquired() bool { return c.Disposition == SnapshotClaimAcquired }

type SnapshotRepository interface {
	Claim(context.Context, SnapshotBeginRequest) (SnapshotClaim, error)
	Complete(context.Context, SnapshotCompleteRequest) error
	Fail(context.Context, SnapshotFailRequest) error
	Latest(context.Context, string, string, string) (Snapshot, bool, error)
}
