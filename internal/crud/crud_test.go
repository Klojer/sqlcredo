package crud_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Klojer/sqlcredo/internal/crud"
	"github.com/Klojer/sqlcredo/internal/domain"
	"github.com/Klojer/sqlcredo/internal/mocks"
	"github.com/Klojer/sqlcredo/internal/table"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCRUD_GetAll(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectMany", ctx, mock.Anything, "SELECT * FROM `test_table`", mock.Anything).
		Return(nil)

	got := make([]testObj, 0)
	err := c.UnderTest.GetAll(ctx, &got)

	assert.NoError(t, err)
}

func TestCRUD_GetAll_DatabaseError(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectMany", ctx, mock.Anything, "SELECT * FROM `test_table`", mock.Anything).
		Return(fmt.Errorf("database error"))

	got := make([]testObj, 0)
	err := c.UnderTest.GetAll(ctx, &got)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load records")
}

func TestCRUD_GetByID(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectOne", ctx, mock.Anything,
		"SELECT * FROM `test_table` WHERE (`id` = ?)", []any{"test_id"}).
		Return(nil)

	var got testObj
	err := c.UnderTest.GetByID(ctx, &got, "test_id")

	assert.NoError(t, err)
}

func TestCRUD_GetByID_DatabaseError(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectOne", ctx, mock.Anything,
		"SELECT * FROM `test_table` WHERE (`id` = ?)", []any{"non_existent_id"}).
		Return(fmt.Errorf("database error"))

	var got testObj
	err := c.UnderTest.GetByID(ctx, &got, "non_existent_id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to select record")
}

func TestCRUD_GetByIDs(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectMany", ctx, mock.Anything,
		"SELECT * FROM `test_table` WHERE (`id` IN (?, ?, ?)) ORDER BY `id` ASC",
		[]any{"0", "3", "16"}).
		Return(nil)

	got := make([]testObj, 0)
	err := c.UnderTest.GetByIDs(ctx, &got, []string{"0", "3", "16"})

	assert.NoError(t, err)
}

func TestCRUD_GetByIDs_NoMatch(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectMany", ctx, mock.Anything,
		"SELECT * FROM `test_table` WHERE (`id` IN (?, ?)) ORDER BY `id` ASC",
		[]any{"invalid_id_1", "invalid_id_2"}).
		Return(nil)

	got := make([]testObj, 0)
	err := c.UnderTest.GetByIDs(ctx, &got, []string{"invalid_id_1", "invalid_id_2"})

	assert.NoError(t, err)
	assert.Empty(t, got)
}

func TestCRUD_GetByIDs_DatabaseError(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("SelectMany", ctx, mock.Anything,
		"SELECT * FROM `test_table` WHERE (`id` IN (?, ?, ?)) ORDER BY `id` ASC",
		[]any{"0", "3", "16"}).
		Return(fmt.Errorf("select error"))

	got := make([]testObj, 0)
	err := c.UnderTest.GetByIDs(ctx, &got, []string{"0", "3", "16"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "select error")
}

func TestCRUD_Create(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("Exec", ctx,
		"INSERT INTO `test_table` (`id`, `name`) VALUES (?, ?)", []any{"12", "test12"}).
		Return(mocks.NewSQLResult(1, 1), nil)

	_, err := c.UnderTest.Create(ctx, &testObj{ID: "12", Name: "test12"})

	assert.NoError(t, err)
}

func TestCRUD_Create_DatabaseError(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("Exec", ctx,
		"INSERT INTO `test_table` (`id`, `name`) VALUES (?, ?)",
		[]any{"12", "test12"}).
		Return(mocks.NewSQLResult(-1, -1), fmt.Errorf("insert error"))

	_, err := c.UnderTest.Create(ctx, &testObj{ID: "12", Name: "test12"})

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

func TestCRUD_DeleteAll_DatabaseError(t *testing.T) {
	c, ctx := newTestCase(t)

	c.Executor.On("Exec", ctx, "DELETE FROM test_table;", mock.Anything).
		Return(mocks.NewSQLResult(-1, -1), fmt.Errorf("database error"))

	_, err := c.UnderTest.DeleteAll(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestCRUD_Delete(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("Exec", ctx,
		"DELETE FROM `test_table` WHERE (`id` = ?)", []any{"test_id"}).
		Return(mocks.NewSQLResult(1, 1), nil)

	_, err := c.UnderTest.Delete(ctx, "test_id")

	assert.NoError(t, err)
}

func TestCRUD_Delete_NonExistentID(t *testing.T) {
	c, ctx := newTestCase(t)

	c.Executor.On("Exec", ctx,
		"DELETE FROM `test_table` WHERE (`id` = ?)",
		[]any{"non_existent_id"}).
		Return(mocks.NewSQLResult(-1, -1), fmt.Errorf("delete error"))

	_, err := c.UnderTest.Delete(ctx, "non_existent_id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
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
		"UPDATE `test_table` SET `id`=?,`name`=? WHERE (`id` = ?)",
		[]any{"12", "new name", "12"}).
		Return(mocks.NewSQLResult(1, 1), nil)

	_, err := c.UnderTest.Update(ctx, "12", &testObj{ID: "12", Name: "new name"})

	assert.NoError(t, err)
}

func TestCRUD_Update_NonExistentID(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("Exec", ctx,
		"UPDATE `test_table` SET `id`=?,`name`=? WHERE (`id` = ?)",
		[]any{"non_existent_id", "updated_name", "non_existent_id"}).
		Return(mocks.NewSQLResult(-1, -1), fmt.Errorf("update error"))

	updateObj := &testObj{ID: "non_existent_id", Name: "updated_name"}
	_, err := c.UnderTest.Update(ctx, "non_existent_id", updateObj)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestCRUD_Update_NonExistent(t *testing.T) {
	c, ctx := newTestCase(t)
	c.Executor.On("Exec", ctx,
		"UPDATE `test_table` SET `id`=?,`name`=? WHERE (`id` = ?)",
		[]any{"non_existent_id", "new name", "non_existent_id"}).
		Return(mocks.NewSQLResult(-1, -1), fmt.Errorf("update error"))

	_, err := c.UnderTest.Update(ctx, "non_existent_id",
		&testObj{ID: "non_existent_id", Name: "new name"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

type testCaseData struct {
	ctx       context.Context
	ctxCancel func()

	Executor  *mocks.SQLExecutor
	UnderTest domain.CRUD[testObj, string]
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
	ID   string `db:"id"`
	Name string `db:"name"`
}
