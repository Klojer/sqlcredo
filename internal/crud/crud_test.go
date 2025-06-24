package crud_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Klojer/sqlcredo/internal/crud"
	"github.com/Klojer/sqlcredo/internal/mocks"
	"github.com/Klojer/sqlcredo/internal/table"
	"github.com/Klojer/sqlcredo/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCRUD_GetAll(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectMany", ctx, mock.Anything, "SELECT * FROM `test_table`", mock.Anything).
		Return(nil)

	_, err := c.UnderTest.GetAll(ctx)

	assert.NoError(t, err)
}

func TestCRUD_GetByID(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectOne", ctx, mock.Anything,
		"SELECT * FROM `test_table` WHERE (`id` = ?)", []any{"test_id"}).
		Return(nil)

	_, err := c.UnderTest.GetByID(ctx, "test_id")

	assert.NoError(t, err)
}

func TestCRUD_GetByID_Error(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectOne", ctx, mock.Anything,
		"SELECT * FROM `test_table` WHERE (`id` = ?)", []any{"non_existent_id"}).
		Return(fmt.Errorf("database error"))

	_, err := c.UnderTest.GetByID(ctx, "non_existent_id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to select record")
}

func TestCRUD_GetByIDs(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectMany", ctx, mock.Anything,
		"SELECT * FROM `test_table` WHERE (`id` IN (?, ?, ?)) ORDER BY `id` ASC",
		[]any{"0", "3", "16"}).
		Return(nil)

	_, err := c.UnderTest.GetByIDs(ctx, []string{"0", "3", "16"})

	assert.NoError(t, err)
}

func TestCRUD_GetByIDs_NoMatch(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectMany", ctx, mock.Anything,
		"SELECT * FROM `test_table` WHERE (`id` IN (?, ?)) ORDER BY `id` ASC",
		[]any{"invalid_id_1", "invalid_id_2"}).
		Return(nil)

	result, err := c.UnderTest.GetByIDs(ctx, []string{"invalid_id_1", "invalid_id_2"})

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestCRUD_Create(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("Exec", ctx,
		"INSERT INTO `test_table` (`id`, `name`) VALUES (?, ?)", []any{"12", "test12"}).
		Return(mocks.NewSQLResult(1, 1), nil)

	_, err := c.UnderTest.Create(ctx, &testObj{Id: "12", Name: "test12"})

	assert.NoError(t, err)
}

func TestCRUD_Create_Error(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("Exec", ctx,
		"INSERT INTO `test_table` (`id`, `name`) VALUES (?, ?)", []any{"12", "test12"}).
		Return(mocks.NewSQLResult(-1, -1), fmt.Errorf("insert error"))

	_, err := c.UnderTest.Create(ctx, &testObj{Id: "12", Name: "test12"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert error")
}

func TestCRUD_DeleteAll(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("Exec", ctx, "DELETE FROM test_table;", mock.Anything).
		Return(mocks.NewSQLResult(1, 1), nil)

	_, err := c.UnderTest.DeleteAll(ctx)

	assert.NoError(t, err)
}

func TestCRUD_Delete(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("Exec", ctx,
		"DELETE FROM `test_table` WHERE (`id` = ?)", []any{"test_id"}).
		Return(mocks.NewSQLResult(1, 1), nil)

	_, err := c.UnderTest.Delete(ctx, "test_id")

	assert.NoError(t, err)
}

func TestCRUD_Delete_NonExistent(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("Exec", ctx,
		"DELETE FROM `test_table` WHERE (`id` = ?)", []any{"non_existent_id"}).
		Return(mocks.NewSQLResult(-1, -1), fmt.Errorf("delete error"))

	_, err := c.UnderTest.Delete(ctx, "non_existent_id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestCRUD_Update(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("Exec", ctx,
		"UPDATE `test_table` SET `id`=?,`name`=? WHERE (`id` = ?)", []any{"12", "new name", "12"}).
		Return(mocks.NewSQLResult(1, 1), nil)

	_, err := c.UnderTest.Update(ctx, "12", &testObj{Id: "12", Name: "new name"})

	assert.NoError(t, err)
}

func TestCRUD_Update_NonExistent(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("Exec", ctx,
		"UPDATE `test_table` SET `id`=?,`name`=? WHERE (`id` = ?)",
		[]any{"non_existent_id", "new name", "non_existent_id"}).
		Return(mocks.NewSQLResult(-1, -1), fmt.Errorf("update error"))

	_, err := c.UnderTest.Update(ctx, "non_existent_id",
		&testObj{Id: "non_existent_id", Name: "new name"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

type testCaseData struct {
	ctx       context.Context
	ctxCancel func()

	Executor  *mocks.SQLExecutor
	UnderTest api.CRUD[testObj, string]
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
		UnderTest: crud.NewCRUD[testObj, string](tableInfo, executor, "sqlite3"),
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
	Id   string `db:"id"`
	Name string `db:"name"`
}
