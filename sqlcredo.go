package sqlcredo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
)

var ErrRecordNotFound = errors.New("record not found")

const (
	_countQueryTemplate           = "SELECT COUNT(*) FROM %s;"
	_truncateQueryTemplateSqlite3 = "DELETE FROM %s;"
	_truncateQueryTemplateDefault = "TRUNCATE %s;"
)

type SQLExecutor interface {
	SelectOne(dest any, query string, args ...any) error
	SelectOneContext(ctx context.Context, dest any, query string, args ...any) error

	SelectMany(dest any, query string, args ...any) error
	SelectManyContext(ctx context.Context, dest any, query string, args ...any) error

	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)

	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type pagingParams struct {
	Offset      *uint
	Limit       *uint
	OrderColumn *string
	OrderDesc   bool
}

type PagingOpt func(*pagingParams)

func WithOffset(offset uint) PagingOpt {
	return func(p *pagingParams) {
		p.Offset = &offset
	}
}

func WithLimit(limit uint) PagingOpt {
	return func(p *pagingParams) {
		p.Limit = &limit
	}
}

func WithOrderColumn(orderColumn string) PagingOpt {
	return func(p *pagingParams) {
		p.OrderColumn = &orderColumn
	}
}

func WithOrderColumnAndDirection(orderColumn string, desc bool) PagingOpt {
	return func(p *pagingParams) {
		p.OrderColumn = &orderColumn
		p.OrderDesc = desc
	}
}

type CRUD[T any, I comparable] interface {
	InitSchema(sql string) error
	InitSchemaContext(ctx context.Context, sql string) error

	GetAll(opts ...PagingOpt) ([]*T, error)
	GetAllContext(ctx context.Context, opts ...PagingOpt) ([]*T, error)
	GetAllValues(opts ...PagingOpt) ([]T, error)
	GetAllValuesContext(ctx context.Context, opts ...PagingOpt) ([]T, error)

	GetByID(id I) (*T, error)
	GetByIDContext(ctx context.Context, id I) (*T, error)
	GetValueByID(id I) (T, error)
	GetValueByIDContext(ctx context.Context, id I) (T, error)

	GetByIDs(ids []I) ([]*T, error)
	GetByIDsContext(ctx context.Context, ids []I) ([]*T, error)
	GetValuesByIDs(ids []I) ([]T, error)
	GetValuesByIDsContext(ctx context.Context, ids []I) ([]T, error)

	Create(e *T) error
	CreateContext(ctx context.Context, e *T) error

	DeleteAll() error
	DeleteAllContext(ctx context.Context) error

	Delete(id I) error
	DeleteContext(ctx context.Context, id I) error

	Update(id I, e *T) error
	UpdateContext(ctx context.Context, id I, e *T) error

	Count() (int64, error)
	CountContext(ctx context.Context) (int64, error)
}

type DebugFunc func(sql string, args ...any)

type SQLCredo[T any, I comparable] interface {
	SQLExecutor
	CRUD[T, I]

	WithDebugFunc(newDebugFunc DebugFunc) SQLCredo[T, I]
	GetDebugFunc() DebugFunc
}

type queryBuilder func() (sql string, params []any, err error)

type sqlCredo[T any, I comparable] struct {
	db            *sqlx.DB
	table         string
	idColumn      string
	countQuery    string
	truncateQuery string
	debugFunc     DebugFunc
}

func NewSQLCredo[T any, I comparable](db *sql.DB, driver string, table string, idColumn string) SQLCredo[T, I] {
	return &sqlCredo[T, I]{
		db:            sqlx.NewDb(db, driver),
		table:         table,
		idColumn:      idColumn,
		countQuery:    fmt.Sprintf(_countQueryTemplate, table),
		truncateQuery: createTruncateQuery(driver, table),
		debugFunc:     func(sql string, args ...any) {},
	}
}

func createTruncateQuery(driver string, table string) string {
	if driver == "sqlite3" {
		return fmt.Sprintf(_truncateQueryTemplateSqlite3, table)
	}
	return fmt.Sprintf(_truncateQueryTemplateDefault, table)
}

func (r *sqlCredo[T, I]) WithDebugFunc(newDebugFunc DebugFunc) SQLCredo[T, I] {
	r.debugFunc = newDebugFunc
	return r
}

func (r *sqlCredo[T, I]) GetDebugFunc() DebugFunc {
	return r.debugFunc
}

