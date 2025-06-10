package page

import (
	"context"
	"fmt"
	"math"

	"github.com/Klojer/sqlcredo/internal/goquext"
	"github.com/Klojer/sqlcredo/internal/table"
	"github.com/Klojer/sqlcredo/pkg/model"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

const (
	countQueryTemplate = `SELECT COUNT(%s) FROM %s;`
)

type PageResolver[T any] struct {
	table      table.Info
	executor   model.SQLExecutor
	countQuery string
	emptyPage  model.Page[T]
	dialect    goqu.DialectWrapper
}

var _ model.PageResolver[any] = &PageResolver[any]{}

func NewPageResolver[T any](table table.Info,
	executor model.SQLExecutor, driver string,
) *PageResolver[T] {
	return &PageResolver[T]{
		table:      table,
		executor:   executor,
		countQuery: fmt.Sprintf(countQueryTemplate, table.IDColumn, table.Name),
		emptyPage:  newEmptyPage[T](),
		dialect:    goquext.CreateDialect(driver),
	}
}

func (r *PageResolver[T]) GetPage(ctx context.Context, opts ...model.PageOpt) (model.Page[T], error) {
	req, err := newPageParams(r.table.IDColumn, opts...)
	if err != nil {
		return r.emptyPage, fmt.Errorf("unable to create page request: %w", err)
	}

	query, args, err := r.createPageQueryBuilder(req)
	if err != nil {
		return r.emptyPage, fmt.Errorf("unable to create page sql query: %w", err)
	}

	pageRecords, err := r.selectMany(ctx, query, args...)
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

	totalPages := uint(math.Ceil(float64(totalRecords) / float64(req.PageSize)))

	return model.Page[T]{
		Number:     req.PageNumber,
		Size:       uint(len(pageRecords)),
		Total:      totalRecords,
		TotalPages: totalPages,
		Content:    pageRecords,
	}, nil
}

func (r *PageResolver[T]) createPageQueryBuilder(params model.PageParams) (string, []any, error) {
	builder := r.dialect.From(r.table.Name).Prepared(true)
	builder = builder.Offset(params.PageNumber * params.PageSize)
	builder = builder.Limit(params.PageSize)
	builder = builder.Order(buildOrderExprs(params)...)
	return builder.ToSQL()
}

func buildOrderExprs(params model.PageParams) []exp.OrderedExpression {
	orderExprs := make([]exp.OrderedExpression, 0, len(params.SortBy))
	for _, s := range params.SortBy {
		if params.SortDesc {
			orderExprs = append(orderExprs, goqu.I(s).Desc())
		} else {
			orderExprs = append(orderExprs, goqu.I(s).Asc())
		}
	}
	return orderExprs
}

func (r *PageResolver[T]) Count(ctx context.Context) (uint64, error) {
	var res uint64
	if err := r.executor.SelectOne(ctx, &res, r.countQuery); err != nil {
		return 0, fmt.Errorf("unable to count records: %w", err)
	}
	return res, nil
}

func (r *PageResolver[T]) selectMany(ctx context.Context, query string, args ...any) ([]T, error) {
	var records []T
	if err := r.executor.SelectMany(ctx, &records, query, args...); err != nil {
		return nil, fmt.Errorf("unable to load page records: %w", err)
	}
	return records, nil
}

func newPageParams(idColumn string, opts ...model.PageOpt) (model.PageParams, error) {
	params := model.PageParams{
		PageNumber: 0,
		PageSize:   10,
		SortDesc:   false,
	}

	for _, o := range opts {
		o(&params)
	}

	if err := params.Validate(); err != nil {
		return model.PageParams{}, fmt.Errorf("invalid page params: %w", err)
	}

	if params.SortBy == nil {
		params.SortBy = []string{idColumn}
	}

	return params, nil
}

func newEmptyPage[T any]() model.Page[T] {
	return model.Page[T]{
		Number:     0,
		Size:       0,
		Total:      0,
		TotalPages: 0,
		Content:    nil,
	}
}
