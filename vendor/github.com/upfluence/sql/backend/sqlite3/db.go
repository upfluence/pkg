package sqlite3

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/mattn/go-sqlite3"

	"github.com/upfluence/sql"
)

var (
	argRegexp = regexp.MustCompile(`\$\d+`)

	ErrInvalidArgsNumber = errors.New("backend/sqlite3: invalid arg number")
)

type db struct {
	*queryer

	db sql.DB
}

func NewDB(d sql.DB) sql.DB {
	return &db{queryer: &queryer{q: d}, db: d}
}

func (db *db) BeginTx(ctx context.Context) (sql.Tx, error) {
	dtx, err := db.db.BeginTx(ctx)

	if err != nil {
		return nil, err
	}

	return &tx{queryer: &queryer{q: dtx}, tx: dtx}, nil
}

type tx struct {
	*queryer

	tx sql.Tx
}

func (tx *tx) Commit() error   { return tx.tx.Commit() }
func (tx *tx) Rollback() error { return tx.tx.Rollback() }

func (db *db) Driver() string { return db.db.Driver() }

type queryer struct {
	q sql.Queryer
}

func (q *queryer) Exec(ctx context.Context, stmt string, vs ...interface{}) (sql.Result, error) {
	stmt, vs, err := q.rewrite(stmt, vs)

	if err != nil {
		return nil, err
	}

	res, err := q.q.Exec(ctx, stmt, vs...)

	return res, wrapErr(err)
}

func (q *queryer) QueryRow(ctx context.Context, stmt string, vs ...interface{}) sql.Scanner {
	stmt, vs, err := q.rewrite(stmt, vs)

	if err != nil {
		return errScanner{err}
	}

	return &scanner{sc: q.q.QueryRow(ctx, stmt, vs...)}
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

type errScanner struct {
	error
}

func (es errScanner) Scan(...interface{}) error { return es.error }

func (q *queryer) Query(ctx context.Context, stmt string, vs ...interface{}) (sql.Cursor, error) {
	stmt, vs, err := q.rewrite(stmt, vs)

	if err != nil {
		return nil, err
	}

	cur, err := q.q.Query(ctx, stmt, vs...)

	return &cursor{Cursor: cur}, wrapErr(err)
}

func (q *queryer) rewrite(stmt string, vs []interface{}) (string, []interface{}, error) {
	var (
		args = make(map[int]int)

		i int
	)

	vs = sql.StripReturningFields(vs)

	rstmt := argRegexp.ReplaceAllStringFunc(stmt, func(frag string) string {
		v, err := strconv.Atoi(strings.TrimPrefix(frag, "$"))

		if err != nil {
			panic(err)
		}

		args[v] = i
		i++

		return "?"
	})

	if len(vs) != len(args) {
		return "", nil, ErrInvalidArgsNumber
	}

	rvs := make([]interface{}, len(vs))

	for k, i := range args {
		if k > len(rvs) {
			return "", nil, ErrInvalidArgsNumber
		}

		rvs[i] = vs[k-1]
	}

	return rstmt, rvs, nil
}

func wrapErr(err error) error {
	if err == nil {
		return nil
	}

	sqlErr, ok := err.(sqlite3.Error)

	if !ok || sqlErr.Code != sqlite3.ErrConstraint {
		return err
	}

	werr := sql.ConstraintError{Cause: err}

	switch sqlErr.ExtendedCode {
	case sqlite3.ErrConstraintPrimaryKey:
		werr.Type = sql.PrimaryKey
	case sqlite3.ErrConstraintForeignKey:
		werr.Type = sql.ForeignKey
	case sqlite3.ErrConstraintNotNull:
		werr.Type = sql.NotNull
	case sqlite3.ErrConstraintUnique:
		werr.Type = sql.Unique
	}

	return werr
}

func IsSQLite3DB(d sql.DB) bool {
	_, ok := d.(*db)
	return ok
}
