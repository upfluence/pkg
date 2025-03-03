module github.com/upfluence/pkg

go 1.23.0

toolchain go1.23.1

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/cyberdelia/statsd v1.0.0
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/garyburd/redigo v1.6.4
	github.com/gocql/gocql v1.6.0
	github.com/golang/snappy v0.0.4
	github.com/jinzhu/gorm v1.9.16
	github.com/jinzhu/inflection v1.0.1-0.20231016084002-bbe0a3e7399f // indirect
	github.com/jonboulle/clockwork v0.2.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lib/pq v1.10.9
	github.com/mattes/migrate v3.0.2-0.20180508041624-4768a648fbd9+incompatible
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.17.0
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/streadway/amqp v1.1.0
	github.com/stretchr/testify v1.8.4
	github.com/upfluence/cfg v0.3.5
	github.com/upfluence/errors v0.2.9
	github.com/upfluence/log v0.0.5
	github.com/upfluence/stats v0.1.4
	github.com/upfluence/thrift v2.6.8+incompatible
	golang.org/x/exp v0.0.0-20241108190413-2d47ceb2692f
	golang.org/x/oauth2 v0.19.0
	golang.org/x/sync v0.9.0
	golang.org/x/term v0.26.0
	golang.org/x/text v0.20.0
	golang.org/x/time v0.3.0
)

require github.com/upfluence/base v0.1.138

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/coreos/bbolt v1.3.4 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/denisenkom/go-mssqldb v0.11.0 // indirect
	github.com/getsentry/sentry-go v0.25.0 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/jinzhu/now v1.1.2 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	go.uber.org/atomic v1.6.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	golang.org/x/crypto v0.29.0 // indirect
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
