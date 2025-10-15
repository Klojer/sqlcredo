package page

import (
	"context"
	"fmt"
	"math"

	"github.com/Klojer/sqlcredo/internal/goquext"
	"github.com/Klojer/sqlcredo/internal/table"
	"github.com/Klojer/sqlcredo/pkg/api"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

const (
	countQueryTemplate = `SELECT COUNT(%s) FROM %s;`
)

type PageResolver[T any] struct {
	table      table.Info
	executor   api.SQLExecutor
	countQuery string
	emptyPage  api.Page[T]
	dialect    goqu.DialectWrapper
}

var _ api.PageResolver[any] = &PageResolver[any]{}

func NewPageResolver[T any](table table.Info,
	executor api.SQLExecutor, driver string,
) *PageResolver[T] {
	return &PageResolver[T]{
		table:      table,
		executor:   executor,
		countQuery: fmt.Sprintf(countQueryTemplate, table.IDColumn, table.Name),
		emptyPage:  api.NewEmptyPage[T](),
		dialect:    goqu.Dialect(goquext.CreateDialectString(driver)),
	}
}

func (r *PageResolver[T]) GetPage(ctx context.Context, dest *api.Page[T], opts ...api.PageOpt) error {
	req, err := api.NewPageOpts(r.table.IDColumn, opts...)
	if err != nil {
		return fmt.Errorf("unable to create page params: %w", err)
	}

	query, args, err := r.createPageQueryBuilder(req)
	if err != nil {
		return fmt.Errorf("unable to create page sql query: %w", err)
	}

	err = r.selectMany(ctx, &dest.Content, query, args...)
	if err != nil {
		return fmt.Errorf("unable to get page items: %w", err)
	}

	totalRecords, err := r.Count(ctx)
	if err != nil {
		return fmt.Errorf("unable to count all items: %w", err)
	}

	if len(dest.Content) == 0 {
		return nil
	}

	totalPages := uint(math.Ceil(float64(totalRecords) / float64(req.PageSize)))

	dest.Number = req.PageNumber
	dest.Size = uint(len(dest.Content))
	dest.Total = totalRecords
	dest.TotalPages = totalPages

	return nil
}

func (r *PageResolver[T]) createPageQueryBuilder(params api.PageOpts) (string, []any, error) {
	builder := r.dialect.From(r.table.Name).Prepared(true)
	builder = builder.Offset(params.PageNumber * params.PageSize)
	builder = builder.Limit(params.PageSize)
	builder = builder.Order(buildOrderExprs(params)...)
	return builder.ToSQL()
}

func buildOrderExprs(params api.PageOpts) []exp.OrderedExpression {
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

func (r *PageResolver[T]) selectMany(ctx context.Context, dest *[]T, query string, args ...any) error {
	if err := r.executor.SelectMany(ctx, dest, query, args...); err != nil {
		return fmt.Errorf("unable to load page records: %w", err)
	}
	return nil
}
