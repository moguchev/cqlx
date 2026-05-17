// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package main

import (
	"strconv"

	gocqlv2 "github.com/apache/cassandra-gocql-driver/v2"
)

// mapCqlTypeToGoType maps a CQL TypeInfo (Apache v2) to a Go type string.
func mapCqlTypeToGoType(t gocqlv2.TypeInfo) string {
	switch t.Type() {
	case gocqlv2.TypeAscii, gocqlv2.TypeText, gocqlv2.TypeVarchar, gocqlv2.TypeInet:
		return "string"
	case gocqlv2.TypeBigInt, gocqlv2.TypeVarint:
		return "int64"
	case gocqlv2.TypeBlob:
		return "[]byte"
	case gocqlv2.TypeBoolean:
		return "bool"
	case gocqlv2.TypeCounter:
		return "int"
	case gocqlv2.TypeDate, gocqlv2.TypeTimestamp:
		return "time.Time"
	case gocqlv2.TypeDecimal:
		return "inf.Dec"
	case gocqlv2.TypeDouble:
		return "float64"
	case gocqlv2.TypeFloat:
		return "float32"
	case gocqlv2.TypeInt:
		return "int32"
	case gocqlv2.TypeSmallInt:
		return "int16"
	case gocqlv2.TypeTinyInt:
		return "int8"
	case gocqlv2.TypeTime:
		return "time.Duration"
	case gocqlv2.TypeUUID, gocqlv2.TypeTimeUUID:
		return "[16]byte"
	case gocqlv2.TypeDuration:
		return "gocqlv2.Duration"
	case gocqlv2.TypeList, gocqlv2.TypeSet:
		if ct, ok := t.(gocqlv2.CollectionType); ok {
			return "[]" + mapCqlTypeToGoType(ct.Elem)
		}
		return "[]any"
	case gocqlv2.TypeMap:
		if ct, ok := t.(gocqlv2.CollectionType); ok {
			return "map[" + mapCqlTypeToGoType(ct.Key) + "]" + mapCqlTypeToGoType(ct.Elem)
		}
		return "map[any]any"
	case gocqlv2.TypeTuple:
		tt, ok := t.(gocqlv2.TupleTypeInfo)
		if !ok {
			return "any"
		}
		s := "struct {\n"
		for i, e := range tt.Elems {
			s += "\t\tField" + strconv.Itoa(i+1) + " " + mapCqlTypeToGoType(e) + "\n"
		}
		s += "\t}"
		return s
	case gocqlv2.TypeUDT:
		udt, ok := t.(gocqlv2.UDTTypeInfo)
		if !ok {
			return "any"
		}
		return camelize(udt.Name) + "UserType"
	}
	return "any"
}