func (r *sqlCredo[T, I]) InitSchema(sql string) error {
	return r.InitSchemaContext(context.Background(), sql)
}

func (r *sqlCredo[T, I]) InitSchemaContext(ctx context.Context, sql string) error {
	_, err := r.db.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	return nil
}

func (r *sqlCredo[T, I]) GetAll(opts ...PagingOpt) ([]*T, error) {
	return r.GetAllContext(context.Background(), opts...)
}

func (r *sqlCredo[T, I]) GetAllContext(ctx context.Context, opts ...PagingOpt) ([]*T, error) {
	return r.selectBuilderContext(ctx, r.getAllQueryBuilder(ctx, opts...))
}

func (r *sqlCredo[T, I]) GetAllValues(opts ...PagingOpt) ([]T, error) {
	return r.GetAllValuesContext(context.Background(), opts...)
}

func (r *sqlCredo[T, I]) GetAllValuesContext(ctx context.Context, opts ...PagingOpt) ([]T, error) {
	return r.selectValuesBuilderContext(ctx, r.getAllQueryBuilder(ctx, opts...))
}

func (r *sqlCredo[T, I]) getAllQueryBuilder(ctx context.Context, opts ...PagingOpt) queryBuilder {
	params := &pagingParams{}

	for _, o := range opts {
		o(params)
	}

	builder := goqu.From(r.table).Prepared(true)

	if params.Offset != nil {
		builder = builder.Offset(*params.Offset)
	}

	if params.Limit != nil {
		builder = builder.Limit(*params.Limit)
	}

	if params.OrderColumn != nil {
		if params.OrderDesc {
			builder = builder.Order(goqu.I(*params.OrderColumn).Desc())
		} else {
			builder = builder.Order(goqu.I(*params.OrderColumn).Asc())
		}
	}

	return builder.ToSQL
}

func (r *sqlCredo[T, I]) GetByID(id I) (*T, error) {
	return r.GetByIDContext(context.Background(), id)
}

func (r *sqlCredo[T, I]) GetByIDContext(ctx context.Context, id I) (*T, error) {
	builder := goqu.From(r.table).
		Where(goqu.I(r.idColumn).Eq(id)).
		Prepared(true)

	entities, err := r.selectBuilderContext(context.Background(), builder.ToSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to select entities: %w", err)
	}

	// TODO: check there is exact one entity in result
	if len(entities) < 1 {
		return nil, errors.New("record not found")
	}

	return entities[0], nil
}

func (r *sqlCredo[T, I]) GetValueByID(id I) (T, error) {
	return r.GetValueByIDContext(context.Background(), id)
}

func (r *sqlCredo[T, I]) GetValueByIDContext(ctx context.Context, id I) (T, error) {
	builder := goqu.From(r.table).
		Where(goqu.I(r.idColumn).Eq(id)).
		Prepared(true)

	var empty T

	entities, err := r.selectValuesBuilderContext(context.Background(), builder.ToSQL)
	if err != nil {
		return empty, fmt.Errorf("failed to select entities values: %w", err)
	}

	// TODO: check there is exact one entity in result
	if len(entities) < 1 {
		return empty, ErrRecordNotFound
	}

	return entities[0], nil
}

func (r *sqlCredo[T, I]) GetByIDs(ids []I) ([]*T, error) {
	return r.GetByIDsContext(context.Background(), ids)
}

func (r *sqlCredo[T, I]) GetByIDsContext(ctx context.Context, ids []I) ([]*T, error) {
	builder := goqu.From(r.table).
		Where(goqu.I(r.idColumn).In(ids)).
		Order(goqu.I(r.idColumn).Asc()).
		Prepared(true)

	entities, err := r.selectBuilderContext(ctx, builder.ToSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to select entities: %w", err)
	}

	return entities, nil
}

func (r *sqlCredo[T, I]) GetValuesByIDs(ids []I) ([]T, error) {
	return r.GetValuesByIDsContext(context.Background(), ids)
}

func (r *sqlCredo[T, I]) GetValuesByIDsContext(ctx context.Context, ids []I) ([]T, error) {
	builder := goqu.From(r.table).
		Where(goqu.I(r.idColumn).In(ids)).
		Order(goqu.I(r.idColumn).Asc()).
		Prepared(true)

	entities, err := r.selectValuesBuilderContext(ctx, builder.ToSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to select values entities: %w", err)
	}

	return entities, nil
}

func (r *sqlCredo[T, I]) Create(e *T) error {
	return r.CreateContext(context.Background(), e)
}

