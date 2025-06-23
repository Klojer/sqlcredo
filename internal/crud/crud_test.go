package crud

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/Klojer/sqlcredo/internal/table"
	"github.com/Klojer/sqlcredo/pkg/api"

	"github.com/stretchr/testify/assert"
)

func TestCRUD_GetAll(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.GetAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, c.Executor.queries[0], "SELECT * FROM `test_table`")
}

func TestCRUD_GetByID(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.GetByID(ctx, "test_id")
	assert.NoError(t, err)
	assert.Equal(t, c.Executor.queries[0], "SELECT * FROM `test_table` WHERE (`id` = ?)")
	assert.Equal(t, c.Executor.args[0], []any{"test_id"})
}

func TestCRUD_GetByIDs(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.GetByIDs(ctx, []string{"0", "3", "16"})
	assert.NoError(t, err)
	assert.Equal(t, c.Executor.queries[0], "SELECT * FROM `test_table` WHERE (`id` IN (?, ?, ?)) ORDER BY `id` ASC")
	assert.Equal(t, c.Executor.args[0], []any{"0", "3", "16"})
}

func TestCRUD_Create(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.Create(ctx, &testObj{Id: "12", Name: "test12"})
	assert.NoError(t, err)
	assert.Contains(t, c.Executor.queries[0], "INSERT INTO `test_table` (`id`, `name`) VALUES (?, ?)")
	assert.Equal(t, c.Executor.args[0], []any{"12", "test12"})
}

func TestCRUD_DeleteAll(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.DeleteAll(ctx)
	assert.NoError(t, err)
	assert.Contains(t, c.Executor.queries[0], "DELETE FROM test_table;")
	assert.Nil(t, c.Executor.args[0])
}

func TestCRUD_Delete(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.Delete(ctx, "test_id")
	assert.NoError(t, err)
	assert.Contains(t, c.Executor.queries[0], "DELETE FROM `test_table` WHERE (`id` = ?)")
	assert.Equal(t, c.Executor.args[0], []any{"test_id"})
}

func TestCRUD_Update(t *testing.T) {
	c, ctx := newTestCase(t)

	_, err := c.UnderTest.Update(ctx, "12", &testObj{Id: "12", Name: "test12"})
	assert.NoError(t, err)
	assert.Contains(t, c.Executor.queries[0], "UPDATE `test_table` SET `id`=?,`name`=? WHERE (`id` = ?)")
	assert.Equal(t, c.Executor.args[0], []any{"12", "test12", "12"})
}

type testCaseData struct {
	ctx       context.Context
	ctxCancel func()

	Executor  *mockExecutor
	UnderTest api.CRUD[testObj, string]
}

func newTestCase(t *testing.T) (*testCaseData, context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	executor := &mockExecutor{}
	tableInfo := table.Info{Name: "test_table", IDColumn: "id"}

	c := &testCaseData{
		ctx:       ctx,
		ctxCancel: cancel,

		Executor:  executor,
		UnderTest: NewCRUD[testObj, string](tableInfo, executor, "sqlite3"),
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

type mockExecutor struct {
	queries []string
	args    [][]any
}

func (m *mockExecutor) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	m.queries = append(m.queries, query)
	m.args = append(m.args, args)
	return nil, nil
}

func (m *mockExecutor) SelectOne(ctx context.Context, dest any, query string, args ...any) error {
	m.queries = append(m.queries, query)
	m.args = append(m.args, args)
	return nil
}

func (m *mockExecutor) SelectMany(ctx context.Context, dest any, query string, args ...any) error {
	m.queries = append(m.queries, query)
	m.args = append(m.args, args)
	return nil
}

func (m *mockExecutor) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, errors.New("not implemented")
}
