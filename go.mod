module github.com/upfluence/pkg

go 1.13

replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/certifi/gocertifi v0.0.0-20200211180108-c7c1fbc02894 // indirect
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/cyberdelia/statsd v0.0.0-20191230050547-9a74169bea7b
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/garyburd/redigo v1.6.0
	github.com/getsentry/raven-go v0.2.0
	github.com/gocql/gocql v0.0.0-20200624222514-34081eda590e
	github.com/golang/snappy v0.0.1
	github.com/jinzhu/gorm v1.9.14
	github.com/lib/pq v1.7.0
	github.com/mattes/migrate v3.0.2-0.20180508041624-4768a648fbd9+incompatible
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.6.1
	github.com/upfluence/base v0.1.51
	github.com/upfluence/cfg v0.1.0
	github.com/upfluence/log v0.0.0-20200124211732-c9875854d3b8
	github.com/upfluence/sql v0.2.6
	github.com/upfluence/stats v0.0.0-20200119200538-5dd0f0409179
	github.com/upfluence/thrift v2.1.0+incompatible
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/text v0.3.3
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
)
