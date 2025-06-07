package page

import (
	"context"
	"fmt"
	"math"

	"gitlab.com/onrooh/sqlcredo/pkg/model"

	"github.com/doug-martin/goqu/v9"
)

const (
	countQueryTemplate = "SELECT COUNT(*) FROM %s;"
)

type PageResolver[T any] struct {
	table      model.TableInfo
	executor   model.SQLExecutor
	countQuery string
	emptyPage  model.Page[T]
}

var _ model.PageResolver[any] = &PageResolver[any]{}

func NewPageResolver[T any](table model.TableInfo, executor model.SQLExecutor) *PageResolver[T] {
	return &PageResolver[T]{
		table:      table,
		executor:   executor,
		countQuery: fmt.Sprintf(countQueryTemplate, table.Name),
		emptyPage:  newEmptyPage[T](),
	}
}

func (r *PageResolver[T]) GetPage(ctx context.Context, opts ...model.PageOpt) (model.Page[T], error) {
	req, err := newPageRequest(r.table.IDColumn, opts...)
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

func (r *PageResolver[T]) createPageQueryBuilder(req model.PageRequest) (string, []interface{}, error) {
	builder := goqu.From(r.table.Name).Prepared(true)

	offset := req.PageNumber * req.PageSize

	builder = builder.Offset(offset)
	builder = builder.Limit(req.PageSize)

	if req.SortDesc {
		builder = builder.Order(goqu.I(req.SortBy).Desc())
	} else {
		builder = builder.Order(goqu.I(req.SortBy).Asc())
	}

	return builder.ToSQL()
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

func newPageRequest(idColumn string, opts ...model.PageOpt) (model.PageRequest, error) {
	req := model.PageRequest{
		PageNumber: 0,
		PageSize:   10,
		SortBy:     idColumn,
		SortDesc:   false,
	}

	for _, o := range opts {
		o(&req)
	}

	if err := req.Validate(); err != nil {
		return model.PageRequest{}, fmt.Errorf("invalid page request: %w", err)
	}

	return req, nil
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
