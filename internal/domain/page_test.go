package domain_test

import (
	"testing"

	sc "github.com/Klojer/sqlcredo"
	"github.com/Klojer/sqlcredo/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestPageParams_Validate(t *testing.T) {
	tests := []struct {
		name     string
		pageSize uint
		wantErr  bool
	}{
		{name: "valid page size", pageSize: 10, wantErr: false},
		{name: "zero page size", pageSize: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := domain.PageOpts{PageSize: tt.pageSize}
			err := params.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewPageParams(t *testing.T) {
	testCases := []struct {
		desc    string
		opts    []domain.PageOpt
		want    domain.PageOpts
		wantErr bool
	}{
		{
			desc: "no options",
			opts: nil,
			want: domain.PageOpts{
				PageNumber: domain.DefaultPageNumber,
				PageSize:   domain.DefaultPageSize,
				SortBy:     []string{"testId"},
				SortDesc:   domain.DefaultSortDesc,
			},
		},
		{
			desc:    "invalid opts",
			opts:    []domain.PageOpt{sc.WithPageSize(0)},
			want:    domain.PageOpts{},
			wantErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, gotErr := domain.NewPageOpts("testId", tC.opts...)

			if tC.wantErr {
				assert.Error(t, gotErr)
				assert.Equal(t, domain.PageOpts{}, got)
			} else {
				assert.NoError(t, gotErr)
				assert.Equal(t, tC.want, got)
			}
		})
	}
}

func TestWithPageNumber(t *testing.T) {
	params := &domain.PageOpts{}
	pageNumber := uint(2)

	sc.WithPageNumber(pageNumber)(params)

	assert.Equal(t, pageNumber, params.PageNumber)
}

func TestWithPageSize(t *testing.T) {
	params := &domain.PageOpts{}
	pageSize := uint(5)

	sc.WithPageSize(pageSize)(params)

	assert.Equal(t, pageSize, params.PageSize)
}

func TestWithSortBy(t *testing.T) {
	params := &domain.PageOpts{}
	column := "name"

	sc.WithSortBy(column)(params)

	assert.Equal(t, []string{column}, params.SortBy)
}

func TestWithSortDesc(t *testing.T) {
	params := &domain.PageOpts{}

	sc.WithSortDesc("name")(params)

	assert.True(t, params.SortDesc)
}

func TestNewEmptyPage(t *testing.T) {
	got := sc.NewEmptyPage[string]()

	assert.Equal(t, domain.Page[string]{}, got)
}
