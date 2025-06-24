package sqlexec_test

import (
	"context"
	"testing"
	"time"

	"github.com/Klojer/sqlcredo/internal/sqlexec"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
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

func TestSQLExecutor_BeginTx(t *testing.T) {
	c, ctx := newTestCase(t)

	c.Mock.ExpectBegin()

	tx, err := c.UnderTest.BeginTx(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
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
