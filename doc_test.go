// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package cqlx_test

import (
	gocqlv2 "github.com/apache/cassandra-gocql-driver/v2"

	"github.com/moguchev/cqlx"
	"github.com/moguchev/cqlx/qb"
)

func ExampleSession() {
	cluster := gocqlv2.NewCluster("host")
	session, err := cqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		// handle error
	}

	builder := qb.Select("foo")
	session.Query(builder.ToCql())
}

func ExampleUDT() {
	// Just add cqlx.UDT to a type, no need to implement marshalling functions
	type FullName struct {
		cqlx.UDT
		FirstName string
		LastName  string
	}

	_ = FullName{}
}

func ExampleUDT_wraper() {
	type FullName struct {
		FirstName string
		LastName  string
	}

	// Create new UDT wrapper type
	type FullNameUDT struct {
		cqlx.UDT
		*FullName
	}

	_ = FullNameUDT{}
}
