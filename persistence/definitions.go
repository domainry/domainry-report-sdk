package persistence

import (
	"context"

	reportmodel "github.com/domainry/domainry-report-sdk/model"
)

type DefinitionSnapshot struct {
	SchemaVersion string
	SourceKind    string
	SourceID      string
	Definitions   []reportmodel.ReportDefinitionSchema
}

type DefinitionRepository interface {
	SyncDefinitions(context.Context, DefinitionSnapshot) error
	DefinitionSnapshot(context.Context) (DefinitionSnapshot, error)
}

type Binding interface {
	DefinitionRepository() DefinitionRepository
}
