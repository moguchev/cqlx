// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package cqlxtest

import (
	"flag"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	gocqlv2 "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/apache/cassandra-gocql-driver/v2/snappy"

	"github.com/moguchev/cqlx"
)

var (
	flagCluster      = flag.String("cluster", "127.0.0.1", "a comma-separated list of host:port tuples")
	flagKeyspace     = flag.String("keyspace", "cqlx_test", "keyspace name")
	flagProto        = flag.Int("proto", 0, "protcol version")
	flagCQL          = flag.String("cql", "3.0.0", "CQL version")
	flagRF           = flag.Int("rf", 1, "replication factor for test keyspace")
	flagRetry        = flag.Int("retries", 5, "number of times to retry queries")
	flagCompressTest = flag.String("compressor", "", "compressor to use")
	flagTimeout      = flag.Duration("gocqlv2.timeout", 5*time.Second, "sets the connection `timeout` for all operations")
)

// initOnce is per-process, so each test binary drops+creates the shared
// keyspace exactly once. Running multiple packages concurrently (e.g.
// `go test ./...` without `-p 1`) makes them race for the same keyspace and
// wipe each other's tables. Use `make test` / `make test-coverage`.
// TODO(release-after-v0.1.0): make keyspace name unique per package
// (e.g. `cqlx_test_<hash>`) so parallel `go test ./...` is safe.
var initOnce sync.Once

// CreateSession creates a new cqlx session from flags.
func CreateSession(tb testing.TB) cqlx.Session {
	tb.Helper()

	cluster := CreateCluster()
	return createSessionFromCluster(tb, cluster)
}

// CreateCluster creates gocql ClusterConfig from flags.
func CreateCluster() *gocqlv2.ClusterConfig {
	if !flag.Parsed() {
		flag.Parse()
	}
	clusterHosts := strings.Split(*flagCluster, ",")

	cluster := gocqlv2.NewCluster(clusterHosts...)
	cluster.ProtoVersion = *flagProto
	cluster.CQLVersion = *flagCQL
	cluster.Timeout = *flagTimeout
	cluster.Consistency = gocqlv2.Quorum
	cluster.MaxWaitSchemaAgreement = 2 * time.Minute // travis might be slow
	cluster.ReconnectionPolicy = &gocqlv2.ConstantReconnectionPolicy{
		MaxRetries: 10,
		Interval:   3 * time.Second,
	}
	if *flagRetry > 0 {
		cluster.RetryPolicy = &gocqlv2.SimpleRetryPolicy{NumRetries: *flagRetry}
	}

	switch *flagCompressTest {
	case "snappy":
		cluster.Compressor = &snappy.SnappyCompressor{}
	case "":
	default:
		panic("invalid compressor: " + *flagCompressTest)
	}

	return cluster
}

// CreateKeyspace creates keyspace with SimpleStrategy and RF derived from flags.
func CreateKeyspace(cluster *gocqlv2.ClusterConfig, keyspace string) error {
	c := *cluster
	c.Keyspace = "system"
	c.Timeout = 30 * time.Second

	session, err := cqlx.WrapSession(c.CreateSession())
	if err != nil {
		return err
	}
	defer session.Close()

	{
		err := session.ExecStmt(`DROP KEYSPACE IF EXISTS ` + keyspace)
		if err != nil {
			return fmt.Errorf("drop keyspace: %w", err)
		}
	}

	{
		err := session.ExecStmt(fmt.Sprintf(`CREATE KEYSPACE %s WITH replication = {'class' : 'SimpleStrategy', 'replication_factor' : %d}`, keyspace, *flagRF))
		if err != nil {
			return fmt.Errorf("create keyspace: %w", err)
		}
	}

	return nil
}

func createSessionFromCluster(tb testing.TB, cluster *gocqlv2.ClusterConfig) cqlx.Session {
	tb.Helper()
	if !flag.Parsed() {
		flag.Parse()
	}
	// Drop and re-create the keyspace once. Different tests should use their own
	// individual tables, but can assume that the table does not exist before.
	initOnce.Do(func() {
		if err := CreateKeyspace(cluster, *flagKeyspace); err != nil {
			tb.Fatal(err)
		}
	})

	cluster.Keyspace = *flagKeyspace
	session, err := cqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		tb.Fatal("CreateSession:", err)
	}
	return session
}
