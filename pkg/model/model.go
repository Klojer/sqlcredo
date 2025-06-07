package model

import (
	"context"
	"database/sql"
	"fmt"
)

type TableInfo struct {
	Name     string
	IDColumn string
}

type (
	DebugFunc func(sql string, args ...any)
)

type SQLExecutor interface {
	SelectOne(ctx context.Context, dest any, query string, args ...any) error
	SelectMany(ctx context.Context, dest any, query string, args ...any) error
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type CRUD[T any, I comparable] interface {
	GetAll(ctx context.Context) ([]T, error)
	GetByID(ctx context.Context, id I) (T, error)
	GetByIDs(ctx context.Context, ids []I) ([]T, error)
	Create(ctx context.Context, e *T) (sql.Result, error)
	DeleteAll(ctx context.Context) (sql.Result, error)
	Delete(ctx context.Context, id I) (sql.Result, error)
	Update(ctx context.Context, id I, e *T) (sql.Result, error)
}

type Page[T any] struct {
	Number     uint
	Size       uint
	Total      uint64
	TotalPages uint
	Content    []T
}

type PageRequest struct {
	PageNumber uint
	PageSize   uint
	SortBy     string
	SortDesc   bool
}

type PageOpt func(*PageRequest)

func (p PageRequest) Validate() error {
	if p.PageSize <= 0 {
		return fmt.Errorf("page size must be greater than 0, but received %d: %w",
			p.PageSize, ErrInvalidPageSize)
	}
	return nil
}

func WithPageNumber(number uint) PageOpt {
	return func(p *PageRequest) {
		p.PageNumber = number
	}
}

func WithPageSize(size uint) PageOpt {
	return func(p *PageRequest) {
		p.PageSize = size
	}
}

func WithSort(column string) PageOpt {
	return func(p *PageRequest) {
		p.SortBy = column
		p.SortDesc = false
	}
}

func WithSortDesc(column string) PageOpt {
	return func(p *PageRequest) {
		p.SortBy = column
		p.SortDesc = true
	}
}

type PageResolver[T any] interface {
	GetPage(ctx context.Context, opts ...PageOpt) (Page[T], error)
	Count(ctx context.Context) (uint64, error)
}
