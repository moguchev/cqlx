// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package cqlx

import (
	"context"

	gocqlv2 "github.com/apache/cassandra-gocql-driver/v2"
)

// This file contains wrappers around gocqlv2.Query that make Queryx expose the
// same interface but return *Queryx, this should be inlined by compiler.

// Consistency sets the consistency level for this query. If no consistency
// level have been set, the default consistency level of the cluster
// is used.
func (q *Queryx) Consistency(c gocqlv2.Consistency) *Queryx {
	q.Query.Consistency(c)
	return q
}

// CustomPayload sets the custom payload level for this query.
func (q *Queryx) CustomPayload(customPayload map[string][]byte) *Queryx {
	q.Query.CustomPayload(customPayload)
	return q
}

// Trace enables tracing of this query. Look at the documentation of the
// Tracer interface to learn more about tracing.
func (q *Queryx) Trace(trace gocqlv2.Tracer) *Queryx {
	q.Query.Trace(trace)
	return q
}

// Observer enables query-level observer on this query.
// The provided observer will be called every time this query is executed.
func (q *Queryx) Observer(observer gocqlv2.QueryObserver) *Queryx {
	q.Query.Observer(observer)
	return q
}

// PageSize will tell the iterator to fetch the result in pages of size n.
// This is useful for iterating over large result sets, but setting the
// page size too low might decrease the performance. This feature is only
// available in Cassandra 2 and onwards.
func (q *Queryx) PageSize(n int) *Queryx {
	q.Query.PageSize(n)
	return q
}

// DefaultTimestamp will enable the with default timestamp flag on the query.
// If enable, this will replace the server side assigned
// timestamp as default timestamp. Note that a timestamp in the query itself
// will still override this timestamp. This is entirely optional.
//
// Only available on protocol >= 3
func (q *Queryx) DefaultTimestamp(enable bool) *Queryx {
	q.Query.DefaultTimestamp(enable)
	return q
}

// WithTimestamp will enable the with default timestamp flag on the query
// like DefaultTimestamp does. But also allows to define value for timestamp.
// It works the same way as USING TIMESTAMP in the query itself, but
// should not break prepared query optimization
//
// Only available on protocol >= 3
func (q *Queryx) WithTimestamp(timestamp int64) *Queryx {
	q.Query.WithTimestamp(timestamp)
	return q
}

// RoutingKey sets the routing key to use when a token aware connection
// pool is used to optimize the routing of this query.
func (q *Queryx) RoutingKey(routingKey []byte) *Queryx {
	q.Query.RoutingKey(routingKey)
	return q
}

// WithContext stores ctx on q. The context controls the entire lifetime of
// executing a query: queries will be canceled and return once the context is
// canceled. Subsequent Exec/Iter/Scan call the underlying ExecContext/
// IterContext/ScanContext driver methods with the stored context.
//
// Note: unlike gocqlx, this does not create a shallow copy of the embedded
// *gocql.Query. Apache v2 deprecated Query.WithContext in favor of passing
// context at execution time; cqlx stores the context on the wrapper to keep
// the public API gocqlx-compatible. See docs/api-deltas.md.
func (q *Queryx) WithContext(ctx context.Context) *Queryx {
	if ctx == nil {
		ctx = context.Background()
	}
	q.ctx = ctx
	return q
}

// Prefetch sets the default threshold for pre-fetching new pages. If
// there are only p*pageSize rows remaining, the next page will be requested
// automatically.
func (q *Queryx) Prefetch(p float64) *Queryx {
	q.Query.Prefetch(p)
	return q
}

// RetryPolicy sets the policy to use when retrying the query.
func (q *Queryx) RetryPolicy(r gocqlv2.RetryPolicy) *Queryx {
	q.Query.RetryPolicy(r)
	return q
}

// SetSpeculativeExecutionPolicy sets the execution policy.
func (q *Queryx) SetSpeculativeExecutionPolicy(sp gocqlv2.SpeculativeExecutionPolicy) *Queryx {
	q.Query.SetSpeculativeExecutionPolicy(sp)
	return q
}

// Idempotent marks the query as being idempotent or not depending on
// the value.
func (q *Queryx) Idempotent(value bool) *Queryx {
	q.Query.Idempotent(value)
	return q
}

// SerialConsistency sets the consistency level for the
// serial phase of conditional updates. That consistency can only be
// either SERIAL or LOCAL_SERIAL and if not present, it defaults to
// SERIAL. This option will be ignored for anything else that a
// conditional update/insert.
func (q *Queryx) SerialConsistency(cons gocqlv2.SerialConsistency) *Queryx {
	q.Query.SerialConsistency(cons)
	return q
}

// PageState sets the paging state for the query to resume paging from a specific
// point in time. Setting this will disable to query paging for this query, and
// must be used for all subsequent pages.
func (q *Queryx) PageState(state []byte) *Queryx {
	q.Query.PageState(state)
	return q
}

// NoSkipMetadata will override the internal result metadata cache so that the driver does not
// send skip_metadata for queries, this means that the result will always contain
// the metadata to parse the rows and will not reuse the metadata from the prepared
// staement. This should only be used to work around cassandra bugs, such as when using
// CAS operations which do not end in Cas.
//
// See https://issues.apache.org/jira/browse/CASSANDRA-11099
func (q *Queryx) NoSkipMetadata() *Queryx {
	q.Query.NoSkipMetadata()
	return q
}

// SetHostID allows to define the host the query should be executed against. If the
// host was filtered or otherwise unavailable, then the query will error. If an empty
// string is sent, the default behavior, using the configured HostSelectionPolicy will
// be used. A hostID can be obtained from HostInfo.HostID() after calling GetHosts().
func (q *Queryx) SetHostID(hostID string) *Queryx {
	q.Query.SetHostID(hostID)
	return q
}

// SetKeyspace will enable per-query keyspace override (Cassandra 5.0+).
func (q *Queryx) SetKeyspace(keyspace string) *Queryx {
	q.Query.SetKeyspace(keyspace)
	return q
}

// WithNowInSeconds sets the "now" value (in seconds) used by the server to
// evaluate TTLs and write times for this query. Available on protocol >= 5.
func (q *Queryx) WithNowInSeconds(now int) *Queryx {
	q.Query.WithNowInSeconds(now)
	return q
}
