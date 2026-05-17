# Relationship to gocqlx

## TL;DR

cqlx is a derivative work of [scylladb/gocqlx](https://github.com/scylladb/gocqlx).
The public API matches gocqlx 1:1. The underlying CQL driver is the upstream
[apache/cassandra-gocql-driver/v2](https://github.com/apache/cassandra-gocql-driver),
not the Scylla fork. This means cqlx is consumable from any Go project without
a `replace github.com/gocql/gocql => github.com/scylladb/gocql` directive.

## Why a fork?

gocqlx is built on top of `github.com/gocql/gocql`, with the expectation that
downstream `go.mod` files carry:

```text
replace github.com/gocql/gocql => github.com/scylladb/gocql v1.x.x
```

This is incompatible with using the upstream Apache driver in the same module.
Many users want gocqlx's ergonomics (named params, struct binding, qb, table
CRUD, schemagen) against vanilla Apache Cassandra, without pulling in Scylla's
fork of the driver.

A build-tag / driver-agnostic interface in gocqlx itself was considered and
rejected — user code imports a single concrete `gocql` package, so the
abstraction never leaks usefully, would double the count of driver-touching
files, and would absorb the large v1.18 / v2 API shape difference inside the
library. A clean fork is simpler and honest.

## Why a fresh git history?

The Apache v2 driver and Scylla's gocql diverge enough on metadata types
(`KeyspaceMetadata.{Views,Types,Indexes}` → `MaterializedViews/UserTypes`,
`ColumnMetadata.Type` from `string` to `TypeInfo`, etc.) that nearly every
schemagen-touching file required a non-mechanical rewrite. Replaying upstream
history with rebased commits would have produced noisy, hard-to-review
intermediate states. Instead, cqlx starts with one `port from gocqlx@<sha>`
commit anchoring the lineage, and proceeds from there.

The [`NOTICE`](../NOTICE) file declares the upstream commit and license terms.

## What carries over verbatim?

- The public surface area: `Session`, `Queryx`, `Iterx`, `Batch`, `UDT`,
  `Transformer`, the `qb`, `table` and `dbutil` subpackages, and the
  `cmd/schemagen` CLI.
- Test conventions: integration tests gated by `//go:build all || integration`;
  `t.Parallel()` on independent unit tests.
- License: Apache 2.0, with dual-copyright header on every ported file.

## What is intentionally different?

See [api-deltas.md](api-deltas.md) for the exhaustive list. In short: a handful
of Scylla-only methods are dropped, `cmd/schemagen` no longer emits index
models, and the `migrate/` package is not ported.
