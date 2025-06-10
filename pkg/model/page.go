package model

import (
	"context"
	"fmt"
)

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
