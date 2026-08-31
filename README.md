# Domainry Report SDK

Deployment-neutral contracts between Domainry Runtime and the Report owner.

## Package layout

- The root package is the stable Report `Factory`, `Binding`, rendering, and export entrypoint.
- `model` contains Report-owned snapshot values shared at the SDK boundary.
- `persistence` owns definition projection and snapshot claim, lease, fencing, completion, and failure contracts.
- `modulehost` describes host database, dialect, migration registrar, record access, and notification capabilities borrowed by an embedded module.

Report owns snapshot DDL and DML behind `persistence` contracts while embedded deployments participate in the host database, transaction boundary, and single `_schema_migrations` ledger.

Run `go test ./...` before publishing an immutable SDK version.
