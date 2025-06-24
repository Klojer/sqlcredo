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
			params := api.PageParams{PageSize: tt.pageSize}
			err := params.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWithPageNumber(t *testing.T) {
	params := &api.PageParams{}
	pageNumber := uint(2)

	api.WithPageNumber(pageNumber)(params)

	assert.Equal(t, pageNumber, params.PageNumber)
}

func TestWithPageSize(t *testing.T) {
	params := &api.PageParams{}
	pageSize := uint(5)

	api.WithPageSize(pageSize)(params)

	assert.Equal(t, pageSize, params.PageSize)
}

func TestWithSortBy(t *testing.T) {
	params := &api.PageParams{}
	column := "name"

	api.WithSortBy(column)(params)

	assert.Equal(t, []string{column}, params.SortBy)
}

func TestWithSortDesc(t *testing.T) {
	params := &api.PageParams{}

	api.WithSortDesc("name")(params)

	assert.True(t, params.SortDesc)
}
