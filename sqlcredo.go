package sqlcredo

import (
	"context"
	"database/sql"
	"fmt"

	"gitlab.com/onrooh/sqlcredo/internal/crud"
	"gitlab.com/onrooh/sqlcredo/internal/page"
	"gitlab.com/onrooh/sqlcredo/internal/sqlexec"
	"gitlab.com/onrooh/sqlcredo/pkg/model"

	"github.com/jmoiron/sqlx"
)

type SQLCredo[T any, I comparable] interface {
	model.SQLExecutor
	model.CRUD[T, I]
	model.PageResolver[T]

	InitSchema(ctx context.Context, sql string) (sql.Result, error)

	WithDebugFunc(newDebugFunc model.DebugFunc) SQLCredo[T, I]
	GetDebugFunc() model.DebugFunc
}

type sqlCredo[T any, I comparable] struct {
	*sqlexec.SQLExecutor
	*crud.CRUD[T, I]
	*page.PageResolver[T]
}

var _ SQLCredo[any, string] = &sqlCredo[any, string]{}

func NewSQLCredo[T any, I comparable](db *sql.DB, driver string, table string, idColumn string) SQLCredo[T, I] {
	tableInfo := model.TableInfo{Name: table, IDColumn: idColumn}
	dbx := sqlx.NewDb(db, driver)
	executor := sqlexec.NewSQLExecutor(dbx)

	return &sqlCredo[T, I]{
		SQLExecutor:  executor,
		CRUD:         crud.NewCRUD[T, I](tableInfo, executor, driver),
		PageResolver: page.NewPageResolver[T](tableInfo, executor),
	}
}

func (r *sqlCredo[T, I]) InitSchema(ctx context.Context, sql string) (sql.Result, error) {
	res, err := r.Exec(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}
	return res, nil
}

func (r *sqlCredo[T, I]) WithDebugFunc(newDebugFunc model.DebugFunc) SQLCredo[T, I] {
	r.DebugFunc = newDebugFunc
	return r
}

func (r *sqlCredo[T, I]) GetDebugFunc() model.DebugFunc {
	return r.DebugFunc
}
