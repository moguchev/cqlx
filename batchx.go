// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package cqlx

import (
	"context"
	"fmt"

	gocqlv2 "github.com/apache/cassandra-gocql-driver/v2"
)

// Batch is a wrapper around gocqlv2.Batch
type Batch struct {
	*gocqlv2.Batch
	ctx context.Context
}

// ctxOrBackground returns the stored context or context.Background if none was set.
//
// Apache v2 deprecated Batch.WithContext and Session.Execute*Batch* in favor of
// passing context at execution time (Batch.ExecContext / ExecCASContext /
// MapExecCASContext). cqlx preserves the gocqlx-style WithContext / ContextBatch /
// Session.ExecuteBatch* API by holding the context on the wrapper and forwarding
// it here.
func (b *Batch) ctxOrBackground() context.Context {
	if b.ctx == nil {
		return context.Background()
	}
	return b.ctx
}

// NewBatch creates a new batch operation using defaults defined in the cluster.
//
// Deprecated: use session.Batch instead
func (s *Session) NewBatch(bt gocqlv2.BatchType) *Batch {
	return &Batch{
		Batch: s.Session.Batch(bt),
	}
}

// Batch creates a new batch operation using defaults defined in the cluster.
func (s *Session) Batch(bt gocqlv2.BatchType) *Batch {
	return &Batch{
		Batch: s.Session.Batch(bt),
	}
}

// ContextBatch creates a new batch operation using defaults defined in the cluster with attached context.
func (s *Session) ContextBatch(ctx context.Context, bt gocqlv2.BatchType) *Batch {
	return &Batch{
		Batch: s.Session.Batch(bt),
		ctx:   ctx,
	}
}

// BindStruct binds query named parameters to values from arg using a mapper.
// If value cannot be found an error is reported.
func (b *Batch) BindStruct(qry *Queryx, arg any) error {
	args, err := qry.bindStructArgs(arg, nil)
	if err != nil {
		return err
	}
	b.Query(qry.Statement(), args...)
	return nil
}

// Bind binds query parameters to values from args.
// If value cannot be found an error is reported.
func (b *Batch) Bind(qry *Queryx, args ...any) error {
	if len(qry.Names) != len(args) {
		return fmt.Errorf("query requires %d arguments, but %d provided", len(qry.Names), len(args))
	}
	b.Query(qry.Statement(), args...)
	return nil
}

// BindMap binds query named parameters to values from arg using a mapper.
// If value cannot be found an error is reported.
func (b *Batch) BindMap(qry *Queryx, arg map[string]any) error {
	args, err := qry.bindMapArgs(arg)
	if err != nil {
		return err
	}
	b.Query(qry.Statement(), args...)
	return nil
}

// BindStructMap binds query named parameters to values from arg0 and arg1 using a mapper.
// If value cannot be found an error is reported.
func (b *Batch) BindStructMap(qry *Queryx, arg0 any, arg1 map[string]any) error {
	args, err := qry.bindStructArgs(arg0, arg1)
	if err != nil {
		return err
	}
	b.Query(qry.Statement(), args...)
	return nil
}

// Consistency sets the consistency level for this batch. If no consistency
// level has been set, the default consistency level of the cluster is used.
func (b *Batch) Consistency(c gocqlv2.Consistency) *Batch {
	b.Batch.Consistency(c)
	return b
}

// DefaultTimestamp will enable the with default timestamp flag on the query.
// If enabled, this will replace the server side assigned
// timestamp as default timestamp. Note that a timestamp in the query itself
// will still override this timestamp. This is entirely optional.
//
// Only available on protocol >= 3
func (b *Batch) DefaultTimestamp(enable bool) *Batch {
	b.Batch.DefaultTimestamp(enable)
	return b
}

// Observer enables batch-level observer on this batch.
// The provided observer will be called every time this batched query is executed.
func (b *Batch) Observer(observer gocqlv2.BatchObserver) *Batch {
	b.Batch.Observer(observer)
	return b
}

