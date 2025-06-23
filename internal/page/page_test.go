package page_test

import (
	"context"
	"testing"
	"time"

	"github.com/Klojer/sqlcredo/internal/mocks"
	"github.com/Klojer/sqlcredo/internal/page"
	"github.com/Klojer/sqlcredo/internal/table"
	"github.com/Klojer/sqlcredo/pkg/api"

	"github.com/stretchr/testify/assert"
)

func TestPageResolver_GetPage(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.GetPage(ctx, api.WithPageNumber(0), api.WithPageSize(10))
	assert.NoError(t, err)
	assert.Equal(t, c.Executor.Queries[0], "SELECT * FROM `test_table` ORDER BY `id` ASC LIMIT ?")
	assert.Equal(t, c.Executor.Args[0], []any{int64(10)})
}

func TestPageResolver_Count(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, c.Executor.Queries[0], "SELECT COUNT(id) FROM test_table;")
	assert.Nil(t, c.Executor.Args[0])
}

type testCaseData struct {
	ctx       context.Context
	ctxCancel func()

	Executor  *mocks.SQLExecutor
	UnderTest api.PageResolver[testObj]
}

func newTestCase(t *testing.T) (*testCaseData, context.Context) {
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
	c.ctxCancel()
}

type testObj struct {
	Id   string `db:"id"`
	Name string `db:"name"`
}
