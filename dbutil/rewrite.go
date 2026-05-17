// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package dbutil

import (
	"github.com/moguchev/cqlx"
	"github.com/moguchev/cqlx/table"
)

// RewriteTable rewrites src table to dst table.
// Rows can be transformed using the transform function.
// If row map is empty after transformation the row is skipped.
// Additional options can be passed to modify the insert query.
func RewriteTable(session cqlx.Session, dst, src *table.Table, transform func(map[string]any), options ...func(q *cqlx.Queryx)) error {
	insert := dst.InsertQuery(session)

	// Apply query options
	for _, o := range options {
		o(insert)
	}

	// Iterate over all rows and reinsert them to dst table
	iter := session.Query(src.SelectAll()).Iter()
	m := make(map[string]any)
	for iter.MapScan(m) {
		if transform != nil {
			transform(m)
		}
		if len(m) == 0 {
			continue // map is empty - no need to clean
		}
		if err := insert.BindMap(m).Exec(); err != nil {
			return err
		}
		m = map[string]any{}
	}
	return iter.Close()
}
