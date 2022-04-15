package pgsvc

import (
	"context"
	"database/sql"
	"fmt"

	"optrispace.com/work/pkg/db/pgdao"
)

var (
	defaultRwTxOpts = &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: false}
	defaultRoTxOpts = &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true}
)

// doWithQueries runs f with specified function supplied with appropriate Queries
func doWithQueries(ctx context.Context, db *sql.DB, opts *sql.TxOptions, f func(queries *pgdao.Queries) error) error {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("unable to start transaction: %w", err)
	}
	defer tx.Rollback() // nolint: errcheck

	queries := pgdao.New(nil).WithTx(tx)

	if e := f(queries); e != nil {
		return e // function should supply with additional details
	}
	if e := tx.Commit(); e != nil {
		return fmt.Errorf("unable to commit transaction: %w", e)
	}
	return nil
}
