// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

//go:build all || integration
// +build all integration

package cqlx_test

import (
	"reflect"
	"testing"

	gocqlv2 "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/google/go-cmp/cmp"

	"github.com/moguchev/cqlx"
	"github.com/moguchev/cqlx/cqlxtest"
	"github.com/moguchev/cqlx/qb"
)

func TestBatch(t *testing.T) {
	cluster := cqlxtest.CreateCluster()
	if err := cqlxtest.CreateKeyspace(cluster, "batch_test"); err != nil {
		t.Fatal("create keyspace:", err)
	}

	session, err := cqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		t.Fatal("create session:", err)
	}
	t.Cleanup(func() {
		session.Close()
	})

	cqlxtest.BasicCreateAndPopulateKeyspace(t, session, "batch_test")

	song := cqlxtest.Song{
		ID:     cqlxtest.MustParseUUID("60fc234a-8481-4343-93bb-72ecab404863"),
		Title:  "La Petite Tonkinoise",
		Album:  "Bye Bye Blackbird",
		Artist: "Joséphine Baker",
		Tags:   []string{"jazz"},
		Data:   []byte("music"),
	}
	playlist := cqlxtest.PlaylistItem{
		ID:     cqlxtest.MustParseUUID("6a6255d9-680f-4cb5-b9a2-27cf4a810344"),
		Title:  "La Petite Tonkinoise",
		Album:  "Bye Bye Blackbird",
		Artist: "Joséphine Baker",
		SongID: cqlxtest.MustParseUUID("60fc234a-8481-4343-93bb-72ecab404863"),
	}

	t.Run("batch inserts", func(t *testing.T) {
		tcases := []struct {
			name           string
			methodSong     func(*cqlx.Batch, *cqlx.Queryx, cqlxtest.Song) error
			methodPlaylist func(*cqlx.Batch, *cqlx.Queryx, cqlxtest.PlaylistItem) error
		}{
			{
				name: "BindStruct",
				methodSong: func(b *cqlx.Batch, q *cqlx.Queryx, song cqlxtest.Song) error {
					return b.BindStruct(q, song)
				},
				methodPlaylist: func(b *cqlx.Batch, q *cqlx.Queryx, playlist cqlxtest.PlaylistItem) error {
					return b.BindStruct(q, playlist)
				},
			},
			{
				name: "BindMap",
				methodSong: func(b *cqlx.Batch, q *cqlx.Queryx, song cqlxtest.Song) error {
					return b.BindMap(q, map[string]any{
						"id":     song.ID,
						"title":  song.Title,
						"album":  song.Album,
						"artist": song.Artist,
						"tags":   song.Tags,
						"data":   song.Data,
					})
				},
				methodPlaylist: func(b *cqlx.Batch, q *cqlx.Queryx, playlist cqlxtest.PlaylistItem) error {
					return b.BindMap(q, map[string]any{
						"id":      playlist.ID,
						"title":   playlist.Title,
						"album":   playlist.Album,
						"artist":  playlist.Artist,
						"song_id": playlist.SongID,
					})
				},
			},
			{
				name: "Bind",
				methodSong: func(b *cqlx.Batch, q *cqlx.Queryx, song cqlxtest.Song) error {
					return b.Bind(q, song.ID, song.Title, song.Album, song.Artist, song.Tags, song.Data)
				},
				methodPlaylist: func(b *cqlx.Batch, q *cqlx.Queryx, playlist cqlxtest.PlaylistItem) error {
					return b.Bind(q, playlist.ID, playlist.Title, playlist.Album, playlist.Artist, playlist.SongID)
				},
			},
			{
				name: "BindStructMap",
				methodSong: func(b *cqlx.Batch, q *cqlx.Queryx, song cqlxtest.Song) error {
					in := map[string]any{
						"title": song.Title,
						"album": song.Album,
					}
					return b.BindStructMap(q, struct {
						ID     gocqlv2.UUID
						Artist string
						Tags   []string
						Data   []byte
					}{
						ID:     song.ID,
						Artist: song.Artist,
						Tags:   song.Tags,
						Data:   song.Data,
					}, in)
				},
				methodPlaylist: func(b *cqlx.Batch, q *cqlx.Queryx, playlist cqlxtest.PlaylistItem) error {
					in := map[string]any{
						"title": playlist.Title,
						"album": playlist.Album,
					}
					return b.BindStructMap(q, struct {
						ID     gocqlv2.UUID
						Artist string
						SongID gocqlv2.UUID
					}{
						ID:     playlist.ID,
						Artist: playlist.Artist,
						SongID: playlist.SongID,
					},
						in,
					)
				},
			},
		}
		for _, tcase := range tcases {
			t.Run(tcase.name, func(t *testing.T) {
				insertSong := qb.Insert("batch_test.songs").
					Columns("id", "title", "album", "artist", "tags", "data").Query(session)
				insertPlaylist := qb.Insert("batch_test.playlists").
					Columns("id", "title", "album", "artist", "song_id").Query(session)
				selectSong := qb.Select("batch_test.songs").Where(qb.Eq("id")).Query(session)
				selectPlaylist := qb.Select("batch_test.playlists").Where(qb.Eq("id")).Query(session)
				deleteSong := qb.Delete("batch_test.songs").Where(qb.Eq("id")).Query(session)
				deletePlaylist := qb.Delete("batch_test.playlists").Where(qb.Eq("id")).Query(session)

				b := session.NewBatch(gocqlv2.LoggedBatch)

				if err = tcase.methodSong(b, insertSong, song); err != nil {
					t.Fatal("insert song:", err)
				}
				if err = tcase.methodPlaylist(b, insertPlaylist, playlist); err != nil {
					t.Fatal("insert playList:", err)
				}

				if err := session.ExecuteBatch(b); err != nil {
					t.Fatal("batch execution:", err)
				}

				// verify song was inserted
				var gotSong cqlxtest.Song
				if err := selectSong.BindStruct(song).Get(&gotSong); err != nil {
					t.Fatal("select song:", err)
				}
				if diff := cmp.Diff(gotSong, song); diff != "" {
					t.Errorf("expected %v song, got %v, diff: %q", song, gotSong, diff)
				}

				// verify playlist item was inserted
				var gotPlayList cqlxtest.PlaylistItem
				if err := selectPlaylist.BindStruct(playlist).Get(&gotPlayList); err != nil {
					t.Fatal("select playList:", err)
				}
				if diff := cmp.Diff(gotPlayList, playlist); diff != "" {
					t.Errorf("expected %v playList, got %v, diff: %q", playlist, gotPlayList, diff)
				}
				if err = deletePlaylist.BindStruct(playlist).Exec(); err != nil {
					t.Error("delete playlist:", err)
				}
				if err = deleteSong.BindStruct(song).Exec(); err != nil {
					t.Error("delete song:", err)
				}
			})
		}
	})
}

func TestBatchAllWrapped(t *testing.T) {
	t.Parallel()

	var (
		gocqlType = reflect.TypeOf((*gocqlv2.Batch)(nil))
		cqlxType  = reflect.TypeOf((*cqlx.Batch)(nil))
	)

	for i := range gocqlType.NumMethod() {
		m, ok := cqlxType.MethodByName(gocqlType.Method(i).Name)
		if !ok {
			t.Fatalf("Batch missing method %s", gocqlType.Method(i).Name)
		}

		for j := range m.Type.NumOut() {
			if m.Type.Out(j) == gocqlType {
				t.Errorf("Batch method %s not wrapped", m.Name)
			}
		}
	}
}
