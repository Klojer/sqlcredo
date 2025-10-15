package api_test

import (
	"testing"

	"github.com/Klojer/sqlcredo/pkg/api"

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
			params := api.PageOpts{PageSize: tt.pageSize}
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
		opts    []api.PageOpt
		want    api.PageOpts
		wantErr bool
	}{
		{
			desc: "no options",
			opts: nil,
			want: api.PageOpts{
				PageNumber: api.DefaultPageNumber,
				PageSize:   api.DefaultPageSize,
				SortBy:     []string{"testId"},
				SortDesc:   api.DefaultSortDesc,
			},
		},
		{
			desc:    "invalid opts",
			opts:    []api.PageOpt{api.WithPageSize(0)},
			want:    api.PageOpts{},
			wantErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, gotErr := api.NewPageOpts("testId", tC.opts...)

			if tC.wantErr {
				assert.Error(t, gotErr)
				assert.Equal(t, api.PageOpts{}, got)
			} else {
				assert.NoError(t, gotErr)
				assert.Equal(t, tC.want, got)
			}
		})
	}
}

func TestWithPageNumber(t *testing.T) {
	params := &api.PageOpts{}
	pageNumber := uint(2)

	api.WithPageNumber(pageNumber)(params)

	assert.Equal(t, pageNumber, params.PageNumber)
}

func TestWithPageSize(t *testing.T) {
	params := &api.PageOpts{}
	pageSize := uint(5)

	api.WithPageSize(pageSize)(params)

	assert.Equal(t, pageSize, params.PageSize)
}

func TestWithSortBy(t *testing.T) {
	params := &api.PageOpts{}
	column := "name"

	api.WithSortBy(column)(params)

	assert.Equal(t, []string{column}, params.SortBy)
}

func TestWithSortDesc(t *testing.T) {
	params := &api.PageOpts{}

	api.WithSortDesc("name")(params)

	assert.True(t, params.SortDesc)
}

func TestNewEmptyPage(t *testing.T) {
	got := api.NewEmptyPage[string]()

	assert.Equal(t, api.Page[string]{}, got)
}
