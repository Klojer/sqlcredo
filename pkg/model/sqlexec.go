package model

import (
	"context"
	"database/sql"
)

type SQLExecutor interface {
	SelectOne(ctx context.Context, dest any, query string, args ...any) error
	SelectMany(ctx context.Context, dest any, query string, args ...any) error
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
