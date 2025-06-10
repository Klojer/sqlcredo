package sqlcredo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Klojer/sqlcredo/internal/crud"
	"github.com/Klojer/sqlcredo/internal/page"
	"github.com/Klojer/sqlcredo/internal/sqlexec"
	"github.com/Klojer/sqlcredo/internal/table"
	"github.com/Klojer/sqlcredo/pkg/model"

	"github.com/jmoiron/sqlx"
)

// SQLCredo is a comprehensive interface that combines SQL execution, CRUD operations,
// and pagination capabilities for a specific entity type.
//
// Type Parameters:
//   - T: The entity type being managed (can be any type)
//   - I: The type of the entity's ID field (must be comparable)
type SQLCredo[T any, I comparable] interface {
	model.SQLExecutor
	model.CRUD[T, I]
	model.PageResolver[T]

	// InitSchema executes a SQL query to initialize the database schema.
	// Typically used for creating tables and other database objects.
	InitSchema(ctx context.Context, sql string) (sql.Result, error)

	// WithDebugFunc sets a debug function for SQL query logging.
	// The debug function will be called before executing any SQL query.
	// Returns the modified SQLCredo instance for method chaining.
	WithDebugFunc(newDebugFunc model.DebugFunc) SQLCredo[T, I]

	// GetDebugFunc returns the currently set debug function.
	// Returns nil if no debug function is set.
	GetDebugFunc() model.DebugFunc
}

type sqlCredo[T any, I comparable] struct {
	*sqlexec.SQLExecutor
	*crud.CRUD[T, I]
	*page.PageResolver[T]
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
	}
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
func (r *sqlCredo[T, I]) WithDebugFunc(newDebugFunc model.DebugFunc) SQLCredo[T, I] {
	r.DebugFunc = newDebugFunc
	return r
}

// GetDebugFunc returns the currently set debug function.
// Returns nil if no debug function has been set.
func (r *sqlCredo[T, I]) GetDebugFunc() model.DebugFunc {
	return r.DebugFunc
}
