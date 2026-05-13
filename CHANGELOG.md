# Changelog

All notable changes to this project will be documented in this file. The format
is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

- Initial port of [scylladb/gocqlx](https://github.com/scylladb/gocqlx) onto
  [apache/cassandra-gocql-driver/v2](https://github.com/apache/cassandra-gocql-driver),
  published as a standalone module `github.com/moguchev/cqlx`. No `replace`
  directive required in downstream `go.mod`.
- Ported subpackages: `cqlx` (root), `qb`, `table`, `dbutil`, `cqlxtest`,
  `cmd/schemagen`.
- `examples/` directory with the full integration example suite
  (`//go:build all || integration`).
- `Queryx` wrappers for Apache v2's newer `*Query` methods: `SetKeyspace`,
  `WithNowInSeconds`.
- Documentation: [`docs/relationship-to-gocqlx.md`](docs/relationship-to-gocqlx.md),
  [`docs/api-deltas.md`](docs/api-deltas.md).
- GitHub Actions matrix CI against `cassandra:4.1` and `cassandra:5.0`.

### Changed

- `cmd/schemagen` reworked for Apache v2's `gocql.TypeInfo` metadata model.
  `mapScyllaToGoType(string)` renamed to `mapCqlTypeToGoType(TypeInfo)`; the
  views template block iterates the underlying base table; user-type and
  materialized-view bindings renamed to match Apache v2's `UserTypes` and
  `MaterializedViews` keyspace fields.
- `Queryx.*Release` methods are now no-op wrappers around their non-Release
  counterparts. Apache v2's `*Query` does not expose `Release()`, and resources
  are reclaimed automatically by the driver.

### Removed

- `Queryx.{Get,Set}RequestTimeout` — Scylla-only extensions absent in Apache v2.
- `Batch.{Get,Set}RequestTimeout` and `Batch.SetHostID` — same reason.
- `cmd/schemagen` index-model generation (`-ignore-indexes`, `Indexes`
  template block, `testdata/no_ignore_indexes/`). Apache v2's
  `KeyspaceMetadata` no longer carries `Indexes`. Per-column index data is
  still exposed by the driver and can be reintroduced later.

### Not ported

- `migrate/` subpackage — out of scope for v0.1.0.

### Supported versions

- Cassandra 4.1.x and 5.0.x (CI matrix). ScyllaDB is not in the matrix; use
  upstream gocqlx there.
