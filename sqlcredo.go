// Package sqlcredo provides a type-safe generic SQL CRUD operations wrapper for Go.
// It offers built-in CRUD operations, pagination support, transaction management,
// and SQL query debugging capabilities. The package simplifies database interactions
// by providing generic implementations for common database operations while allowing
// custom raw SQL queries for extended functionality. Built on top of sqlx and goqu.
package sqlcredo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Klojer/sqlcredo/internal/crud"
	"github.com/Klojer/sqlcredo/internal/page"
	"github.com/Klojer/sqlcredo/internal/sqlexec"
	"github.com/Klojer/sqlcredo/internal/table"
	"github.com/Klojer/sqlcredo/internal/transaction"

	"github.com/jmoiron/sqlx"
)

// SQLCredo is a comprehensive interface that combines SQL execution, CRUD operations,
// and pagination capabilities for a specific entity type.
//
// Type Parameters:
//   - T: The entity type being managed (can be any type)
//   - I: The type of the entity's ID field (must be comparable)
type SQLCredo[T any, I comparable] interface {
	SQLExecutor
	CRUD[T, I]
	PageResolver[T]
	TransactionExecutor[T, I]

	// InitSchema executes a SQL query to initialize the database schema.
	// Typically used for creating tables and other database objects.
	InitSchema(ctx context.Context, sql string) (sql.Result, error)

	// WithDebugFunc sets a debug function for SQL query logging.
	// The debug function will be called before executing any SQL query.
	// Returns the modified SQLCredo instance for method chaining.
	WithDebugFunc(newDebugFunc DebugFunc) SQLCredo[T, I]

	// GetDebugFunc returns the currently set debug function.
	// Returns nil if no debug function is set.
	GetDebugFunc() DebugFunc
}

type sqlCredo[T any, I comparable] struct {
	*sqlexec.SQLExecutor
	*crud.CRUD[T, I]
	*page.PageResolver[T]

	tableInfo table.Info
	driver    string
	dbx       *sqlx.DB
}

var _ SQLCredo[any, string] = &sqlCredo[any, string]{}

// NewSQLCredo creates a new instance of SQLCredo for the specified entity type and ID type.
//
// Parameters:
//   - db: A pointer to the underlying database connection
//   - driver: The database driver name (e.g., "postgres", "mysql")
//   - tableName: The name of the database table for the entity
//   - idColumn: The name of the ID column in the table
//
// Returns a fully initialized SQLCredo instance
func NewSQLCredo[T any, I comparable](db *sql.DB, driver string, tableName string, idColumn string) SQLCredo[T, I] {
	tableInfo := table.Info{Name: tableName, IDColumn: idColumn}
	dbx := sqlx.NewDb(db, driver)
	executor := sqlexec.NewSQLExecutor(dbx)

	return &sqlCredo[T, I]{
		SQLExecutor:  executor,
		CRUD:         crud.NewCRUD[T, I](tableInfo, executor, driver),
		PageResolver: page.NewPageResolver[T](tableInfo, executor, driver),
		tableInfo:    tableInfo,
		driver:       driver,
		dbx:          dbx,
	}
}

func (r *sqlCredo[T, I]) BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction[T, I], error) {
	return transaction.NewTx[T, I](ctx, r.dbx, r.tableInfo, r.driver, r.DebugFunc, opts)
}

// InitSchema executes a SQL query to initialize the database schema.
// This method is typically used during application startup to ensure
// the required database structure exists.
func (r *sqlCredo[T, I]) InitSchema(ctx context.Context, sql string) (sql.Result, error) {
	res, err := r.Exec(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}
	return res, nil
}

// WithDebugFunc sets a new debug function for SQL query logging.
// The debug function will be called before executing any SQL query,
// allowing for query inspection and logging.
func (r *sqlCredo[T, I]) WithDebugFunc(newDebugFunc DebugFunc) SQLCredo[T, I] {
	r.DebugFunc = newDebugFunc
	return r
}

// GetDebugFunc returns the currently set debug function.
// Returns nil if no debug function has been set.
func (r *sqlCredo[T, I]) GetDebugFunc() DebugFunc {
	return r.DebugFunc
}
