package sql

import (
	"context"
	"database/sql"
)

type (
	Result     = sql.Result
	NullInt64  = sql.NullInt64
	NullString = sql.NullString
	NullBool   = sql.NullBool
)

var (
	ErrConnDone = sql.ErrConnDone
	ErrNoRows   = sql.ErrNoRows
	ErrTxDone   = sql.ErrTxDone
)

type Scanner interface {
	Scan(...interface{}) error
}

type Queryer interface {
	Exec(context.Context, string, ...interface{}) (Result, error)
	QueryRow(context.Context, string, ...interface{}) Scanner
	Query(context.Context, string, ...interface{}) (Cursor, error)
}

type DB interface {
	Queryer

	BeginTx(context.Context) (Tx, error)
	Driver() string
}

type Returning struct {
	Field string
}

func StripReturningFields(vs []interface{}) []interface{} {
	var res []interface{}

	for _, v := range vs {
		if _, ok := v.(*Returning); !ok {
			res = append(res, v)
		}
	}

	return res
}

type MiddlewareFactory interface {
	Wrap(DB) DB
}
