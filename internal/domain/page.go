package domain

import (
	"context"
	"fmt"
)

const (
	DefaultPageNumber = 0
	DefaultPageSize   = 10
	DefaultSortDesc   = false

	DefaultContentInitSize = 10
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

// PageOpts defines the parameters for pagination and sorting.
type PageOpts struct {
	PageNumber uint     // The current page number (0-based)
	PageSize   uint     // Number of items per page
	SortBy     []string // List of columns to sort by
	SortDesc   bool     // If true, sort in descending order
}

// PageOpt is a function type that modifies PageParams.
// It follows the functional options pattern for configuring pagination parameters.
type PageOpt func(*PageOpts)

// NewPageOpts creates a new PageOpts with the given options and validate it.
// If no SortBy specified the idColumn will be used for sorting.
func NewPageOpts(idColumn string, opts ...PageOpt) (PageOpts, error) {
	optsObj := PageOpts{
		PageNumber: DefaultPageNumber,
		PageSize:   DefaultPageSize,
		SortDesc:   DefaultSortDesc,
	}

	for _, o := range opts {
		o(&optsObj)
	}

	if err := optsObj.Validate(); err != nil {
		return PageOpts{}, fmt.Errorf("invalid page params: %w", err)
	}

	if optsObj.SortBy == nil {
		optsObj.SortBy = []string{idColumn}
	}

	return optsObj, nil
}

func (p PageOpts) Validate() error {
	if p.PageSize <= 0 {
		return fmt.Errorf("page size must be greater than 0, but received %d: %w",
			p.PageSize, ErrInvalidPageSize)
	}
	return nil
}

// PageResolver is an interface for retrieving paginated results of type T.
type PageResolver[T any] interface {
	// GetPage retrieves a single page of results based on the provided pagination options.
	GetPage(ctx context.Context, dest *Page[T], opts ...PageOpt) error

	// Count returns the total number of items available across all pages.
	// This is useful for calculating total pages and displaying pagination metadata.
	Count(ctx context.Context) (uint64, error)
}

type NewPageOptsObj[T any] struct {
	Content         *[]T
	ContentInitSize *int
}

type NewPageOpt[T any] func(*NewPageOptsObj[T])
