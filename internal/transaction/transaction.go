package transaction

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Klojer/sqlcredo/internal/crud"
	"github.com/Klojer/sqlcredo/internal/page"
	"github.com/Klojer/sqlcredo/internal/sqlexec"
	"github.com/Klojer/sqlcredo/internal/table"
	"github.com/Klojer/sqlcredo/pkg/api"

	"github.com/jmoiron/sqlx"
)

type Wrapper[T any, I comparable] struct {
	tx        *sqlx.Tx
	debugFunc api.DebugFunc

	api.SQLExecutor
	api.CRUD[T, I]
	api.PageResolver[T]
}

var _ api.Transaction[any, string] = &Wrapper[any, string]{}

func NewTx[T any, I comparable](ctx context.Context,
	db *sqlx.DB, tableInfo table.Info, driver string, debugFunc api.DebugFunc, opts *sql.TxOptions,
) (*Wrapper[T, I], error) {
	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to begin transaction: %w", err)
	}

	debugFunc("begin transaction")

	exec := sqlexec.NewSQLExecutor(tx)
	exec.DebugFunc = debugFunc

	return &Wrapper[T, I]{
		tx:           tx,
		debugFunc:    debugFunc,
		SQLExecutor:  exec,
		CRUD:         crud.NewCRUD[T, I](tableInfo, exec, driver),
		PageResolver: page.NewPageResolver[T](tableInfo, exec, driver),
	}, nil
}

func (t *Wrapper[T, I]) Commit() error {
	t.debugFunc("commit transaction")
	return t.tx.Commit()
}

func (t *Wrapper[T, I]) Rollback() error {
	t.debugFunc("rollback transaction")
	return t.tx.Rollback()
}
