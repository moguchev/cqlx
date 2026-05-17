// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package main

import (
	"testing"

	gocqlv2 "github.com/apache/cassandra-gocql-driver/v2"
)

func Test_usedInTables(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		columnType gocqlv2.TypeInfo
		typeName   string
	}{
		"matches direct UDT": {
			columnType: gocqlv2.UDTTypeInfo{Name: "album"},
			typeName:   "album",
		},
		"matches first of two tuple elements": {
			columnType: gocqlv2.TupleTypeInfo{Elems: []gocqlv2.TypeInfo{
				gocqlv2.UDTTypeInfo{Name: "first"},
				gocqlv2.UDTTypeInfo{Name: "second"},
			}},
			typeName: "first",
		},
		"matches second of two tuple elements": {
			columnType: gocqlv2.TupleTypeInfo{Elems: []gocqlv2.TypeInfo{
				gocqlv2.UDTTypeInfo{Name: "first"},
				gocqlv2.UDTTypeInfo{Name: "second"},
			}},
			typeName: "second",
		},
		"matches nested tuple": {
			columnType: gocqlv2.TupleTypeInfo{Elems: []gocqlv2.TypeInfo{
				gocqlv2.UDTTypeInfo{Name: "outer"},
				gocqlv2.TupleTypeInfo{Elems: []gocqlv2.TypeInfo{
					gocqlv2.UDTTypeInfo{Name: "inner"},
				}},
			}},
			typeName: "inner",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tables := map[string]*gocqlv2.TableMetadata{
				"table": {Columns: map[string]*gocqlv2.ColumnMetadata{
					"column": {Type: tt.columnType},
				}},
			}
			if !usedInTables(tt.typeName, tables) {
				t.Fatal()
			}
		})
	}

	t.Run("doesn't panic with empty type name", func(t *testing.T) {
		t.Parallel()
		tables := map[string]*gocqlv2.TableMetadata{
			"table": {Columns: map[string]*gocqlv2.ColumnMetadata{
				"column": {Type: gocqlv2.UDTTypeInfo{Name: "album"}},
			}},
		}
		usedInTables("", tables)
	})
}
