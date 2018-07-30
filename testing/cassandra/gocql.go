package cassandra

import (
	"fmt"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/mattes/migrate"
	_ "github.com/mattes/migrate/database/cassandra"
	"github.com/mattes/migrate/source"

	"github.com/upfluence/pkg/cfg"
)

const (
	defaultTimeout  = time.Minute
	cassandraPort   = 9042
	protocolVersion = 3
)

var (
	cassandraIP = cfg.FetchString("CASSANDRA_IP", "127.0.0.1")
	keySpace    = cfg.FetchString("CASSANDRA_KEY_SPACE", "test")
)

func BuildKeySpace(t testing.TB, driver source.Driver, tables []string) *gocql.Session {
	cluster := gocql.NewCluster(cassandraIP)
	cluster.Consistency = gocql.All
	cluster.ProtoVersion = protocolVersion
	cluster.Timeout = defaultTimeout

	session, err := cluster.CreateSession()

	if err != nil {
		t.Errorf("cant create cql session: %v", err)
	}

	session.Query(
		`CREATE KEYSPACE ` + keySpace + ` WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }`,
	).Exec()

	if driver != nil {
		m, err := migrate.NewWithSourceInstance(
			"testing_source",
			driver,
			fmt.Sprintf(
				"cassandra://%s:%d/%s?protocol=%d",
				cassandraIP,
				cassandraPort,
				keySpace,
				protocolVersion,
			),
		)

		if err != nil {
			t.Errorf("cant open migrate: %v", err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			t.Errorf("cant run migration: %v", err)
		}
	}

	session.Close()

	cluster.Keyspace = keySpace

	session, err = cluster.CreateSession()

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for _, table := range tables {
		if err := session.Query(
			fmt.Sprintf("TRUNCATE %s", table),
		).Exec(); err != nil {
			t.Log(err)
		}
	}

	return session
}
