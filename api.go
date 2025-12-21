package sqlcredo

import (
	"github.com/Klojer/sqlcredo/internal/domain"
)

type (
	DebugFunc                                = domain.DebugFunc
	SQLExecutor                              = domain.SQLExecutor
	CRUD[T any, I comparable]                = domain.CRUD[T, I]
	Page[T any]                              = domain.Page[T]
	PageResolver[T any]                      = domain.PageResolver[T]
	PageOpt                                  = domain.PageOpt
	NewPageOpt[T any]                        = domain.NewPageOpt[T]
	Transaction[T any, I comparable]         = domain.Transaction[T, I]
	TransactionExecutor[T any, I comparable] = domain.TransactionExecutor[T, I]
)

// WithPageNumber sets the page number.
// The page number is 0-based, meaning the first page is 0.
func WithPageNumber(number uint) PageOpt {
	return func(p *domain.PageOpts) {
		p.PageNumber = number
	}
}

// WithPageSize sets the number of items per page.
// The page size must be greater than 0.
func WithPageSize(size uint) PageOpt {
	return func(p *domain.PageOpts) {
		p.PageSize = size
	}
}

// WithSortBy adds a column to sort by.
// Multiple calls will append to the list of sort columns.
// If no sort columns are specified, the default is to sort by ID.
func WithSortBy(column string) PageOpt {
	return func(p *domain.PageOpts) {
		if p.SortBy == nil {
			p.SortBy = make([]string, 0)
		}
		p.SortBy = append(p.SortBy, column)
	}
}

// WithSortDesc sets the sort order to descending.
// By default, the sort order is ascending.
func WithSortDesc(column string) PageOpt {
	return func(p *domain.PageOpts) {
		p.SortDesc = true
	}
}

// NewEmptyPage creates a page with no content and zero counters.
func NewEmptyPage[T any]() Page[T] {
	return Page[T]{
		Number:     0,
		Size:       0,
		Total:      0,
		TotalPages: 0,
		Content:    nil,
	}
}

// WithNewPageContent sets page's the desitination slice.
// If set, ContentInitSize will be ignored
func WithNewPageContent[T any](dest *[]T) NewPageOpt[T] {
	return func(p *domain.NewPageOptsObj[T]) {
		p.Content = dest
	}
}

// WithNewPageContentInitSize sets page's the desitination init size.
func WithNewPageContentInitSize[T any](initSize int) NewPageOpt[T] {
	return func(p *domain.NewPageOptsObj[T]) {
		p.ContentInitSize = &initSize
	}
}

// NewPage creates a page to fill.
// By default create new slice for content with capacity equal to DefaultContentInitSize
func NewPage[T any](opts ...NewPageOpt[T]) Page[T] {
	optsObj := domain.NewPageOptsObj[T]{
		ContentInitSize: &[]int{domain.DefaultContentInitSize}[0],
	}

	for _, o := range opts {
		o(&optsObj)
	}

	var content []T
	if optsObj.Content != nil {
		content = *optsObj.Content
	} else {
		content = make([]T, 0, *optsObj.ContentInitSize)
	}

	return Page[T]{
		Number:     0,
		Size:       0,
		Total:      0,
		TotalPages: 0,
		Content:    content,
	}
}
