package sqlcredo

import (
	"testing"

	"github.com/Klojer/sqlcredo/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestWithPageNumber(t *testing.T) {
	params := &domain.PageOpts{}
	pageNumber := uint(2)

	WithPageNumber(pageNumber)(params)

	assert.Equal(t, pageNumber, params.PageNumber)
}

func TestWithPageSize(t *testing.T) {
	params := &domain.PageOpts{}
	pageSize := uint(5)

	WithPageSize(pageSize)(params)

	assert.Equal(t, pageSize, params.PageSize)
}

func TestWithSortBy(t *testing.T) {
	params := &domain.PageOpts{}
	column := "name"

	WithSortBy(column)(params)

	assert.Equal(t, []string{column}, params.SortBy)
}

func TestWithSortDesc(t *testing.T) {
	params := &domain.PageOpts{}

	WithSortDesc("name")(params)

	assert.True(t, params.SortDesc)
}

func TestNewEmptyPage(t *testing.T) {
	got := NewEmptyPage[string]()

	expected := Page[string]{
		Number:     0,
		Size:       0,
		Total:      0,
		TotalPages: 0,
		Content:    nil,
	}
	assert.Equal(t, expected, got)
}

func TestWithNewPageContent(t *testing.T) {
	dest := &[]string{"existing", "content"}
	optsObj := &domain.NewPageOptsObj[string]{}

	WithNewPageContent(dest)(optsObj)

	assert.Equal(t, dest, optsObj.Content)
}

func TestWithNewPageContentInitSize(t *testing.T) {
	initSize := 25
	optsObj := &domain.NewPageOptsObj[string]{}

	WithNewPageContentInitSize[string](initSize)(optsObj)

	assert.Equal(t, &initSize, optsObj.ContentInitSize)
}

func TestNewPage(t *testing.T) {
	t.Run("default initialization", func(t *testing.T) {
		page := NewPage[string]()

		assert.Equal(t, uint(0), page.Number)
		assert.Equal(t, uint(0), page.Size)
		assert.Equal(t, uint64(0), page.Total)
		assert.Equal(t, uint(0), page.TotalPages)
		assert.NotNil(t, page.Content)
		assert.Equal(t, 0, len(page.Content))
		assert.Equal(t, domain.DefaultContentInitSize, cap(page.Content))
	})

	t.Run("with custom content", func(t *testing.T) {
		customContent := &[]string{"pre", "filled"}
		page := NewPage(WithNewPageContent(customContent))

		assert.Equal(t, uint(0), page.Number)
		assert.Equal(t, uint(0), page.Size)
		assert.Equal(t, uint64(0), page.Total)
		assert.Equal(t, uint(0), page.TotalPages)
		assert.Equal(t, customContent, &page.Content)
		assert.Equal(t, []string{"pre", "filled"}, page.Content)
	})

	t.Run("with custom init size", func(t *testing.T) {
		customSize := 25
		page := NewPage(WithNewPageContentInitSize[string](customSize))

		assert.Equal(t, uint(0), page.Number)
		assert.Equal(t, uint(0), page.Size)
		assert.Equal(t, uint64(0), page.Total)
		assert.Equal(t, uint(0), page.TotalPages)
		assert.NotNil(t, page.Content)
		assert.Equal(t, 0, len(page.Content))
		assert.Equal(t, customSize, cap(page.Content))
	})
}
