package domain

import (
	"context"
	"database/sql"
)

// Transaction is a comprehensive interface that combines major SQLCredo capabilities,
// but wrap them in transaction.
type Transaction[T any, I comparable] interface {
	SQLExecutor
	CRUD[T, I]
	PageResolver[T]

	// Commit commits the transaction
	Commit() error

	// Rollback aborts the transaction
	Rollback() error
}

// TransactionExecutor deifnes an interface to execute transactions.
type TransactionExecutor[T any, I comparable] interface {
	// BeginTx creates and starts new transaction.
	// Returns the transaction object and any error encountered during transaction creation.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction[T, I], error)
}
