package model

import (
	"context"
	"database/sql"
)

// SQLExecutor defines an interface for executing SQL operations.
// It provides a high-level abstraction for common database operations
// like selecting single/multiple rows, executing statements, and managing transactions.
type SQLExecutor interface {
	// SelectOne executes a query that is expected to return at most one row.
	// It scans the resulting row into the dest parameter, which must be a pointer.
	// Returns an error if the query fails or if scanning into dest fails.
	SelectOne(ctx context.Context, dest any, query string, args ...any) error

	// SelectMany executes a query that can return multiple rows.
	// It scans all resulting rows into the dest parameter, which must be a pointer to a slice.
	// Returns an error if the query fails or if scanning into dest fails.
	SelectMany(ctx context.Context, dest any, query string, args ...any) error

	// Exec executes a query that doesn't return rows (like INSERT, UPDATE, DELETE).
	// Returns a Result object summarizing the effect of the query and any error encountered.
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)

	// BeginTx starts a new transaction with the given options.
	// Returns the transaction object and any error encountered during transaction creation.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
