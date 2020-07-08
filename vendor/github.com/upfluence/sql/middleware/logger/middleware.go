package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/upfluence/log"
	"github.com/upfluence/log/record"

	"github.com/upfluence/sql"
)

type OpType string

const (
	Exec     OpType = "Exec"
	QueryRow OpType = "QueryRow"
	Query    OpType = "Query"
	Commit   OpType = "Commit"
	Rollback OpType = "Rollback"
)

type Logger interface {
	Log(OpType, string, []interface{}, time.Duration)
}

type simplifiedLogger struct {
	level  record.Level
	logger log.Logger
}

type durationField struct {
	d time.Duration
}

func (d *durationField) GetKey() string   { return "duration" }
func (d *durationField) GetValue() string { return fmt.Sprintf("%v", d.d) }

type dynamicField struct {
	name  string
	value interface{}
}

func (d *dynamicField) GetKey() string   { return d.name }
func (d *dynamicField) GetValue() string { return fmt.Sprintf("%v", d.value) }

func (l *simplifiedLogger) Log(_ OpType, q string, vs []interface{}, d time.Duration) {
	var fs = make([]record.Field, len(vs)+1)

	fs[0] = &durationField{d}

	for i, v := range vs {
		fs[i+1] = &dynamicField{name: fmt.Sprintf("$%d", i+1), value: v}
	}

	l.logger.WithFields(fs...).Log(l.level, q)
}

func NewFactory(l Logger) sql.MiddlewareFactory {
	return &factory{l: l}
}

func NewLevelFactory(l log.Logger, lvl record.Level) sql.MiddlewareFactory {
	return NewFactory(&simplifiedLogger{logger: l, level: lvl})
}

func NewDebugFactory(l log.Logger) sql.MiddlewareFactory {
	return NewLevelFactory(l, record.Debug)
}

type factory struct {
	l Logger
}

func (f *factory) Wrap(d sql.DB) sql.DB {
	return &db{queryer: &queryer{Queryer: d, l: f.l}, db: d}
}

type db struct {
	*queryer

	db sql.DB
}

func (d *db) Driver() string { return d.db.Driver() }

func (d *db) BeginTx(ctx context.Context) (sql.Tx, error) {
	var t, err = d.db.BeginTx(ctx)

	if err != nil {
		return nil, err
	}

	return &tx{queryer: &queryer{Queryer: t, l: d.queryer.l}, tx: t}, nil
}

type tx struct {
	*queryer

	tx sql.Tx
}

func (t *tx) Commit() error {
	var t0 = time.Now()

	defer t.queryer.logRequest(Commit, t0, "COMMIT", nil)

	return t.tx.Commit()
}

func (t *tx) Rollback() error {
	var t0 = time.Now()

	defer t.queryer.logRequest(Rollback, t0, "Rollback", nil)

	return t.tx.Rollback()
}

type queryer struct {
	sql.Queryer
	l Logger
}

func (q *queryer) logRequest(t OpType, t0 time.Time, qry string, vs []interface{}) {
	q.l.Log(t, qry, vs, time.Since(t0))
}

func (q *queryer) Exec(ctx context.Context, qry string, vs ...interface{}) (sql.Result, error) {
	var t0 = time.Now()

	defer q.logRequest(Exec, t0, qry, vs)

	return q.Queryer.Exec(ctx, qry, vs...)
}

func (q *queryer) QueryRow(ctx context.Context, qry string, vs ...interface{}) sql.Scanner {
	var t0 = time.Now()

	defer q.logRequest(QueryRow, t0, qry, vs)

	return q.Queryer.QueryRow(ctx, qry, vs...)
}

func (q *queryer) Query(ctx context.Context, qry string, vs ...interface{}) (sql.Cursor, error) {
	var t0 = time.Now()

	defer q.logRequest(QueryRow, t0, qry, vs)

	return q.Queryer.Query(ctx, qry, vs...)
}
