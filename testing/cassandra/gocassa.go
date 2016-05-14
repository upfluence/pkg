package cassandra

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/gocql/gocql"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/hailocab/gocassa"
	_ "github.com/upfluence/goutils/Godeps/_workspace/src/github.com/mattes/migrate/driver/cassandra"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/mattes/migrate/migrate"
)

const (
	defaultCassandraIP = "127.0.0.1"
	defaultKeyspace    = "test"
	cassandraPort      = 9042
	protocolVersion    = 3
)

func BuildKeySpace(migrationsPath *string) (gocassa.KeySpace, error) {
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

	conn := gocassa.NewConnection(gocassa.GoCQLSessionToQueryExecutor(session))

	conn.DropKeySpace(keySpace)
	conn.CreateKeySpace(keySpace)

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

			conn.Close()
			return nil, errors.New(strings.Join(strErrs, ","))
		}
	}

	ks := conn.KeySpace(keySpace)
	ks.DebugMode(true)

	return ks, nil
}
