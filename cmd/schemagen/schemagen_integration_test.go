// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

//go:build all || integration
// +build all integration

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/moguchev/cqlx/cqlxtest"
)

var flagUpdate = flag.Bool("update", false, "update golden file")

func TestSchemagen(t *testing.T) {
	flag.Parse()
	createTestSchema(t)

	// add ignored types and table
	*flagIgnoreNames = strings.Join([]string{
		"composers",
		"composers_by_name",
		"label",
	}, ",")

	b := runSchemagen(t, "schemagentest")
	assertDiff(t, b, "testdata/models.go")
}

func assertDiff(t *testing.T, actual []byte, goldenFile string) {
	t.Helper()

	if *flagUpdate {
		if err := os.WriteFile(goldenFile, actual, os.ModePerm); err != nil {
			t.Fatal(err)
		}
	}
	golden, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(string(golden), string(actual)); diff != "" {
		t.Fatal(diff)
	}
}

func createTestSchema(t *testing.T) {
	t.Helper()

	session := cqlxtest.CreateSession(t)
	defer session.Close()

	err := session.ExecStmt(`CREATE KEYSPACE IF NOT EXISTS schemagen WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`)
	if err != nil {
		t.Fatal("create keyspace:", err)
	}

	err = session.ExecStmt(`CREATE TABLE IF NOT EXISTS schemagen.songs (
		id uuid PRIMARY KEY,
		title text,
		album text,
		artist text,
		duration duration,
		tags set<text>,
		data blob)`)
	if err != nil {
		t.Fatal("create table:", err)
	}

	err = session.ExecStmt(`CREATE TYPE IF NOT EXISTS schemagen.album (
		name text,
		songwriters set<text>)`)
	if err != nil {
		t.Fatal("create type:", err)
	}

	err = session.ExecStmt(`CREATE TABLE IF NOT EXISTS schemagen.playlists (
		id uuid,
		title text,
		album frozen<album>,
		artist text,
		song_id uuid,
		PRIMARY KEY (id, title, album, artist))`)
	if err != nil {
		t.Fatal("create table:", err)
	}

	err = session.ExecStmt(`CREATE TABLE IF NOT EXISTS schemagen.composers (
		id uuid PRIMARY KEY,
		name text)`)
	if err != nil {
		t.Fatal("create table:", err)
	}

	err = session.ExecStmt(`CREATE MATERIALIZED VIEW IF NOT EXISTS schemagen.composers_by_name AS
    	SELECT id, name
    	FROM composers
    	WHERE id IS NOT NULL AND name IS NOT NULL
    	PRIMARY KEY (id, name)`)
	if err != nil {
		t.Fatal("create view:", err)
	}

	err = session.ExecStmt(`CREATE TYPE IF NOT EXISTS schemagen.label (
		name text,
		artists set<text>)`)
	if err != nil {
		t.Fatal("create type:", err)
	}
}

func runSchemagen(t *testing.T, pkgname string) []byte {
	t.Helper()

	dir, err := os.MkdirTemp("", "cqlx")
	if err != nil {
		t.Fatal(err)
	}
	keyspace := "schemagen"
	cl := "127.0.0.1"

	flagCluster = &cl
	flagKeyspace = &keyspace
	flagPkgname = &pkgname
	flagOutput = &dir

	if err := schemagen(); err != nil {
		t.Fatalf("schemagen() error %s", err)
	}

	f := fmt.Sprintf("%s/%s.go", dir, pkgname)
	b, err := os.ReadFile(f)
	if err != nil {
		t.Fatalf("%s: %s", f, err)
	}
	return b
}
