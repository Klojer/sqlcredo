package model

import (
	"context"
	"fmt"
)

// Page represents a paginated result set containing items of type T.
// It includes metadata about the current page, total items, and the actual content.
type Page[T any] struct {
	Number     uint   // Current page number (0-based)
	Size       uint   // Number of items in the current page
	Total      uint64 // Total number of items across all pages
	TotalPages uint   // Total number of pages
	Content    []T    // Slice containing the page's items
}

// PageParams defines the parameters for pagination and sorting.
type PageParams struct {
	PageNumber uint     // The current page number (0-based)
	PageSize   uint     // Number of items per page
	SortBy     []string // List of columns to sort by
	SortDesc   bool     // If true, sort in descending order
}

// PageOpt is a function type that modifies PageParams.
// It follows the functional options pattern for configuring pagination parameters.
type PageOpt func(*PageParams)

func (p PageParams) Validate() error {
	if p.PageSize <= 0 {
		return fmt.Errorf("page size must be greater than 0, but received %d: %w",
			p.PageSize, ErrInvalidPageSize)
	}
	return nil
}

// WithPageNumber sets the page number.
// The page number is 0-based, meaning the first page is 0.
func WithPageNumber(number uint) PageOpt {
	return func(p *PageParams) {
		p.PageNumber = number
	}
}

// WithPageSize sets the number of items per page.
// The page size must be greater than 0.
func WithPageSize(size uint) PageOpt {
	return func(p *PageParams) {
		p.PageSize = size
	}
}

// WithSortBy adds a column to sort by.
// Multiple calls will append to the list of sort columns.
// If no sort columns are specified, the default is to sort by ID.
func WithSortBy(column string) PageOpt {
	return func(p *PageParams) {
		if p.SortBy == nil {
			p.SortBy = make([]string, 0)
		}
		p.SortBy = append(p.SortBy, column)
	}
}

// WithSortDesc sets the sort order to descending.
// By default, the sort order is ascending.
func WithSortDesc(column string) PageOpt {
	return func(p *PageParams) {
		p.SortDesc = true
	}
}

// PageResolver is an interface for retrieving paginated results of type T.
type PageResolver[T any] interface {
	// GetPage retrieves a single page of results based on the provided pagination options.
	GetPage(ctx context.Context, opts ...PageOpt) (Page[T], error)

	// Count returns the total number of items available across all pages.
	// This is useful for calculating total pages and displaying pagination metadata.
	Count(ctx context.Context) (uint64, error)
}
