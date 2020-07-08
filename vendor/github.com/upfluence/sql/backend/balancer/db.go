package balancer

import (
	"context"

	"github.com/upfluence/sql"
)

func NewDB(bb BalancerBuilder, dbs ...sql.DB) sql.DB {
	switch len(dbs) {
	case 0:
		return nil
	case 1:
		return dbs[0]
	}

	return &db{b: bb.Build(dbs), driver: dbs[0].Driver()}
}

type db struct {
	b Balancer

	driver string
}

func (d *db) Driver() string { return d.driver }

func (d *db) BeginTx(ctx context.Context) (sql.Tx, error) {
	db, cfn, err := d.b.Get(ctx)

	if err != nil {
		return nil, err
	}

	subTx, err := db.BeginTx(ctx)

	if err != nil {
		return nil, err
	}

	return &tx{Tx: subTx, cfn: cfn}, nil
}

func (d *db) Exec(ctx context.Context, q string, vs ...interface{}) (sql.Result, error) {
	db, cfn, err := d.b.Get(ctx)

	if err != nil {
		return nil, err
	}

	res, err := db.Exec(ctx, q, vs...)

	cfn(err)

	return res, err
}

type errScanner struct {
	err error
}

func (esc errScanner) Scan(...interface{}) error { return esc.err }

func (d *db) QueryRow(ctx context.Context, q string, vs ...interface{}) sql.Scanner {
	db, cfn, err := d.b.Get(ctx)

	if err != nil {
		return errScanner{err}
	}

	return &scanner{sc: db.QueryRow(ctx, q, vs...), cfn: cfn}
}

func (d *db) Query(ctx context.Context, q string, vs ...interface{}) (sql.Cursor, error) {
	db, cfn, err := d.b.Get(ctx)

	if err != nil {
		return nil, err
	}

	cur, err := db.Query(ctx, q, vs...)

	if err != nil {
		return nil, err
	}

	return &cursor{Cursor: cur, cfn: cfn}, nil
}

type scanner struct {
	sc  sql.Scanner
	cfn CloseFunc
}

func (sc *scanner) Scan(vs ...interface{}) error {
	err := sc.sc.Scan(vs...)

	sc.cfn(err)

	return err
}

type cursor struct {
	sql.Cursor

	cfn CloseFunc
}

func (c *cursor) Close() error {
	err := c.Cursor.Close()

	c.cfn(err)

	return err
}

type tx struct {
	sql.Tx

	cfn CloseFunc
}

func (tx *tx) Commit() error {
	err := tx.Tx.Commit()

	tx.cfn(err)

	return err
}

func (tx *tx) Rollback() error {
	err := tx.Tx.Rollback()

	tx.cfn(err)

	return err
}
