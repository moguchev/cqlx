// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package cqlxtest

import (
	"fmt"
	"testing"

	gocqlv2 "github.com/apache/cassandra-gocql-driver/v2"

	"github.com/moguchev/cqlx"
	"github.com/moguchev/cqlx/qb"
	"github.com/moguchev/cqlx/table"
)

// Song is a fixture struct used by integration tests and examples.
type Song struct {
	Title  string
	Album  string
	Artist string
	Tags   []string
	Data   []byte
	ID     gocqlv2.UUID
}

// PlaylistItem is a fixture struct used by integration tests and examples.
type PlaylistItem struct {
	Title  string
	Album  string
	Artist string
	ID     gocqlv2.UUID
	SongID gocqlv2.UUID
}

// MustParseUUID parses a UUID string and fails the program if it is invalid.
// It is intended only for use in tests/examples with fixed literal UUIDs.
func MustParseUUID(s string) gocqlv2.UUID {
	u, err := gocqlv2.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}

// BasicCreateAndPopulateKeyspace creates the example "songs" + "playlists"
// schema in the given keyspace and inserts one song and one playlist item.
//
// It exists in cqlxtest so that integration tests in other packages (root cqlx
// batch tests, examples/basic, etc.) can share the same seed without copying.
func BasicCreateAndPopulateKeyspace(t *testing.T, session cqlx.Session, keyspace string) {
	t.Helper()

	if err := session.ExecStmt(fmt.Sprintf(
		`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`,
		keyspace,
	)); err != nil {
		t.Fatal("create keyspace:", err)
	}

	if err := session.ExecStmt(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.songs (
		id uuid PRIMARY KEY,
		title text,
		album text,
		artist text,
		tags set<text>,
		data blob)`, keyspace)); err != nil {
		t.Fatal("create table:", err)
	}

	if err := session.ExecStmt(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.playlists (
		id uuid,
		title text,
		album text,
		artist text,
		song_id uuid,
		PRIMARY KEY (id, title, album, artist))`, keyspace)); err != nil {
		t.Fatal("create table:", err)
	}

	playlistTable := table.New(table.Metadata{
		Name:    fmt.Sprintf("%s.playlists", keyspace),
		Columns: []string{"id", "title", "album", "artist", "song_id"},
		PartKey: []string{"id"},
		SortKey: []string{"title", "album", "artist", "song_id"},
	})

	insertSong := qb.Insert(fmt.Sprintf("%s.songs", keyspace)).
		Columns("id", "title", "album", "artist", "tags", "data").Query(session)

	insertSong.BindStruct(Song{
		ID:     MustParseUUID("756716f7-2e54-4715-9f00-91dcbea6cf50"),
		Title:  "La Petite Tonkinoise",
		Album:  "Bye Bye Blackbird",
		Artist: "Joséphine Baker",
		Tags:   []string{"jazz", "2013"},
		Data:   []byte("music"),
	})
	if err := insertSong.ExecRelease(); err != nil {
		t.Fatal("insert song:", err)
	}

	insertPlaylist := playlistTable.InsertQuery(session)
	insertPlaylist.BindStruct(PlaylistItem{
		ID:     MustParseUUID("2cc9ccb7-6221-4ccb-8387-f22b6a1b354d"),
		Title:  "La Petite Tonkinoise",
		Album:  "Bye Bye Blackbird",
		Artist: "Joséphine Baker",
		SongID: MustParseUUID("756716f7-2e54-4715-9f00-91dcbea6cf50"),
	})
	if err := insertPlaylist.ExecRelease(); err != nil {
		t.Fatal("insert playlist:", err)
	}
}
