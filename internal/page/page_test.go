package page_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sc "github.com/Klojer/sqlcredo"
	"github.com/Klojer/sqlcredo/internal/domain"
	"github.com/Klojer/sqlcredo/internal/mocks"
	"github.com/Klojer/sqlcredo/internal/page"
	"github.com/Klojer/sqlcredo/internal/table"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testPageData = []testObj{
	{ID: "0", Name: "n0"},
	{ID: "1", Name: "n1"},
}

func TestPageResolver_GetPage(t *testing.T) {
	testCases := []struct {
		desc          string
		configureMock func(context.Context, *testCaseData)
		opts          []domain.PageOpt
		want          domain.Page[testObj]
		wantErr       string
	}{
		{
			desc: "base positive case",
			opts: []domain.PageOpt{sc.WithPageNumber(0), sc.WithPageSize(2)},
			configureMock: func(ctx context.Context, c *testCaseData) {
				c.Executor.On("SelectMany", ctx, mock.Anything,
					"SELECT * FROM `test_table` ORDER BY `id` ASC LIMIT ?", []any{int64(2)}).
					Return(nil)
				c.Executor.SetOnSelectManyCb(replaceDestWithTestObjSlice(t, testPageData))
				c.Executor.On("SelectOne", ctx, mock.Anything,
					"SELECT COUNT(id) FROM test_table;", mock.Anything).
					Return(nil)
				c.Executor.SetOnSelectOneCb(replaceDestWithUint64(t, 3))
			},
			want: domain.Page[testObj]{
				Number:     0,
				Size:       2,
				Total:      3,
				TotalPages: 2,
				Content:    testPageData,
			},
		},
		{
			desc: "empty page",
			opts: []domain.PageOpt{sc.WithPageNumber(0), sc.WithPageSize(2)},
			configureMock: func(ctx context.Context, c *testCaseData) {
				c.Executor.On("SelectMany", ctx, mock.Anything,
					"SELECT * FROM `test_table` ORDER BY `id` ASC LIMIT ?", []any{int64(2)}).
					Return(nil)
				c.Executor.SetOnSelectManyCb(replaceDestWithTestObjSlice(t, []testObj{}))
				c.Executor.On("SelectOne", ctx, mock.Anything,
					"SELECT COUNT(id) FROM test_table;", mock.Anything).
					Return(nil)
				c.Executor.SetOnSelectOneCb(replaceDestWithUint64(t, 0))
			},
			want: sc.NewPage[testObj](),
		},
		{
			desc: "reverse sort order",
			opts: []domain.PageOpt{sc.WithPageNumber(0), sc.WithPageSize(2), sc.WithSortDesc("id")},
			configureMock: func(ctx context.Context, c *testCaseData) {
				c.Executor.On("SelectMany", ctx, mock.Anything,
					"SELECT * FROM `test_table` ORDER BY `id` DESC LIMIT ?", []any{int64(2)}).
					Return(nil)
				c.Executor.SetOnSelectManyCb(replaceDestWithTestObjSlice(t, []testObj{}))
				c.Executor.On("SelectOne", ctx, mock.Anything,
					"SELECT COUNT(id) FROM test_table;", mock.Anything).
					Return(nil)
				c.Executor.SetOnSelectOneCb(replaceDestWithUint64(t, 0))
			},
			want: sc.NewPage[testObj](),
		},
		{
			desc:          "invalid opts",
			opts:          []domain.PageOpt{sc.WithPageSize(0)},
			configureMock: func(ctx context.Context, c *testCaseData) {},
			wantErr:       "unable to create page params",
		},
		{
			desc: "select many fail",
			opts: []domain.PageOpt{sc.WithPageNumber(0), sc.WithPageSize(2)},
			configureMock: func(ctx context.Context, c *testCaseData) {
				c.Executor.On("SelectMany", ctx, mock.Anything,
					"SELECT * FROM `test_table` ORDER BY `id` ASC LIMIT ?", []any{int64(2)}).
					Return(errors.New("unable to exec select query"))
			},
			wantErr: "unable to get page items",
		},
		{
			desc: "select one fail",
			opts: []domain.PageOpt{sc.WithPageNumber(0), sc.WithPageSize(2)},
			configureMock: func(ctx context.Context, c *testCaseData) {
				c.Executor.On("SelectMany", ctx, mock.Anything,
					"SELECT * FROM `test_table` ORDER BY `id` ASC LIMIT ?", []any{int64(2)}).
					Return(nil)
				c.Executor.On("SelectOne", ctx, mock.Anything,
					"SELECT COUNT(id) FROM test_table;", mock.Anything).
					Return(errors.New("unable to exec count query"))
			},
			wantErr: "unable to count all items",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			c, ctx := newTestCase(t)
			tC.configureMock(ctx, c)

			content := make([]testObj, 0)
			got := sc.NewPage(sc.WithNewPageContent(&content))
			gotErr := c.UnderTest.GetPage(ctx, &got, tC.opts...)

			if tC.wantErr != "" {
				assert.ErrorContains(t, gotErr, tC.wantErr)
				assert.Equal(t, sc.NewPage[testObj](), got)
			} else {
				assert.NoError(t, gotErr)
				assert.Equal(t, tC.want, got)
			}
		})
	}
}

func TestPageResolver_Count(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectOne", ctx, mock.Anything,
		"SELECT COUNT(id) FROM test_table;", mock.Anything).
		Return(nil)

	_, err := c.UnderTest.Count(ctx)

	assert.NoError(t, err)
}

func replaceDestWithTestObjSlice(t *testing.T, data []testObj) mocks.OnSelectCb {
	return func(ctx context.Context, dest any, query string, args ...any) {
		t.Helper()
		switch v := dest.(type) {
		case *[]testObj:
			*v = data
		default:
			t.Log("unable to cast dest to []testObj")
			t.Fail()
		}
	}
}

func replaceDestWithUint64(t *testing.T, data uint64) mocks.OnSelectCb {
	return func(ctx context.Context, dest any, query string, args ...any) {
		t.Helper()
		switch v := dest.(type) {
		case *uint64:
			*v = data
		default:
			t.Log("unable to cast dest to uint64")
			t.Fail()
		}
	}
}

type testCaseData struct {
	ctx       context.Context
	ctxCancel func()

	Executor  *mocks.SQLExecutor
	UnderTest domain.PageResolver[testObj]
}

func newTestCase(t *testing.T) (*testCaseData, context.Context) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	executor := mocks.NewSQLExecutor()
	tableInfo := table.Info{Name: "test_table", IDColumn: "id"}

	c := &testCaseData{
		ctx:       ctx,
		ctxCancel: cancel,

		Executor:  executor,
		UnderTest: page.NewPageResolver[testObj](tableInfo, executor, "sqlite3"),
	}

	t.Cleanup(func() {
		c.TearDown(t)
	})

	return c, ctx
}

func (c *testCaseData) TearDown(t *testing.T) {
	t.Helper()

	c.Executor.AssertExpectations(t)
	c.ctxCancel()
}

type testObj struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}
