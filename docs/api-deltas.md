# API deltas relative to gocqlx

This document tracks every intentional difference between cqlx's public surface
and the upstream [scylladb/gocqlx](https://github.com/scylladb/gocqlx). Anything
not listed here is identical to gocqlx and can be migrated by changing the
import path.

## Removed methods (Scylla-only)

Apache v2 does not expose these methods. Calls would not compile in cqlx.

| Symbol                          | Why removed                                              |
| ------------------------------- | -------------------------------------------------------- |
| `Queryx.GetRequestTimeout`      | `Query.GetRequestTimeout` is a Scylla extension; absent in Apache v2. |
| `Queryx.SetRequestTimeout`      | Same as above.                                           |
| `Batch.GetRequestTimeout`       | Same as above on `Batch`.                                |
| `Batch.SetRequestTimeout`       | Same as above.                                           |
| `Batch.SetHostID`               | Not exposed on `Batch` in Apache v2.                     |

`Queryx.SetHostID` **is** preserved because Apache v2 keeps `Query.SetHostID`.

## Changed semantics

| Method                              | Change |
| ----------------------------------- | ------ |
| `Queryx.ExecRelease`                | Apache v2 has no `Query.Release()`. The method is kept for API symmetry but is now a no-op wrapper that simply calls `Exec`. Same for `ExecCASRelease`, `GetRelease`, `GetCASRelease`, `SelectRelease`. The driver releases the underlying query resources automatically. |
| `Queryx.WithContext` / `Batch.WithContext` / `Session.ContextQuery` / `Session.ContextBatch` | Store the context on the cqlx wrapper instead of delegating to the driver's `Query.WithContext` / `Batch.WithContext`. Executing methods (`Queryx.Exec`, `Iter`, `Scan`, `Session.ExecuteBatch`, `ExecuteBatchCAS`, `MapExecuteBatchCAS`) forward the stored context to the corresponding `*Context` variants of Apache v2 (`ExecContext` / `IterContext` / `ScanContext` / `ExecCASContext` / `MapExecCASContext`). Reason: Apache v2 deprecated `Query.WithContext`, `Batch.WithContext`, `Session.ExecuteBatch`, `Session.ExecuteBatchCAS`, and `Session.MapExecuteBatchCAS` in favor of passing the context at execution time; cqlx keeps the gocqlx-compatible public surface by moving the context-holding one level up. Visible consequence: `Batch.WithContext(ctx)` returns a wrapper that **shares** the underlying `*gocql.Batch` with the original — mutations (`Query(...)`, `RetryPolicy`, etc.) on either copy are observed by both. In gocqlx this used to be a shallow copy of the driver's batch. In practice `WithContext` is called right after construction, before any mutation, so this divergence rarely matters. |

## Added wrappers

Apache v2 added methods on `*Query` that did not exist in Scylla's gocql.
cqlx wraps them so `*Queryx` remains a strict superset of `*Query`:

- `Queryx.SetKeyspace`
- `Queryx.WithNowInSeconds`

(`Queryx.SetHostID` was already wrapped in gocqlx.)

Symmetrically, Apache v2 added the same methods on `*Batch`. cqlx wraps them
so `*Batch` chained calls keep returning `*cqlx.Batch`:

- `Batch.SetKeyspace`
- `Batch.WithNowInSeconds`

`Batch.Consistency` is also wrapped in cqlx (it was not wrapped in upstream
gocqlx) so that chained calls return `*cqlx.Batch` instead of the underlying
`*gocql.Batch`. The invariant — every `*gocql.Batch` method whose return type
is `*gocql.Batch` must be re-wrapped on `*cqlx.Batch` — is enforced by
`TestBatchAllWrapped`.

## cmd/schemagen

The `gocql.KeyspaceMetadata` shape changed in Apache v2; this propagates
through schemagen.

- **`-ignore-indexes` flag removed.** Apache v2 dropped the `Indexes` field
  from `KeyspaceMetadata` (per-column metadata still has `Index`). Schemagen
  no longer emits "Index models" / "Index structs" blocks. The `IgnoreIndexes`
  subtest and `testdata/no_ignore_indexes/` were deleted accordingly.
- **`mapScyllaToGoType(string) string` → `mapCqlTypeToGoType(gocql.TypeInfo) string`.**
  Apache v2 represents column types as a typed interface, not a CQL string, so
  the function now switches on `TypeInfo.Type()` and recurses into
  `*CollectionType` / `TupleTypeInfo` / `UDTTypeInfo` instead of parsing
  strings with regex.
- **Views walk via `BaseTable`.** `MaterializedViewMetadata` in Apache v2 does
  not carry `OrderedColumns/PartitionKey/ClusteringColumns/Columns` directly;
  those live on the embedded `*TableMetadata` reachable via `BaseTable`. The
  `keyspace.tmpl` views block iterates `.BaseTable.{...}` accordingly, and
  `.ViewName` is now `.Name`.
- **Type aliases.** `KeyspaceMetadata.Views` → `MaterializedViews`,
  `KeyspaceMetadata.Types` → `UserTypes`.

## Scylla-only `qb` clauses kept for gocqlx parity

These query-builder methods exist on the public API for source-level
compatibility with upstream gocqlx, but they generate CQL clauses that **only
ScyllaDB understands**. Executing the resulting statement against Apache
Cassandra (4.1 or 5.0) raises a `SyntaxException` at the server. They are kept
because removing them would break gocqlx-to-cqlx code migration; they are
flagged in godoc on each method.

| Symbol | Generated clause | Status on Apache Cassandra |
| ------ | ---------------- | -------------------------- |
| `qb.SelectBuilder.BypassCache` | `BYPASS CACHE` | Not supported. |
| `qb.SelectBuilder.Timeout` / `TimeoutNamed` | `USING TIMEOUT …` | Not supported (CASSANDRA-15193, Won't Fix). |
| `qb.InsertBuilder.Timeout` / `TimeoutNamed` | `USING TIMEOUT …` | Same. |
| `qb.UpdateBuilder.Timeout` / `TimeoutNamed` | `USING TIMEOUT …` | Same. |
| `qb.DeleteBuilder.Timeout` / `TimeoutNamed` | `USING TIMEOUT …` | Same. |
| `qb.BatchBuilder.Timeout` / `TimeoutNamed` | `USING TIMEOUT …` | Same. |

`USING TTL` and `USING TIMESTAMP` are standard CQL and continue to work
normally — only `USING TIMEOUT` is Scylla-specific.

If a future release decides to break gocqlx parity in favour of "everything in
the public API must work on the CI matrix", these methods are the candidates
to remove.

## Not ported

- **`migrate/`** package — out of scope for v0.1.0. Use upstream gocqlx if you
  rely on it.

## Driver-clean subpackages

`qb`, `table`, and `dbutil` are byte-identical to their gocqlx originals modulo
import-path and dual-copyright header — they touch only `cqlx.Session` /
`cqlx.Queryx`, never the underlying driver directly. One exception:
`dbutil.RewriteTable` no longer calls `defer insert.Release()` (see the
"Changed semantics" row above).

## Supported Cassandra versions

CI exercises both `cassandra:4.1` and `cassandra:5.0` images on every push.
ScyllaDB is **not** part of the matrix; for Scylla, use upstream gocqlx.
