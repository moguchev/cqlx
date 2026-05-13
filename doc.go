// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

// Package cqlx makes working with Cassandra and Scylla easy and error prone without sacrificing performance.
// It’s inspired by Sqlx, a tool for working with SQL databases, but it goes beyond what Sqlx provides.
//
// cqlx is built on top of github.com/apache/cassandra-gocql-driver/v2 — the upstream Apache driver — and
// is a fork of github.com/scylladb/gocqlx with an identical public API.
//
// For more details consult README.
package cqlx
