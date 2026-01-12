package sqlexec_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Klojer/sqlcredo/internal/sqlexec"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSQLExecutor_SelectOne(t *testing.T) {
	c, ctx := newTestCase(t)

	query := "SELECT name FROM users WHERE id = ?"
	rows := sqlmock.NewRows([]string{"name"}).AddRow("John Doe")
	c.Mock.ExpectQuery(query).WithArgs(1).WillReturnRows(rows)

	var name string
	err := c.UnderTest.SelectOne(ctx, &name, query, 1)
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", name)
}

func TestSQLExecutor_SelectMany(t *testing.T) {
	c, ctx := newTestCase(t)

	query := "SELECT name FROM users"
	rows := c.Mock.NewRows([]string{"name"}).AddRow("John Doe").AddRow("Jane Doe")
	c.Mock.ExpectQuery(query).WillReturnRows(rows)

	var names []string
	err := c.UnderTest.SelectMany(ctx, &names, query)
	assert.NoError(t, err)
	assert.Equal(t, []string{"John Doe", "Jane Doe"}, names)
}

func TestSQLExecutor_Exec(t *testing.T) {
	c, ctx := newTestCase(t)

	query := "INSERT INTO users (name) VALUES (?)"
	c.Mock.ExpectExec("INSERT INTO users \\(name\\) VALUES \\(\\?\\)").
		WithArgs("John Doe").WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := c.UnderTest.Exec(ctx, query, "John Doe")
	assert.NoError(t, err)
}

