// Package sqlexec provides the SQL execution implementation for sqlcredo.
// It wraps sqlx database operations with debugging support, implementing
// the domain.SQLExecutor interface for executing queries and statements.
package sqlexec

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Klojer/sqlcredo/internal/domain"
)

type SQLXExecutor interface {
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type SQLExecutor struct {
	db        SQLXExecutor
	DebugFunc domain.DebugFunc
}

var _ domain.SQLExecutor = &SQLExecutor{}

func NewSQLExecutor(db SQLXExecutor) *SQLExecutor {
	return &SQLExecutor{
		db:        db,
		DebugFunc: func(sql string, args ...any) {},
	}
}

// delegatingExecutor wraps a domain.SQLExecutor and delegates all calls to it
type delegatingExecutor struct {
	executor domain.SQLExecutor
}

func (d *delegatingExecutor) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return d.executor.SelectOne(ctx, dest, query, args...)
}

func (d *delegatingExecutor) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return d.executor.SelectMany(ctx, dest, query, args...)
}

func (d *delegatingExecutor) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.executor.Exec(ctx, query, args...)
}

// NewDelegatingExecutor creates a new SQLExecutor that delegates to the provided domain.SQLExecutor.
// This is useful when you want to wrap a transaction or other executor in the SQLExecutor type.
func NewDelegatingExecutor(executor domain.SQLExecutor, debugFunc domain.DebugFunc) *SQLExecutor {
	return &SQLExecutor{
		db:        &delegatingExecutor{executor: executor},
		DebugFunc: debugFunc,
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
