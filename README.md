# cqlx

[![Build](https://github.com/moguchev/cqlx/actions/workflows/main.yml/badge.svg)](https://github.com/moguchev/cqlx/actions/workflows/main.yml)
[![GoDoc](https://pkg.go.dev/badge/github.com/moguchev/cqlx)](https://pkg.go.dev/github.com/moguchev/cqlx)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

A productivity toolkit for [Apache Cassandra](https://cassandra.apache.org/) on top of the upstream
[apache/cassandra-gocql-driver/v2](https://github.com/apache/cassandra-gocql-driver) — named parameters,
struct binding, a query builder, table CRUD and schema generation, all with no `replace` directives in
your `go.mod`.

> **Status:** alpha (v0.x). API tracks gocqlx 1:1 but the driver underneath is the upstream Apache one,
> not the Scylla fork. Expect occasional breakage as v2 of the Apache driver evolves.

## Install

```bash
go get github.com/moguchev/cqlx@latest
```

## Quickstart

Run Cassandra locally:

```bash
docker run --name cqlx-cassandra -p 9042:9042 --rm -d cassandra:5.0
```

Connect and use:

```go
package main

import (
	"context"
	"log"

	"github.com/apache/cassandra-gocql-driver/v2"
	"github.com/moguchev/cqlx"
	"github.com/moguchev/cqlx/qb"
)

type Song struct {
	ID     gocql.UUID
	Title  string
	Album  string
	Artist string
}

func main() {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "examples"
	cluster.Consistency = gocql.Quorum

	session, err := cqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	s := Song{ID: gocql.TimeUUID(), Title: "T", Album: "A", Artist: "X"}
	stmt, names := qb.Insert("songs").Columns("id", "title", "album", "artist").ToCql()
	if err := session.ContextQuery(context.Background(), stmt, names).BindStruct(s).ExecRelease(); err != nil {
		log.Fatal(err)
	}
}
```

## Subpackages

- `cqlx` — root: sessions, queries, iterators, batches, struct binding, UDT.
- `qb` — fluent query builder for `SELECT/INSERT/UPDATE/DELETE/BATCH`.
- `table` — table-scoped CRUD helpers built on `qb`.
- `dbutil` — small helpers for keyspace/table introspection and rewrites.
- `cmd/schemagen` — generates Go models from a live keyspace.
- `cqlxtest` — test helpers used by the integration suite.

## Compatibility

Every release is exercised by the [Build workflow](.github/workflows/main.yml) against the
matrix below. A `✓` means both the unit suite (`make test-unit`) and the integration suite
(`make test`, `-tags all`) pass on the listed configuration.

Overall CI status:
[![Build](https://github.com/moguchev/cqlx/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/moguchev/cqlx/actions/workflows/main.yml?query=branch%3Amain)

| Cassandra image | Driver                             | Go   | OS             | CI job                                                                                                            |
| --------------- | ---------------------------------- | ---- | -------------- | ----------------------------------------------------------------------------------------------------------------- |
| `cassandra:4.1` | `apache/cassandra-gocql-driver/v2` | 1.25 | ubuntu-latest  | [Integration (Cassandra 4.1)](https://github.com/moguchev/cqlx/actions/workflows/main.yml?query=branch%3Amain)    |
| `cassandra:5.0` | `apache/cassandra-gocql-driver/v2` | 1.25 | ubuntu-latest  | [Integration (Cassandra 5.0)](https://github.com/moguchev/cqlx/actions/workflows/main.yml?query=branch%3Amain)    |
| ScyllaDB *      | `apache/cassandra-gocql-driver/v2` | —    | —              | not tested                                                                                                        |

\* For Scylla, use the upstream [`scylladb/gocqlx`](https://github.com/scylladb/gocqlx) — that's what cqlx
is forked from. cqlx exists because gocqlx requires
`replace github.com/gocql/gocql => github.com/scylladb/gocql`, which is incompatible with the
upstream Apache driver.

To run the matrix locally:

```bash
CASSANDRA_IMAGE=cassandra:4.1 make test
CASSANDRA_IMAGE=cassandra:5.0 make test
```

## Relationship to gocqlx

cqlx is a derivative work of [scylladb/gocqlx](https://github.com/scylladb/gocqlx). The public API matches
1:1 except for a handful of Scylla-only methods that are intentionally omitted (`Query.{Get,Set}RequestTimeout`,
`Batch.{Get,Set}RequestTimeout`, `Batch.SetHostID`) and `cmd/schemagen` adjustments needed for upstream
driver metadata. See [`docs/api-deltas.md`](docs/api-deltas.md) and [`docs/relationship-to-gocqlx.md`](docs/relationship-to-gocqlx.md).

The `migrate/` package from gocqlx is **not** ported.

## License

Apache 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
