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
	countQueryTemplate           = "SELECT COUNT(*) FROM %s;"
	truncateQueryTemplateSqlite3 = "DELETE FROM %s;"
	truncateQueryTemplateDefault = "TRUNCATE %s;"
)

type SQLExecutor interface {
	SelectOne(ctx context.Context, dest any, query string, args ...any) error

	SelectMany(ctx context.Context, dest any, query string, args ...any) error

	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)

	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type pagingParams struct {
	Offset      uint
	Limit       uint
	OrderColumn string
	OrderDesc   bool
}

func newPagingParams(opts ...PagingOpt) pagingParams {
	params := pagingParams{}

	for _, o := range opts {
		o(&params)
	}

	return params
}

type PagingOpt func(*pagingParams)

func WithOffset(offset uint) PagingOpt {
	return func(p *pagingParams) {
		p.Offset = offset
	}
}

func WithLimit(limit uint) PagingOpt {
	return func(p *pagingParams) {
		p.Limit = limit
	}
}

func WithOrderColumn(orderColumn string) PagingOpt {
	return func(p *pagingParams) {
		p.OrderColumn = orderColumn
	}
}

func WithOrderColumnAndDirection(orderColumn string, desc bool) PagingOpt {
	return func(p *pagingParams) {
		p.OrderColumn = orderColumn
		p.OrderDesc = desc
	}
}

type Page[T any] struct {
	Number     uint
	Size       uint
	Total      uint64
	TotalPages uint
	Content    []T
}

type CRUD[T any, I comparable] interface {
	InitSchema(ctx context.Context, sql string) error

	GetAll(ctx context.Context) ([]T, error)

	GetPage(ctx context.Context, opts ...PagingOpt) (Page[T], error)

	GetByID(ctx context.Context, id I) (T, error)

	GetByIDs(ctx context.Context, ids []I) ([]T, error)

	Create(ctx context.Context, e *T) error

	DeleteAll(ctx context.Context) error

	Delete(ctx context.Context, id I) error

	Update(ctx context.Context, id I, e *T) error

	Count(ctx context.Context) (uint64, error)
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
	emptyPage     Page[T]
}

func NewSQLCredo[T any, I comparable](db *sql.DB, driver string, table string, idColumn string) SQLCredo[T, I] {
	return &sqlCredo[T, I]{
		db:            sqlx.NewDb(db, driver),
		table:         table,
		idColumn:      idColumn,
		countQuery:    fmt.Sprintf(countQueryTemplate, table),
		truncateQuery: createTruncateQuery(driver, table),
		debugFunc:     func(sql string, args ...any) {},
		emptyPage:     createEmptyPage[T](),
	}
}

func createTruncateQuery(driver string, table string) string {
	if driver == "sqlite3" {
		return fmt.Sprintf(truncateQueryTemplateSqlite3, table)
	}
	return fmt.Sprintf(truncateQueryTemplateDefault, table)
}

func (r *sqlCredo[T, I]) WithDebugFunc(newDebugFunc DebugFunc) SQLCredo[T, I] {
	r.debugFunc = newDebugFunc
	return r
}

func (r *sqlCredo[T, I]) GetDebugFunc() DebugFunc {
	return r.debugFunc
}

func (r *sqlCredo[T, I]) InitSchema(ctx context.Context, sql string) error {
	_, err := r.db.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	return nil
}

func (r *sqlCredo[T, I]) GetAll(ctx context.Context) ([]T, error) {
	builder := goqu.From(r.table).Prepared(true)
	return r.selectValuesBuilderContext(ctx, builder.ToSQL)
}

func (r *sqlCredo[T, I]) GetPage(ctx context.Context, opts ...PagingOpt) (Page[T], error) {
	params := newPagingParams(opts...)

	pageRecords, err := r.selectMany(ctx, params)
	if err != nil {
		return r.emptyPage, fmt.Errorf("unable to get page items: %w", err)
	}
	totalRecords, err := r.Count(ctx)
	if err != nil {
		return r.emptyPage, fmt.Errorf("unable to count all items: %w", err)
	}

	if len(pageRecords) == 0 {
		return r.emptyPage, nil
	}

	pageNumber := params.Offset / params.Limit

	totalPages := uint(totalRecords) / params.Limit
	if uint(totalRecords)%params.Limit != 0 {
		totalPages += 1
	}

	return Page[T]{
		Number:     pageNumber,
		Size:       uint(len(pageRecords)),
		Total:      totalRecords,
		TotalPages: totalPages,
		Content:    pageRecords,
	}, nil
}

