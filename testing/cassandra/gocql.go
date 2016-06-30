package cassandra

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/gocql/gocql"
	_ "github.com/upfluence/goutils/Godeps/_workspace/src/github.com/mattes/migrate/driver/cassandra"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/mattes/migrate/migrate"
)

const (
	defaultCassandraIP = "127.0.0.1"
	defaultKeyspace    = "test"
	cassandraPort      = 9042
	protocolVersion    = 3
)

func BuildKeySpace(migrationsPath *string) (*gocql.Session, error) {
	var (
		cassandraIP = os.Getenv("CASSANDRA_IP")
		keySpace    = os.Getenv("CASSANDRA_KEY_SPACE")
	)

	if cassandraIP == "" {
		cassandraIP = defaultCassandraIP
	}

	if keySpace == "" {
		keySpace = fmt.Sprintf("%s%d", defaultKeyspace, rand.Int31())
	}

	cluster := gocql.NewCluster(cassandraIP)
	cluster.Consistency = gocql.All
	cluster.ProtoVersion = protocolVersion

	session, err := cluster.CreateSession()

	if err != nil {
		return nil, err
	}

	defer session.Close()

	session.Query(`DROP KEYSPACE IF EXISTS ` + keySpace).Exec()
	session.Query(`CREATE KEYSPACE ` + keySpace).Exec()

	if migrationsPath != nil {
		errs, ok := migrate.UpSync(
			fmt.Sprintf(
				"cassandra://%s:%d/%s?protocol=%d",
				cassandraIP,
				cassandraPort,
				keySpace,
				protocolVersion,
			),
			*migrationsPath,
		)

		if !ok {
			strErrs := []string{}
			for _, migrationError := range errs {
				strErrs = append(strErrs, migrationError.Error())
			}

			return nil, errors.New(strings.Join(strErrs, ","))
		}
	}

	cluster.Keyspace = keySpace
	return cluster.CreateSession()
}
