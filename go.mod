module github.com/upfluence/pkg

go 1.18

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/cyberdelia/statsd v0.0.0-20191230050547-9a74169bea7b
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/garyburd/redigo v1.6.0
	github.com/gocql/gocql v0.0.0-20190629212933-1335d3dd7fe2
	github.com/golang/protobuf v1.5.0 // indirect
	github.com/golang/snappy v0.0.2-0.20190904063534-ff6b7dc882cf
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v1.0.1-0.20200216102404-196e6ce06ca4 // indirect
	github.com/jonboulle/clockwork v0.2.0 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/lib/pq v1.3.1-0.20200116171513-9eb3fc897d6f
	github.com/mattes/migrate v3.0.2-0.20180508041624-4768a648fbd9+incompatible
	github.com/mattn/go-sqlite3 v1.14.9
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/procfs v0.1.3 // indirect
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.6.1
	github.com/upfluence/errors v0.2.4
	github.com/upfluence/log v0.0.3
	github.com/upfluence/stats v0.0.0-20200119200538-5dd0f0409179
	github.com/upfluence/thrift v2.0.11+incompatible
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/text v0.3.2
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
)

require (
	github.com/upfluence/cfg v0.2.4
	golang.org/x/exp v0.0.0-20230321023759-10a507213a29
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
	golang.org/x/term v0.6.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/denisenkom/go-mssqldb v0.11.0 // indirect
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/getsentry/sentry-go v0.9.0 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gofrs/uuid v4.1.0+incompatible // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/jinzhu/now v1.1.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/prometheus/common v0.10.0 // indirect
	go.uber.org/atomic v1.6.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200605160147-a5ece683394c // indirect
)

replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
