package modulehost

import (
	"context"

	metadatasdk "github.com/domainry/domainry-metadata-sdk"
	ormmigration "github.com/domainry/domainry-orm/migration"
	"github.com/domainry/domainry-orm/sqlhost"
)

type Database = sqlhost.Database
type DBTX = sqlhost.DBTX

type Dialect interface {
	Identifier(string) string
	Table(string) string
	Placeholder(int) string
	Insert(string, []string) string
}

type SchemaMigration = ormmigration.Migration

type MigrationRegistrar interface {
	Driver() string
	Schema() string
	ApplyOwnedMigrations(context.Context, string, []SchemaMigration) error
}

type Host interface {
	Database() Database
	// DatabaseFor returns the host transaction carried by ctx when one exists,
	// otherwise the host database. Embedded Report persistence must use this
	// instead of inventing a module-owned transaction boundary.
	DatabaseFor(context.Context) DBTX
	Dialect() Dialect
	Migrations() MigrationRegistrar
}

// DefinitionStoreHost supplies the shared installation-scoped Definition
// store. Report borrows this host port in Module mode and supplies the same
// contract from its service database in SaaS mode.
type DefinitionStoreHost interface {
	DefinitionStore() metadatasdk.DefinitionStore
}
