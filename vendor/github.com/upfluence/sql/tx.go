package sql

import (
	"context"
	"errors"
)

var ErrRollback = errors.New("sql: rollback sentinel")

type Tx interface {
	Queryer

	Commit() error
	Rollback() error
}

type QueryerFunc func(Queryer) error

func ExecuteTx(ctx context.Context, db DB, fn QueryerFunc) error {
	tx, err := db.BeginTx(ctx)

	if err != nil {
		return err
	}

	switch err := fn(tx); err {
	case nil:
		return tx.Commit()
	case ErrRollback:
		tx.Rollback()
		return nil
	default:
		tx.Rollback()
		return err
	}
}