func (r *sqlCredo[T, I]) selectMany(ctx context.Context, paging pagingParams) ([]T, error) {
	return r.selectValuesBuilderContext(ctx, r.selectManyQueryBuilder(paging))
}

func (r *sqlCredo[T, I]) selectManyQueryBuilder(params pagingParams) queryBuilder {
	builder := goqu.From(r.table).Prepared(true)

	builder = builder.Offset(params.Offset)
	builder = builder.Limit(params.Limit)

	if params.OrderDesc {
		builder = builder.Order(goqu.I(params.OrderColumn).Desc())
	} else {
		builder = builder.Order(goqu.I(params.OrderColumn).Asc())
	}

	return builder.ToSQL
}

func (r *sqlCredo[T, I]) GetByID(ctx context.Context, id I) (T, error) {
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

func (r *sqlCredo[T, I]) GetByIDs(ctx context.Context, ids []I) ([]T, error) {
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

func (r *sqlCredo[T, I]) Create(ctx context.Context, e *T) error {
	builder := goqu.Insert(r.table).
		Rows(e).
		Prepared(true)

	return r.execBuilderContext(ctx, builder.ToSQL)
}

func (r *sqlCredo[T, I]) DeleteAll(ctx context.Context) error {
	_, err := r.Exec(ctx, r.truncateQuery)
	return err
}

func (r *sqlCredo[T, I]) Delete(ctx context.Context, id I) error {
	builder := goqu.Delete(r.table).
		Where(goqu.I(r.idColumn).Eq(id)).
		Prepared(true)

	return r.execBuilderContext(ctx, builder.ToSQL)
}

func (r *sqlCredo[T, I]) Update(ctx context.Context, id I, e *T) error {
	builder := goqu.Update(r.table).
		Set(*e).
		Where(goqu.I(r.idColumn).Eq(id)).
		Prepared(true)

	return r.execBuilderContext(ctx, builder.ToSQL)
}

func (r *sqlCredo[T, I]) Count(ctx context.Context) (uint64, error) {
	var res uint64
	if err := r.SelectOne(ctx, &res, r.countQuery); err != nil {
		return 0, fmt.Errorf("failed to count entities: %w", err)
	}

	return res, nil
}

func (r *sqlCredo[T, I]) execBuilderContext(ctx context.Context, builder queryBuilder) error {
	sql, args, err := builder()
	if err != nil {
		return fmt.Errorf("failed to create sql query: %w", err)
	}

	if _, err := r.Exec(ctx, sql, args...); err != nil {
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
	if err = r.SelectMany(ctx, &entities, sql, args...); err != nil {
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
	if err = r.SelectMany(ctx, &entities, sql, args...); err != nil {
		return nil, fmt.Errorf("failed to load enitity values: %w", err)
	}

	return entities, nil
}

func (r *sqlCredo[T, I]) SelectOne(ctx context.Context, dest any, query string, args ...any) error {
	r.debugFunc(query, args...)

	if err := r.db.GetContext(ctx, dest, query, args...); err != nil {
		return fmt.Errorf("failed to get data from db: %w", err)
	}

	return nil
}

func (r *sqlCredo[T, I]) SelectMany(ctx context.Context, dest any, query string, args ...any) error {
	r.debugFunc(query, args...)

	if err := r.db.SelectContext(ctx, dest, query, args...); err != nil {
		return fmt.Errorf("failed to select data from db: %w", err)
	}

	return nil
}

func (r *sqlCredo[T, I]) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	r.debugFunc(query, args...)

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to exec db query: %w", err)
	}

	return res, nil
}

func (r *sqlCredo[T, I]) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}

func createEmptyPage[T any]() Page[T] {
	return Page[T]{
		Number:     0,
		Size:       0,
		Total:      0,
		TotalPages: 0,
		Content:    nil,
	}
}
