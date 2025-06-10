package sqlexec

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Klojer/sqlcredo/pkg/model"

	"github.com/jmoiron/sqlx"
)

type SQLExecutor struct {
	db        *sqlx.DB
	DebugFunc model.DebugFunc
}

var _ model.SQLExecutor = &SQLExecutor{}

func NewSQLExecutor(db *sqlx.DB) *SQLExecutor {
	return &SQLExecutor{
		db:        db,
		DebugFunc: func(sql string, args ...any) {},
	}
}

func (r *SQLExecutor) SelectOne(ctx context.Context, dest any, query string, args ...any) error {
	r.DebugFunc(query, args...)

	if err := r.db.GetContext(ctx, dest, query, args...); err != nil {
		return fmt.Errorf("unable to get data from db: %w", err)
	}

	return nil
}

func (r *SQLExecutor) SelectMany(ctx context.Context, dest any, query string, args ...any) error {
	r.DebugFunc(query, args...)

	if err := r.db.SelectContext(ctx, dest, query, args...); err != nil {
		return fmt.Errorf("unable to select data from db: %w", err)
	}

	return nil
}

func (r *SQLExecutor) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	r.DebugFunc(query, args...)

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to exec db query: %w", err)
	}

	return res, nil
}

func (r *SQLExecutor) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}
