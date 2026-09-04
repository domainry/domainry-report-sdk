# Domainry Report SDK

Deployment-neutral contracts between Domainry Runtime and the Report owner.

## Package layout

- The root package is the stable Report factory, application binding, query, snapshot, and export boundary.
- `contract` contains the dataset-plan validator and canonical hashing/filename helpers shared by authoring, owners, hosts, and workers.
- `model` contains Report-owned definitions, caller authority, source projections, requests, results, snapshots, and export values shared at the SDK boundary.
- `query` contains portable object/field metadata; `query/objectsql` owns the deployment-neutral Object SQL definition compiler and parameter contract. Physical SQL execution remains in the Report host adapter.
- `persistence` owns definition projection and snapshot claim, lease, fencing, completion, and failure contracts.
- `modulehost` describes the host database, dialect, migration registrar, authenticated subject resolution, authorized record reads, source-version fencing, Object SQL execution, audit, atomic snapshot/notification commit, and Data Exchange integration borrowed by an embedded module.

Protocol v3 requires `definitions.sync`, `queries.execute`, `snapshots.manage`, `exports.manage`, and `http.adapter`. Report owns four product HTTP operations—summary query, Object SQL query, snapshot refresh, and export preparation—and their OpenAPI operations through a Foundation `modulehttp.Adapter`; Runtime mounts and governs that Adapter without duplicating Report handlers or product contracts. `ExportGateway` only submits a prepared export; Data Exchange owns the job, cancellation, artifact, and download lifecycle after submission. Its asynchronous Report provider uses `Exports.ResolveExecution`, `ReadPage`, and `SourceVersion` for current definition resolution, scope authorization, execution, and fencing; a host schema projection or host-side Report engine is never an online source of truth.

Report owns snapshot DDL and DML behind `persistence` contracts while embedded deployments participate in the host database, transaction boundary, and single `_schema_migrations` ledger. Hosts must resolve trusted identity and authorization facts and provide only authorized source projections; Report owns dataset/Object SQL rules, pagination, materialization orchestration, and its stable response contract.

Run `go test ./...` before publishing an immutable SDK version.
