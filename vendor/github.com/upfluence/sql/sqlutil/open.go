package sqlutil

import (
	stdsql "database/sql"
	"errors"
	"sync"
	"time"

	"github.com/upfluence/sql"
	"github.com/upfluence/sql/backend/postgres"
	"github.com/upfluence/sql/backend/replication"
	"github.com/upfluence/sql/backend/roundrobin"
	"github.com/upfluence/sql/backend/simple"
	"github.com/upfluence/sql/sqlparser"
)

var (
	defaultOptions = &builder{parser: sqlparser.DefaultSQLParser()}

	ErrNoDBProvided = errors.New("sql/sqlutil: No DB provided")

	driverWrapperMu = &sync.Mutex{}
	driverWrapper   = map[string]driverWrapperFunc{"postgres": postgres.NewDB}
)

type driverWrapperFunc func(sql.DB, sqlparser.SQLParser) sql.DB

type DBOption func(*dbInput)

func WithMaxIdleConns(v int) DBOption {
	return func(i *dbInput) {
		v := v
		i.maxIdleConns = &v
	}
}

func WithMaxOpenConns(v int) DBOption {
	return func(i *dbInput) {
		v := v
		i.maxOpenConns = &v
	}
}

func WithConnMaxLifetime(v time.Duration) DBOption {
	return func(i *dbInput) {
		v := v
		i.maxLifetime = &v
	}
}

type dbInput struct {
	isMaster bool

	driver string
	uri    string

	maxIdleConns *int
	maxOpenConns *int
	maxLifetime  *time.Duration
}

func (i *dbInput) prepareDB(db *stdsql.DB) {
	if i.maxIdleConns != nil {
		db.SetMaxIdleConns(*i.maxIdleConns)
	}

	if i.maxOpenConns != nil {
		db.SetMaxOpenConns(*i.maxOpenConns)
	}

	if i.maxLifetime != nil {
		db.SetConnMaxLifetime(*i.maxLifetime)
	}
}

func (i *dbInput) buildDB(p sqlparser.SQLParser) (sql.DB, error) {
	var plainDB, err = stdsql.Open(i.driver, i.uri)

	if err != nil {
		return nil, err
	}

	i.prepareDB(plainDB)

	db := simple.FromStdDB(plainDB, i.driver)

	if wfn, ok := driverWrapper[i.driver]; ok {
		db = wfn(db, p)
	}

	return db, nil
}

type dbs []sql.DB

func (dbs dbs) buildDB() sql.DB {
	if len(dbs) == 1 {
		return dbs[0]
	}

	return roundrobin.NewDB(dbs...)
}

func (b *builder) buildDB() (sql.DB, error) {
	switch len(b.dbs) {
	case 0:
		return nil, ErrNoDBProvided
	case 1:
		return b.dbs[0].buildDB(b.parser)
	}

	var masters, slaves []sql.DB

	for _, i := range b.dbs {
		db, err := i.buildDB(b.parser)

		if err != nil {
			return nil, err
		}

		if i.isMaster {
			masters = append(masters, db)
		} else {
			slaves = append(slaves, db)
		}
	}

	if len(masters) == 0 || len(slaves) == 0 {
		return roundrobin.NewDB(append(masters, slaves...)...), nil
	}

	return replication.NewDB(
		dbs(masters).buildDB(),
		dbs(slaves).buildDB(),
		b.parser,
	), nil
}

type builder struct {
	dbs         []*dbInput
	middlewares []sql.MiddlewareFactory

	parser sqlparser.SQLParser
}

type Option func(*builder)

func WithDatabase(driver, dsn string, readOnly bool, opts ...DBOption) Option {
	i := dbInput{driver: driver, uri: dsn, isMaster: !readOnly}

	for _, opt := range opts {
		opt(&i)
	}

	return func(b *builder) { b.dbs = append(b.dbs, &i) }
}

func WithMaster(driver, dsn string, opts ...DBOption) Option {
	return WithDatabase(driver, dsn, false, opts...)
}

func WithSlave(driver, dsn string, opts ...DBOption) Option {
	return WithDatabase(driver, dsn, true, opts...)
}

func WithMiddleware(f sql.MiddlewareFactory) Option {
	return func(b *builder) { b.middlewares = append(b.middlewares, f) }
}

func Open(os ...Option) (sql.DB, error) {
	var opts = *defaultOptions

	for _, o := range os {
		o(&opts)
	}

	db, err := opts.buildDB()

	if err != nil {
		return nil, err
	}

	for _, m := range opts.middlewares {
		db = m.Wrap(db)
	}

	return db, nil
}
