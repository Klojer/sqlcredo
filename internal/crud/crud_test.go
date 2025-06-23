package crud_test

import (
	"context"
	"testing"
	"time"

	"github.com/Klojer/sqlcredo/internal/crud"
	"github.com/Klojer/sqlcredo/internal/mocks"
	"github.com/Klojer/sqlcredo/internal/table"
	"github.com/Klojer/sqlcredo/pkg/api"

	"github.com/stretchr/testify/assert"
)

func TestCRUD_GetAll(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.GetAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, c.Executor.Queries[0], "SELECT * FROM `test_table`")
}

func TestCRUD_GetByID(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.GetByID(ctx, "test_id")
	assert.NoError(t, err)
	assert.Equal(t, c.Executor.Queries[0], "SELECT * FROM `test_table` WHERE (`id` = ?)")
	assert.Equal(t, c.Executor.Args[0], []any{"test_id"})
}

func TestCRUD_GetByIDs(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.GetByIDs(ctx, []string{"0", "3", "16"})
	assert.NoError(t, err)
	assert.Equal(t, c.Executor.Queries[0], "SELECT * FROM `test_table` WHERE (`id` IN (?, ?, ?)) ORDER BY `id` ASC")
	assert.Equal(t, c.Executor.Args[0], []any{"0", "3", "16"})
}

func TestCRUD_Create(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.Create(ctx, &testObj{Id: "12", Name: "test12"})
	assert.NoError(t, err)
	assert.Contains(t, c.Executor.Queries[0], "INSERT INTO `test_table` (`id`, `name`) VALUES (?, ?)")
	assert.Equal(t, c.Executor.Args[0], []any{"12", "test12"})
}

func TestCRUD_DeleteAll(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.DeleteAll(ctx)
	assert.NoError(t, err)
	assert.Contains(t, c.Executor.Queries[0], "DELETE FROM test_table;")
	assert.Nil(t, c.Executor.Args[0])
}

func TestCRUD_Delete(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.Delete(ctx, "test_id")
	assert.NoError(t, err)
	assert.Contains(t, c.Executor.Queries[0], "DELETE FROM `test_table` WHERE (`id` = ?)")
	assert.Equal(t, c.Executor.Args[0], []any{"test_id"})
}

func TestCRUD_Update(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.Update(ctx, "12", &testObj{Id: "12", Name: "test12"})
	assert.NoError(t, err)
	assert.Contains(t, c.Executor.Queries[0], "UPDATE `test_table` SET `id`=?,`name`=? WHERE (`id` = ?)")
	assert.Equal(t, c.Executor.Args[0], []any{"12", "test12", "12"})
}

type testCaseData struct {
	ctx       context.Context
	ctxCancel func()

	Executor  *mocks.SQLExecutor
	UnderTest api.CRUD[testObj, string]
}

func newTestCase(t *testing.T) (*testCaseData, context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	executor := mocks.NewSQLExecutor()
	tableInfo := table.Info{Name: "test_table", IDColumn: "id"}

	c := &testCaseData{
		ctx:       ctx,
		ctxCancel: cancel,

		Executor:  executor,
		UnderTest: crud.NewCRUD[testObj, string](tableInfo, executor, "sqlite3"),
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
