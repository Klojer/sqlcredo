package crud

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Klojer/sqlcredo/internal/goquext"
	"github.com/Klojer/sqlcredo/internal/table"
	"github.com/Klojer/sqlcredo/pkg/api"

	"github.com/doug-martin/goqu/v9"
)

const (
	truncateQueryTemplateSqlite3 = "DELETE FROM %s;"
	truncateQueryTemplateDefault = "TRUNCATE %s;"
)

type CRUD[T any, I comparable] struct {
	table         table.Info
	executor      api.SQLExecutor
	truncateQuery string
	dialect       goqu.DialectWrapper
}

var _ api.CRUD[any, string] = &CRUD[any, string]{}

func NewCRUD[T any, I comparable](table table.Info,
	executor api.SQLExecutor, driver string,
) *CRUD[T, I] {
	return &CRUD[T, I]{
		table:         table,
		executor:      executor,
		truncateQuery: createTruncateQuery(driver, table.Name),
		dialect:       goqu.Dialect(goquext.CreateDialectString(driver)),
	}
}

func (r *CRUD[T, I]) GetAll(ctx context.Context) ([]T, error) {
	query, args, err := r.dialect.From(r.table.Name).Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("unable to create 'select all' query: %w", err)
	}
	return r.selectMany(ctx, query, args...)
}

func (r *CRUD[T, I]) GetByID(ctx context.Context, id I) (T, error) {
	var record T

	query, args, err := r.dialect.From(r.table.Name).
		Where(goqu.I(r.table.IDColumn).Eq(id)).
		Prepared(true).
		ToSQL()
	if err != nil {
		return record, fmt.Errorf("unable to create 'select by id' query: %w", err)
	}

	err = r.executor.SelectOne(ctx, &record, query, args...)
	if err != nil {
		return record, fmt.Errorf("unable to select record: %w", err)
	}

	return record, nil
}

func (r *CRUD[T, I]) GetByIDs(ctx context.Context, ids []I) ([]T, error) {
	query, args, err := r.dialect.From(r.table.Name).
		Where(goqu.I(r.table.IDColumn).In(ids)).
		Order(goqu.I(r.table.IDColumn).Asc()).
		Prepared(true).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("unable to create 'select by ids' query: %w", err)
	}

	entities, err := r.selectMany(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to select records: %w", err)
	}

	return entities, nil
}

func (r *CRUD[T, I]) Create(ctx context.Context, e *T) (sql.Result, error) {
	query, args, err := r.dialect.Insert(r.table.Name).
		Rows(e).
		Prepared(true).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("unable to create 'insert' query: %w", err)
	}

	return r.executor.Exec(ctx, query, args...)
}

func (r *CRUD[T, I]) DeleteAll(ctx context.Context) (sql.Result, error) {
	return r.executor.Exec(ctx, r.truncateQuery)
}

func (r *CRUD[T, I]) Delete(ctx context.Context, id I) (sql.Result, error) {
	query, args, err := r.dialect.Delete(r.table.Name).
		Where(goqu.I(r.table.IDColumn).Eq(id)).
		Prepared(true).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("unable to create 'delete' query: %w", err)
	}

	return r.executor.Exec(ctx, query, args...)
}

func (r *CRUD[T, I]) Update(ctx context.Context, id I, e *T) (sql.Result, error) {
	query, args, err := r.dialect.Update(r.table.Name).
		Set(*e).
		Where(goqu.I(r.table.IDColumn).Eq(id)).
		Prepared(true).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("unable to create 'update' query: %w", err)
	}

	return r.executor.Exec(ctx, query, args...)
}

func (r *CRUD[T, I]) selectMany(ctx context.Context, query string, args ...any) ([]T, error) {
	var records []T
	if err := r.executor.SelectMany(ctx, &records, query, args...); err != nil {
		return nil, fmt.Errorf("unable to load records: %w", err)
	}
	return records, nil
}

func createTruncateQuery(driver string, table string) string {
	if driver == "sqlite3" {
		return fmt.Sprintf(truncateQueryTemplateSqlite3, table)
	}
	return fmt.Sprintf(truncateQueryTemplateDefault, table)
}
