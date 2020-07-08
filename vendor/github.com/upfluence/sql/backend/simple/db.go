package simple

import (
	"context"
	stdsql "database/sql"
	"sync"

	"github.com/upfluence/sql"
)

type db struct {
	*queryer

	db     *stdsql.DB
	driver string
}

func FromStdDB(stdDB *stdsql.DB, driver string) sql.DB {
	return &db{queryer: &queryer{stdDB}, db: stdDB, driver: driver}
}

func NewDB(driver, uri string) (sql.DB, error) {
	var plainDB, err = stdsql.Open(driver, uri)

	if err != nil {
		return nil, err
	}

	return FromStdDB(plainDB, driver), nil
}

type tx struct {
	sync.Mutex
	ctx context.Context

	ch chan struct{}

	q  *queryer
	tx *stdsql.Tx
}

func (tx *tx) Commit() error {
	select {
	case <-tx.ctx.Done():
		return tx.ctx.Err()
	case tx.ch <- struct{}{}:
	}

	err := tx.tx.Commit()
	<-tx.ch

	return err
}

func (tx *tx) Rollback() error {
	select {
	case <-tx.ctx.Done():
		return tx.ctx.Err()
	case tx.ch <- struct{}{}:
	}

	err := tx.tx.Rollback()
	<-tx.ch

	return err
}

func (tx *tx) Exec(ctx context.Context, qry string, vs ...interface{}) (sql.Result, error) {
	select {
	case <-tx.ctx.Done():
		return nil, tx.ctx.Err()
	case <-ctx.Done():
		return nil, ctx.Err()
	case tx.ch <- struct{}{}:
	}

	res, err := tx.q.ExecContext(ctx, qry, sql.StripReturningFields(vs)...)
	<-tx.ch

	return res, err
}

type scanner struct {
	s  sql.Scanner
	tx *tx
}

func (s *scanner) Scan(vs ...interface{}) error {
	err := s.s.Scan(vs...)

	<-s.tx.ch

	return err
}

type errScanner struct {
	error
}

func (es errScanner) Scan(...interface{}) error {
	return es.error
}

func (tx *tx) QueryRow(ctx context.Context, qry string, vs ...interface{}) sql.Scanner {
	select {
	case <-tx.ctx.Done():
		return errScanner{tx.ctx.Err()}
	case <-ctx.Done():
		return errScanner{ctx.Err()}
	case tx.ch <- struct{}{}:
	}

	return &scanner{
		s:  tx.q.QueryRowContext(ctx, qry, vs...),
		tx: tx,
	}
}

type cursor struct {
	sql.Cursor
	tx *tx
}

func (c *cursor) Close() error {
	err := c.Cursor.Close()

	<-c.tx.ch
	return err
}

func (tx *tx) Query(ctx context.Context, qry string, vs ...interface{}) (sql.Cursor, error) {
	select {
	case <-tx.ctx.Done():
		return nil, tx.ctx.Err()
	case <-ctx.Done():
		return nil, ctx.Err()
	case tx.ch <- struct{}{}:
	}

	cur, err := tx.q.QueryContext(ctx, qry, vs...)

	if err != nil {
		<-tx.ch
		return nil, err
	}

	return &cursor{Cursor: cur, tx: tx}, nil
}

func (d *db) Driver() string { return d.driver }

func (d *db) BeginTx(ctx context.Context) (sql.Tx, error) {
	t, err := d.db.BeginTx(ctx, nil)

	if err != nil {
		return nil, err
	}

	return &tx{
		ctx: ctx,
		ch:  make(chan struct{}, 1),
		q:   &queryer{t},
		tx:  t,
	}, nil
}
