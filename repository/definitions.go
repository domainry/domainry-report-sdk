package repository

import (
	"context"
	"encoding/json"
)

type Definition struct {
	ResourceType string
	Key          string
	ObjectKey    string
	Name         string
	Payload      json.RawMessage
}

type DefinitionSnapshot struct {
	SchemaVersion string
	SourceKind    string
	SourceID      string
	Definitions   []Definition
}

type DefinitionRepository interface {
	SyncDefinitions(context.Context, DefinitionSnapshot) error
	DefinitionSnapshot(context.Context) (DefinitionSnapshot, error)
}

type Binding interface {
	DefinitionRepository() DefinitionRepository
}
