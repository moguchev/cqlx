# Contributing to cqlx

Thanks for taking the time to contribute.

## Quick start

Requirements:

- Go 1.25+
- Docker (for the integration test suite)
- GNU make

```bash
git clone https://github.com/moguchev/cqlx.git
cd cqlx
make get-deps          # download Go modules
make get-tools         # install golangci-lint into $GOPATH/bin

make check             # run lint
make test-unit         # run pure unit tests (no docker)

# Integration suite — boots a local Cassandra container, runs everything
# tagged `all || integration`. Defaults to cassandra:5.0; set CASSANDRA_IMAGE
# to override.
make start-cassandra
make test
make stop-cassandra
```

## Pull requests

- Keep changes focused. One PR = one logical change.
- Add tests for new behavior. Unit tests for pure functions; integration tests
  (under `//go:build all || integration`) when you need a live Cassandra.
- Run `make check` and `make test-unit` before pushing.
- The CI job runs both unit and integration matrices (Cassandra 4.1 + 5.0) —
  expect to fix red CI before review.

## API compatibility

cqlx's public API tracks [scylladb/gocqlx](https://github.com/scylladb/gocqlx)
1:1, except for a small set of intentional divergences documented in
[`docs/api-deltas.md`](docs/api-deltas.md). Please do not introduce new
divergences without first opening an issue and getting agreement.

## Cherry-picking upstream gocqlx fixes

Because every file in cqlx differs from gocqlx by import path and header,
`git cherry-pick` from the upstream repo will always conflict. Instead:

1. Read the upstream commit and reproduce the change against cqlx manually.
2. Preserve authorship with a trailer: `Co-authored-by: Original Author <email>`.
3. Reference the upstream commit hash in the PR description.

We will add a `tools/import-upstream.sh` helper after v0.1.0 to automate the
import-path rewrite — see [issue tracker](https://github.com/moguchev/cqlx/issues)
for the current state.

## Releases

Tags follow semver. Pre-1.0 releases (`v0.x`) are not API-stable; we will bump
the minor version on any backwards-incompatible change and call it out in
`CHANGELOG.md`.
