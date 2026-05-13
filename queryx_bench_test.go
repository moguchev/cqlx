// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package cqlx_test

import (
	"testing"

	gocqlv2 "github.com/apache/cassandra-gocql-driver/v2"

	"github.com/moguchev/cqlx"
)

func BenchmarkCompileNamedQuery(b *testing.B) {
	q := []byte("INSERT INTO cycling.cyclist_name (id, user_uuid, firstname, stars) VALUES (:id, :user_uuid, :firstname, :stars)")
	b.ResetTimer()
	for b.Loop() {
		_, _, err := cqlx.CompileNamedQuery(q)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQueryxBindStruct(b *testing.B) {
	q := cqlx.Queryx{
		Names:  []string{"name", "age", "first", "last"},
		Mapper: cqlx.DefaultMapper,
		Query:  &gocqlv2.Query{},
	}
	type t struct {
		Name  string
		Age   int
		First string
		Last  string
	}
	am := t{"Jason Moiron", 30, "Jason", "Moiron"}
	b.ResetTimer()
	for b.Loop() {
		q.BindStruct(am)
	}
}

func BenchmarkBindMap(b *testing.B) {
	q := cqlx.Queryx{
		Names:  []string{"name", "age", "first", "last"},
		Mapper: cqlx.DefaultMapper,
		Query:  &gocqlv2.Query{},
	}
	am := map[string]any{
		"name":  "Jason Moiron",
		"age":   30,
		"first": "Jason",
		"last":  "Moiron",
	}
	b.ResetTimer()
	for b.Loop() {
		q.BindMap(am)
	}
}
