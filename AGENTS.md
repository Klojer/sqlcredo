# AGENTS.md - Development Guidelines for sqlcredo

SQLCredo is a type-safe generic SQL CRUD operations wrapper for Go, built on top of sqlx and goqu. This document provides comprehensive guidelines for contributing to the project.

## Build/Lint/Test Commands

### Core Commands
- **Install dependencies**: `make dependencies` or `go mod tidy`
- **Install golangci-lint**: `make install-golangci-lint`
- **Install git hooks**: `make install-git-hooks`
- **Show version**: `make version`

### Testing Commands
- **Run all tests**: `make test` or `go test -v -race -shuffle=on ./...`
- **Run single test**: `go test -v -run TestName ./path/to/package`
- **Run tests with coverage**: `make test-cov` (opens HTML report in browser)
- **Run tests in specific package**: `go test -v ./internal/crud/...`

### Linting and Quality
- **Run linter**: `make lint` or `golangci-lint run ./...`
- **Format code**: `gofmt -w .` (though golangci-lint usually handles this)
- **Check for race conditions**: Tests automatically run with `-race` flag

### Pre-commit Hooks
The project uses pre-commit hooks that run:
- `go mod tidy` - Ensures dependencies are clean
- `golangci-lint run` - Linting and formatting checks
- `go test ./...` - Runs all tests

## Code Style Guidelines

### Package Structure
- `internal/` - Private implementations (not importable by external packages)
- `examples/` - Usage demonstrations and extended functionality examples
- Root package - Public API interfaces and main constructors

### Imports
- Standard library imports first (alphabetically sorted)
- Third-party packages second (alphabetically sorted)
- Internal packages last (alphabetically sorted)
- Group with blank lines between sections

```go
import (
    "context"
    "database/sql"
    "fmt"

    "github.com/doug-martin/goqu/v9"
    "github.com/stretchr/testify/assert"

    "github.com/Klojer/sqlcredo/internal/domain"
)
```

### Naming Conventions
- **Exported identifiers**: PascalCase (functions, types, constants, variables)
  - `SQLCredo`, `NewSQLCredo()`, `GetAll()`, `DebugFunc`
- **Unexported identifiers**: camelCase (functions, variables)
  - `tableInfo`, `newDebugFunc`, `selectMany()`
- **Acronyms**: Consistent casing - `SQLCredo` (not `SqlCredo`)
- **Type parameters**: Single letters with constraints - `[T any, I comparable]`
- **Interface names**: End with "er" when appropriate - `SQLExecutor`, `PageResolver`

### Types and Generics
- Use generics extensively for type safety: `[T any, I comparable]`
- Define interfaces for contracts and dependency injection
- Use struct tags with `db:"field_name"` for database mapping
- Prefer pointer receivers for methods that modify state
- Use embedded structs for composition over inheritance

```go
type SQLCredo[T any, I comparable] interface {
    SQLExecutor
    CRUD[T, I]
    PageResolver[T]
    TransactionExecutor[T, I]
}

type sqlCredo[T any, I comparable] struct {
    *sqlexec.SQLExecutor
    *crud.CRUD[T, I]
    *page.PageResolver[T]

    tableInfo table.Info
    driver    string
    dbx       *sqlx.DB
}
```

### Constants
- Group related constants in const blocks
- Use meaningful names with context
- Use typed constants when appropriate

```go
const (
    truncateQueryTemplateSqlite3 = "DELETE FROM %s;"
    truncateQueryTemplateDefault = "TRUNCATE %s;"
)
```

### Error Handling
- Wrap errors with context: `fmt.Errorf("unable to execute query: %w", err)`
- Return early on errors to reduce nesting
- Use domain-specific error contexts
- Return both result and error from functions that can fail

```go
func (r *sqlCredo[T, I]) InitSchema(ctx context.Context, sql string) (sql.Result, error) {
    res, err := r.Exec(ctx, sql)
    if err != nil {
        return nil, fmt.Errorf("unable to execute query: %w", err)
    }
    return res, nil
}
```

### Comments
- **Godoc comments** for all exported functions/types (required)
- Include parameter and return descriptions
- Explain purpose and usage, not just what the code does
- Use complete sentences starting with the name being documented

```go
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
```

### Testing
- Use testify library (`assert`, `require`) for assertions
- Table-driven tests for multiple test cases
- Use `t.Cleanup()` for test cleanup
- Mock external dependencies when needed (sqlmock, testcontainers)
- Test both success and error cases
- Use descriptive test names and subtests

```go
func TestSQLCredo_InitSchema(t *testing.T) {
    tests := []struct {
        name    string
        schema  string
        wantErr bool
    }{
        {
            name: "Valid schema creation",
            schema: `CREATE TABLE test_table (
                id TEXT PRIMARY KEY,
                name TEXT NOT NULL
            )`,
            wantErr: false,
        },
        {
            name:    "Invalid SQL syntax",
            schema:  "INVALID SQL QUERY",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c, ctx := newTestCase(t)
            // ... test implementation
        })
    }
}
```

### SQL and Database Patterns
- Use prepared statements by default (enabled via goqu)
- Support multiple SQL dialects (sqlite3, postgres primarily)
- Use goqu for type-safe query building
- Include debug functionality for SQL query inspection
- Handle database-specific syntax differences

### Code Organization
- Clear separation of concerns across packages
- Internal packages should not be imported by external code
- Examples should demonstrate real-world usage patterns
- Keep interfaces focused and minimal
- Use dependency injection rather than global state

### Dependencies
- **Core dependencies**:
  - `github.com/jmoiron/sqlx` - SQL extensions
  - `github.com/doug-martin/goqu/v9` - SQL query builder
  - `github.com/stretchr/testify` - Testing assertions
- **Database drivers**:
  - `github.com/mattn/go-sqlite3` - SQLite support
  - `github.com/jackc/pgx/v5` - PostgreSQL support
- **Testing**:
  - `github.com/DATA-DOG/go-sqlmock` - SQL mocking
  - `github.com/testcontainers/testcontainers-go` - Integration testing

### Performance Considerations
- Prepared statements enabled by default
- Connection pooling handled by underlying database/sql
- Generics provide compile-time type safety without runtime overhead
- Debug functionality can be conditionally enabled

### Contributing Workflow
1. Fork the repository
2. Create a feature branch from `main`
3. Make changes following these guidelines
4. Run tests: `make test`
5. Run linter: `make lint`
6. Ensure pre-commit hooks pass
7. Submit a pull request with a clear description

### Debugging
- Use `WithDebugFunc()` to enable SQL query logging
- Debug function receives the SQL string and arguments
- Useful for development and troubleshooting

```go
repo.WithDebugFunc(func(sql string, args ...any) {
    log.Printf("SQL: %s Args: %v", sql, args)
})
```</content>
<parameter name="filePath">/home/cora/soul/src/sqlcredo/AGENTS.md