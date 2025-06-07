package sqlcredo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
)

var (
	ErrRecordNotFound            = errors.New("record not found")
	ErrUnexpectedNumberOfRecords = errors.New("unexpected number of records")
	ErrInvalidPageSize           = errors.New("invalid page size")
)

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

type pageRequest struct {
	PageNumber uint
	PageSize   uint
	SortBy     string
	SortDesc   bool
}

type PageOpt func(*pageRequest)

func newPageRequest(idColumn string, opts ...PageOpt) (pageRequest, error) {
	req := pageRequest{
		PageNumber: 0,
		PageSize:   10,
		SortBy:     idColumn,
		SortDesc:   false,
	}

	for _, o := range opts {
		o(&req)
	}

	if err := req.validate(); err != nil {
		return pageRequest{}, fmt.Errorf("invalid page request: %w", err)
	}

	return req, nil
}

func (p pageRequest) validate() error {
	if p.PageSize <= 0 {
		return fmt.Errorf("page size must be greater than 0, but received %d: %w",
			p.PageSize, ErrInvalidPageSize)
	}
	return nil
}

func WithPageNumber(number uint) PageOpt {
	return func(p *pageRequest) {
		p.PageNumber = number
	}
}

func WithPageSize(size uint) PageOpt {
	return func(p *pageRequest) {
		p.PageSize = size
	}
}

func WithSort(column string) PageOpt {
	return func(p *pageRequest) {
		p.SortBy = column
		p.SortDesc = false
	}
}

func WithSortDesc(column string) PageOpt {
	return func(p *pageRequest) {
		p.SortBy = column
		p.SortDesc = true
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
	InitSchema(ctx context.Context, sql string) (sql.Result, error)

	GetAll(ctx context.Context) ([]T, error)

	GetPage(ctx context.Context, opts ...PageOpt) (Page[T], error)

	GetByID(ctx context.Context, id I) (T, error)

	GetByIDs(ctx context.Context, ids []I) ([]T, error)

	Create(ctx context.Context, e *T) (sql.Result, error)

	DeleteAll(ctx context.Context) (sql.Result, error)

	Delete(ctx context.Context, id I) (sql.Result, error)

	Update(ctx context.Context, id I, e *T) (sql.Result, error)

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

func (r *sqlCredo[T, I]) InitSchema(ctx context.Context, sql string) (sql.Result, error) {
	res, err := r.db.ExecContext(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return res, nil
}

func (r *sqlCredo[T, I]) GetAll(ctx context.Context) ([]T, error) {
	builder := goqu.From(r.table).Prepared(true)
	return r.selectValuesBuilderContext(ctx, builder.ToSQL)
}

func (r *sqlCredo[T, I]) GetPage(ctx context.Context, opts ...PageOpt) (Page[T], error) {
	params, err := newPageRequest(r.idColumn, opts...)
	if err != nil {
		return r.emptyPage, fmt.Errorf("unable to create page request: %w", err)
	}

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

	pageNumber := params.PageNumber

	totalPages := uint(math.Ceil(float64(totalRecords) / float64(params.PageSize)))

	return Page[T]{
		Number:     pageNumber,
		Size:       uint(len(pageRecords)),
		Total:      totalRecords,
		TotalPages: totalPages,
		Content:    pageRecords,
	}, nil
}

func (r *sqlCredo[T, I]) selectMany(ctx context.Context, req pageRequest) ([]T, error) {
	return r.selectValuesBuilderContext(ctx, r.selectManyQueryBuilder(req))
}

func (r *sqlCredo[T, I]) selectManyQueryBuilder(req pageRequest) queryBuilder {
	builder := goqu.From(r.table).Prepared(true)

	offset := req.PageNumber * req.PageSize

	builder = builder.Offset(offset)
	builder = builder.Limit(req.PageSize)

	if req.SortDesc {
		builder = builder.Order(goqu.I(req.SortBy).Desc())
	} else {
		builder = builder.Order(goqu.I(req.SortBy).Asc())
	}

	return builder.ToSQL
}

func (r *sqlCredo[T, I]) GetByID(ctx context.Context, id I) (T, error) {
	builder := goqu.From(r.table).
		Where(goqu.I(r.idColumn).Eq(id)).
		Prepared(true)

	var empty T

	entities, err := r.selectValuesBuilderContext(ctx, builder.ToSQL)
	if err != nil {
		return empty, fmt.Errorf("failed to select entities values: %w", err)
	}

	if len(entities) < 1 {
		return empty, ErrRecordNotFound
	}
	if len(entities) > 1 {
		return empty, fmt.Errorf("expected number of records 1, but found %d: %w",
			len(entities), ErrUnexpectedNumberOfRecords)
	}

	return entities[0], nil
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

func (r *sqlCredo[T, I]) Create(ctx context.Context, e *T) (sql.Result, error) {
	builder := goqu.Insert(r.table).
		Rows(e).
		Prepared(true)

	return r.execBuilderContext(ctx, builder.ToSQL)
}

func (r *sqlCredo[T, I]) DeleteAll(ctx context.Context) (sql.Result, error) {
	return r.Exec(ctx, r.truncateQuery)
}

func (r *sqlCredo[T, I]) Delete(ctx context.Context, id I) (sql.Result, error) {
	builder := goqu.Delete(r.table).
		Where(goqu.I(r.idColumn).Eq(id)).
		Prepared(true)

	return r.execBuilderContext(ctx, builder.ToSQL)
}

func (r *sqlCredo[T, I]) Update(ctx context.Context, id I, e *T) (sql.Result, error) {
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

func (r *sqlCredo[T, I]) execBuilderContext(ctx context.Context, builder queryBuilder) (sql.Result, error) {
	sql, args, err := builder()
	if err != nil {
		return nil, fmt.Errorf("failed to create sql query: %w", err)
	}

	res, err := r.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}

	return res, nil
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