// RetryPolicy sets the retry policy to use when executing the batch operation
func (b *Batch) RetryPolicy(policy gocqlv2.RetryPolicy) *Batch {
	b.Batch.RetryPolicy(policy)
	return b
}

// SerialConsistency sets the consistency level for the
// serial phase of conditional updates. That consistency can only be
// either SERIAL or LOCAL_SERIAL and if not present, it defaults to
// SERIAL. This option will be ignored for anything else that a
// conditional update/insert.
//
// Only available for protocol 3 and above
func (b *Batch) SerialConsistency(cons gocqlv2.Consistency) *Batch {
	b.Batch.SerialConsistency(cons)
	return b
}

// SetKeyspace will enable per-batch keyspace override (Cassandra 5.0+).
func (b *Batch) SetKeyspace(keyspace string) *Batch {
	b.Batch.SetKeyspace(keyspace)
	return b
}

// SpeculativeExecutionPolicy sets the speculative execution policy to use when executing the batch operation
func (b *Batch) SpeculativeExecutionPolicy(policy gocqlv2.SpeculativeExecutionPolicy) *Batch {
	b.Batch.SpeculativeExecutionPolicy(policy)
	return b
}

// Trace enables tracing of this batch. Look at the documentation of the
// gocqlv2.Tracer interface to learn more about tracing.
func (b *Batch) Trace(trace gocqlv2.Tracer) *Batch {
	b.Batch.Trace(trace)
	return b
}

// WithNowInSeconds sets the "now" value (in seconds) used by the server to
// evaluate TTLs and write times for this batch. Available on protocol >= 5.
func (b *Batch) WithNowInSeconds(now int) *Batch {
	b.Batch.WithNowInSeconds(now)
	return b
}

// WithContext returns a copy of b with its context set to ctx. The context
// controls the entire lifetime of executing the batch: it will be canceled and
// return once the context is canceled.
//
// Note: unlike gocqlx, the returned *Batch shares the underlying *gocql.Batch
// with the original — Apache v2 deprecated Batch.WithContext, and cqlx stores
// the context on the wrapper to keep the public API gocqlx-compatible. As a
// result, subsequent mutations (e.g. Query, RetryPolicy) on either copy are
// observed by both. See docs/api-deltas.md.
func (b *Batch) WithContext(ctx context.Context) *Batch {
	return &Batch{
		Batch: b.Batch,
		ctx:   ctx,
	}
}

// WithTimestamp will enable the with default timestamp flag on the query
// like DefaultTimestamp does. But also allows to define value for timestamp.
// It works the same way as USING TIMESTAMP in the query itself, but
// should not break prepared query optimization.
//
// Only available on protocol >= 3
func (b *Batch) WithTimestamp(timestamp int64) *Batch {
	b.Batch.WithTimestamp(timestamp)
	return b
}

// Query adds the query to the batch operation
func (b *Batch) Query(stmt string, args ...any) *Batch {
	b.Batch.Query(stmt, args...)
	return b
}

// ExecuteBatch executes a batch operation and returns nil if successful
// otherwise an error describing the failure.
func (s *Session) ExecuteBatch(batch *Batch) error {
	return batch.ExecContext(batch.ctxOrBackground())
}

// ExecuteBatchCAS executes a batch operation and returns true if successful and
// an iterator (to scan additional rows if more than one conditional statement)
// was sent.
// Further scans on the interator must also remember to include
// the applied boolean as the first argument to *Iter.Scan
func (s *Session) ExecuteBatchCAS(batch *Batch, dest ...any) (applied bool, iter *gocqlv2.Iter, err error) {
	return batch.ExecCASContext(batch.ctxOrBackground(), dest...)
}

// MapExecuteBatchCAS executes a batch operation much like ExecuteBatchCAS,
// however it accepts a map rather than a list of arguments for the initial
// scan.
func (s *Session) MapExecuteBatchCAS(batch *Batch, dest map[string]any) (applied bool, iter *gocqlv2.Iter, err error) {
	return batch.MapExecCASContext(batch.ctxOrBackground(), dest)
}
