// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package cqlx

import (
	"reflect"

	gocqlv2 "github.com/apache/cassandra-gocql-driver/v2"
)

// Transformer transforms the value of the named parameter to another value.
type Transformer func(name string, val any) any

// DefaultBindTransformer just do nothing.
//
// A custom transformer can always be set per Query.
var DefaultBindTransformer Transformer

// UnsetEmptyTransformer unsets all empty parameters.
// It helps to avoid tombstones when using the same insert/update
// statement for filled and partially filled named parameters.
var UnsetEmptyTransformer = func(name string, val any) any {
	v := reflect.ValueOf(val)
	if v.IsZero() {
		return gocqlv2.UnsetValue
	}
	return val
}
