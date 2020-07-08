package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/upfluence/sql"
	"github.com/upfluence/sql/sqlparser"
)

type fakeResult int64

func (r fakeResult) LastInsertId() (int64, error) { return int64(r), nil }
func (fakeResult) RowsAffected() (int64, error)   { return 1, nil }

type db struct {
	*queryer

	db sql.DB
}

func NewDB(d sql.DB, p sqlparser.SQLParser) sql.DB {
	return &db{queryer: &queryer{q: d, p: p}, db: d}
}

func (db *db) Driver() string { return db.db.Driver() }

func (db *db) BeginTx(ctx context.Context) (sql.Tx, error) {
	cur, err := db.db.BeginTx(ctx)

	if err != nil {
		return nil, err
	}

	return &tx{queryer: &queryer{q: cur, p: db.p}, tx: cur}, nil
}

type tx struct {
	*queryer

	tx sql.Tx
}

func (tx *tx) Commit() error   { return tx.tx.Commit() }
func (tx *tx) Rollback() error { return tx.tx.Rollback() }

type queryer struct {
	q sql.Queryer
	p sqlparser.SQLParser
}

func (q *queryer) QueryRow(ctx context.Context, stmt string, vs ...interface{}) sql.Scanner {
	return &scanner{sc: q.q.QueryRow(ctx, stmt, vs...)}
}

func (q *queryer) Query(ctx context.Context, stmt string, vs ...interface{}) (sql.Cursor, error) {
	cur, err := q.q.Query(ctx, stmt, vs...)

	if err != nil {
		return nil, wrapErr(err)
	}

	return &cursor{Cursor: cur}, nil
}

func (q *queryer) Exec(ctx context.Context, stmt string, vs ...interface{}) (sql.Result, error) {
	if q.p.GetStatementType(stmt) != sqlparser.StmtInsert {
		res, err := q.q.Exec(ctx, stmt, vs...)
		return res, wrapErr(err)
	}

	var (
		args []interface{}
		ret  *sql.Returning
	)

	for _, v := range vs {
		if r, ok := v.(*sql.Returning); ok {
			ret = r
		} else {
			args = append(args, v)
		}
	}

	if ret != nil {
		var id int64

		if err := q.q.QueryRow(
			ctx,
			fmt.Sprintf("%s RETURNING %s", stmt, ret.Field),
			args...,
		).Scan(&id); err != nil {
			return nil, wrapErr(err)
		}

		return fakeResult(id), nil
	}

	res, err := q.q.Exec(ctx, stmt, vs...)
	return res, wrapErr(err)
}

const constraintClass = pq.ErrorClass("23")

func wrapErr(err error) error {
	if err == nil {
		return err
	}

	pqErr, ok := err.(*pq.Error)

	if !ok || pqErr.Code.Class() != constraintClass {
		return err
	}

	werr := sql.ConstraintError{Cause: err}

	switch pqErr.Code {
	case pq.ErrorCode("23503"):
		werr.Type = sql.ForeignKey
	case pq.ErrorCode("23502"):
		werr.Type = sql.NotNull
	case pq.ErrorCode("23505"):
		if strings.HasSuffix(pqErr.Constraint, "_pkey") {
			werr.Type = sql.PrimaryKey
		} else {
			werr.Type = sql.Unique
		}
	}

	return werr
}

type scanner struct {
	sc sql.Scanner
}

func (sc *scanner) Scan(vs ...interface{}) error {
	return wrapErr(sc.sc.Scan(vs...))
}

type cursor struct {
	sql.Cursor
}

func (c *cursor) Scan(vs ...interface{}) error {
	return wrapErr(c.Cursor.Scan(vs...))
}

func IsPostgresDB(d sql.DB) bool {
	_, ok := d.(*db)
	return ok
}
