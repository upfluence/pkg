package cassandrautil

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"

	"github.com/upfluence/pkg/cfg"
	"github.com/upfluence/pkg/log"
)

var defaultOptions = &options{
	keyspace:        cfg.FetchString("CASSANDRA_KEYSPACE", "test"),
	cassandraURL:    cfg.FetchString("CASSANDRA_URL", "127.0.0.1"),
	port:            9042,
	protocolVersion: 3,
	consistency:     gocql.Quorum,
	timeout:         15 * time.Second,
}

func Keyspace(k string) Option {
	return func(o *options) { o.keyspace = k }
}

func CassandraURL(url string) Option {
	return func(o *options) { o.cassandraURL = url }
}

func Consistency(c gocql.Consistency) Option {
	return func(o *options) { o.consistency = c }
}

func Timeout(t time.Duration) Option {
	return func(o *options) { o.timeout = t }
}

func Port(p int) Option {
	return func(o *options) { o.port = p }
}

type options struct {
	keyspace, cassandraURL string
	port, protocolVersion  int

	consistency gocql.Consistency
	timeout     time.Duration
}

func (o options) cassandraIPs() []string {
	return strings.Split(o.cassandraURL, ",")
}

type Option func(*options)

func BuildSession(opts ...Option) (*gocql.Session, error) {
	opt := *defaultOptions

	for _, optFn := range opts {
		optFn(&opt)
	}

	cluster := gocql.NewCluster(opt.cassandraIPs()...)

	cluster.Consistency = opt.consistency
	cluster.ProtoVersion = opt.protocolVersion
	cluster.Keyspace = opt.keyspace
	cluster.Timeout = opt.timeout

	return cluster.CreateSession()
}

func MustBuildSession(opts ...Option) *gocql.Session {
	var sess, err = BuildSession(opts...)

	if err != nil {
		log.Fatalf("cassandrautil: %v", err)
	}

	return sess
}

func CassandraURI(opts ...Option) string {
	opt := *defaultOptions

	for _, optFn := range opts {
		optFn(&opt)
	}

	return fmt.Sprintf(
		"cassandra://%s:%d/%s?protocol=%d",
		strings.Split(opt.cassandraURL, ",")[0],
		opt.port,
		opt.keyspace,
		opt.protocolVersion,
	)
}