func TestSQLExecutor_SelectOne_Error(t *testing.T) {
	c, ctx := newTestCase(t)

	query := "SELECT name FROM users WHERE id = ?"
	c.Mock.ExpectQuery(query).WithArgs(1).
		WillReturnError(fmt.Errorf("query error"))

	var name string
	err := c.UnderTest.SelectOne(ctx, &name, query, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to get data from db")
}

func TestSQLExecutor_SelectMany_Error(t *testing.T) {
	c, ctx := newTestCase(t)

	query := "SELECT name FROM users"
	c.Mock.ExpectQuery(query).WillReturnError(fmt.Errorf("query error"))

	var names []string
	err := c.UnderTest.SelectMany(ctx, &names, query)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to select data from db")
}

func TestSQLExecutor_Exec_Error(t *testing.T) {
	c, ctx := newTestCase(t)

	query := "INSERT INTO users (name) VALUES (?)"
	c.Mock.ExpectExec(query).WithArgs("John Doe").
		WillReturnError(fmt.Errorf("execution error"))

	_, err := c.UnderTest.Exec(ctx, query, "John Doe")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to exec db query")
}

type testCaseData struct {
	ctx       context.Context
	ctxCancel func()
	Mock      sqlmock.Sqlmock
	UnderTest *sqlexec.SQLExecutor
}

func newTestCase(t *testing.T) (*testCaseData, context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	executor := sqlexec.NewSQLExecutor(sqlxDB)

	c := &testCaseData{
		ctx:       ctx,
		ctxCancel: cancel,
		Mock:      mock,
		UnderTest: executor,
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return c, ctx
}

func (c *testCaseData) TearDown(t *testing.T) {
	c.ctxCancel()
}

// mockSQLExecutor is a mock implementation of domain.SQLExecutor for testing delegatingExecutor
type mockSQLExecutor struct {
	mock.Mock
}

func (m *mockSQLExecutor) SelectOne(ctx context.Context, dest any, query string, args ...any) error {
	argsMock := m.Called(ctx, dest, query, args)
	return argsMock.Error(0)
}

func (m *mockSQLExecutor) SelectMany(ctx context.Context, dest any, query string, args ...any) error {
	argsMock := m.Called(ctx, dest, query, args)
	return argsMock.Error(0)
}

func (m *mockSQLExecutor) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	argsMock := m.Called(ctx, query, args)
	result := argsMock.Get(0)
	if result == nil {
		return nil, argsMock.Error(1)
	}
	return result.(sql.Result), argsMock.Error(1)
}

func TestNewDelegatingExecutor_SelectOne(t *testing.T) {
	m := &mockSQLExecutor{}
	m.On("SelectOne", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	debugFunc := func(sql string, args ...any) {}

	executor := sqlexec.NewDelegatingExecutor(m, debugFunc)
	assert.NotNil(t, executor)

	ctx := context.Background()
	dest := "test"
	query := "SELECT * FROM test"
	args := []any{1, "arg"}

	err := executor.SelectOne(ctx, dest, query, args...)

	assert.NoError(t, err)
	m.AssertCalled(t, "SelectOne", ctx, dest, query, args)
}

func TestNewDelegatingExecutor_SelectMany(t *testing.T) {
	m := &mockSQLExecutor{}
	m.On("SelectMany", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	debugFunc := func(sql string, args ...any) {}

	executor := sqlexec.NewDelegatingExecutor(m, debugFunc)
	assert.NotNil(t, executor)

	ctx := context.Background()
	dest := []string{"test"}
	query := "SELECT * FROM test"
	args := []any{1, "arg"}

	err := executor.SelectMany(ctx, dest, query, args...)

	assert.NoError(t, err)
	m.AssertCalled(t, "SelectMany", ctx, dest, query, args)
}

func TestNewDelegatingExecutor_Exec(t *testing.T) {
	m := &mockSQLExecutor{}
	m.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(sqlmock.NewResult(1, 1), nil)

	debugFunc := func(sql string, args ...any) {}

	executor := sqlexec.NewDelegatingExecutor(m, debugFunc)
	assert.NotNil(t, executor)

	ctx := context.Background()
	query := "INSERT INTO test VALUES (?)"
	args := []any{"value"}

	result, err := executor.Exec(ctx, query, args...)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	m.AssertCalled(t, "Exec", ctx, query, args)
}

func TestNewDelegatingExecutor_SelectOne_Error(t *testing.T) {
	expectedErr := fmt.Errorf("select one error")
	m := &mockSQLExecutor{}
	m.On("SelectOne", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedErr)

	debugFunc := func(sql string, args ...any) {}

	executor := sqlexec.NewDelegatingExecutor(m, debugFunc)

	ctx := context.Background()
	dest := "test"
	query := "SELECT * FROM test"

	err := executor.SelectOne(ctx, dest, query)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to get data from db")
	m.AssertExpectations(t)
}

func TestNewDelegatingExecutor_SelectMany_Error(t *testing.T) {
	expectedErr := fmt.Errorf("select many error")
	m := &mockSQLExecutor{}
	m.On("SelectMany", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedErr)

	debugFunc := func(sql string, args ...any) {}

	executor := sqlexec.NewDelegatingExecutor(m, debugFunc)

	ctx := context.Background()
	dest := []string{"test"}
	query := "SELECT * FROM test"

	err := executor.SelectMany(ctx, dest, query)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to select data from db")
	m.AssertExpectations(t)
}

func TestNewDelegatingExecutor_Exec_Error(t *testing.T) {
	expectedErr := fmt.Errorf("exec error")
	m := &mockSQLExecutor{}
	m.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(nil, expectedErr)

	debugFunc := func(sql string, args ...any) {}

	executor := sqlexec.NewDelegatingExecutor(m, debugFunc)

	ctx := context.Background()
	query := "INSERT INTO test VALUES (?)"

	_, err := executor.Exec(ctx, query)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to exec db query")
	m.AssertExpectations(t)
}

func TestNewDelegatingExecutor_DebugFunc(t *testing.T) {
	m := &mockSQLExecutor{}
	m.On("SelectOne", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	var calledWith []any
	debugFunc := func(sql string, args ...any) {
		calledWith = append([]any{sql}, args...)
	}

	executor := sqlexec.NewDelegatingExecutor(m, debugFunc)
	assert.NotNil(t, executor)

	ctx := context.Background()
	query := "SELECT * FROM test"
	args := []any{1}

	_ = executor.SelectOne(ctx, "dest", query, args...)

	// DebugFunc should be called with the query and args
	assert.Equal(t, query, calledWith[0])
	assert.Equal(t, args[0], calledWith[1])
}