func (r *sqlCredo[T, I]) CreateContext(ctx context.Context, e *T) error {
	builder := goqu.Insert(r.table).
		Rows(e).
		Prepared(true)

	return r.execBuilderContext(ctx, builder.ToSQL)
}

func (r *sqlCredo[T, I]) DeleteAll() error {
	return r.DeleteAllContext(context.Background())
}

func (r *sqlCredo[T, I]) DeleteAllContext(ctx context.Context) error {
	_, err := r.ExecContext(ctx, r.truncateQuery)
	return err
}

func (r *sqlCredo[T, I]) Delete(id I) error {
	return r.DeleteContext(context.Background(), id)
}

func (r *sqlCredo[T, I]) DeleteContext(ctx context.Context, id I) error {
	builder := goqu.Delete(r.table).
		Where(goqu.I(r.idColumn).Eq(id)).
		Prepared(true)

	return r.execBuilderContext(ctx, builder.ToSQL)
}

func (r *sqlCredo[T, I]) Update(id I, e *T) error {
	return r.UpdateContext(context.Background(), id, e)
}

func (r *sqlCredo[T, I]) UpdateContext(ctx context.Context, id I, e *T) error {
	builder := goqu.Update(r.table).
		Set(*e).
		Where(goqu.I(r.idColumn).Eq(id)).
		Prepared(true)

	return r.execBuilderContext(ctx, builder.ToSQL)
}

func (r *sqlCredo[T, I]) Count() (int64, error) {
	return r.CountContext(context.Background())
}

func (r *sqlCredo[T, I]) CountContext(ctx context.Context) (int64, error) {
	var res int64
	if err := r.SelectOneContext(ctx, &res, r.countQuery); err != nil {
		return 0, fmt.Errorf("failed to count entities: %w", err)
	}

	return res, nil
}

func (r *sqlCredo[T, I]) execBuilderContext(ctx context.Context, builder queryBuilder) error {
	sql, args, err := builder()
	if err != nil {
		return fmt.Errorf("failed to create sql query: %w", err)
	}

	if _, err := r.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("failed to execute sql query: %w", err)
	}

	return nil
}

func (r *sqlCredo[T, I]) selectBuilderContext(ctx context.Context, builder queryBuilder) ([]*T, error) {
	sql, args, err := builder()
	if err != nil {
		return nil, fmt.Errorf("failed to create select query: %w", err)
	}

	var entities []*T
	if err = r.SelectManyContext(ctx, &entities, sql, args...); err != nil {
		return nil, fmt.Errorf("failed to load enitites: %w", err)
	}

	return entities, nil
}

func (r *sqlCredo[T, I]) selectValuesBuilderContext(ctx context.Context, builder queryBuilder) ([]T, error) {
	sql, args, err := builder()
	if err != nil {
		return nil, fmt.Errorf("failed to create select values query: %w", err)
	}

	var entities []T
	if err = r.SelectManyContext(ctx, &entities, sql, args...); err != nil {
		return nil, fmt.Errorf("failed to load enitity values: %w", err)
	}

	return entities, nil
}

func (r *sqlCredo[T, I]) SelectOne(dest any, query string, args ...any) error {
	return r.SelectOneContext(context.Background(), dest, query, args...)
}

func (r *sqlCredo[T, I]) SelectOneContext(ctx context.Context, dest any, query string, args ...any) error {
	r.debugFunc(query, args...)

	if err := r.db.GetContext(ctx, dest, query, args...); err != nil {
		return fmt.Errorf("failed to get data from db: %w", err)
	}

	return nil
}

func (r *sqlCredo[T, I]) SelectMany(dest any, query string, args ...any) error {
	return r.SelectManyContext(context.Background(), dest, query, args...)
}

func (r *sqlCredo[T, I]) SelectManyContext(ctx context.Context, dest any, query string, args ...any) error {
	r.debugFunc(query, args...)

	if err := r.db.SelectContext(ctx, dest, query, args...); err != nil {
		return fmt.Errorf("failed to select data from db: %w", err)
	}

	return nil
}

func (r *sqlCredo[T, I]) Exec(query string, args ...any) (sql.Result, error) {
	return r.ExecContext(context.Background(), query, args...)
}

func (r *sqlCredo[T, I]) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	r.debugFunc(query, args...)

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to exec db query: %w", err)
	}

	return res, nil
}

func (r *sqlCredo[T, I]) Begin() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *sqlCredo[T, I]) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}
