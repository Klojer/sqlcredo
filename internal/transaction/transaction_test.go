package transaction_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Klojer/sqlcredo/internal/table"
	"github.com/Klojer/sqlcredo/internal/transaction"
	"github.com/Klojer/sqlcredo/pkg/api"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestTransactionCommit(t *testing.T) {
	c, ctx := newTestCase(t, func(mock sqlmock.Sqlmock) {
		query := "INSERT INTO `test_table` \\(`id`, `name`\\) VALUES \\(\\?, \\?\\)"
		mock.ExpectBegin()
		mock.ExpectExec(query).
			WithArgs("#1", "u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
	})

	_, err := c.UnderTest.Create(ctx, &testObj{
		ID: "#1", Name: "u1",
	})
	assert.NoError(t, err)

	err = c.UnderTest.Commit()
	assert.NoError(t, err)
}

func TestTransactionRollback(t *testing.T) {
	c, ctx := newTestCase(t, func(mock sqlmock.Sqlmock) {
		query := "INSERT INTO `test_table` \\(`id`, `name`\\) VALUES \\(\\?, \\?\\)"
		mock.ExpectBegin()
		mock.ExpectExec(query).
			WithArgs("#1", "u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectRollback()
	})

	_, err := c.UnderTest.Create(ctx, &testObj{
		ID: "#1", Name: "u1",
	})
	assert.NoError(t, err)

	err = c.UnderTest.Rollback()
	assert.NoError(t, err)
}

type testCaseData struct {
	ctx       context.Context
	ctxCancel func()

	DB        *sql.DB
	UnderTest api.Transaction[testObj, string]
}

func newTestCase(t *testing.T, mockCfg func(sqlmock.Sqlmock)) (*testCaseData, context.Context) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	driver := "sqlite3"

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	mockCfg(mock)

	dbx := sqlx.NewDb(db, driver)

	tableInfo := table.Info{Name: "test_table", IDColumn: "id"}

	debugFunc := func(sql string, args ...any) {
		t.Helper()
		t.Logf("Query: '%s'; args: '%v'", sql, args)
	}

	tx, err := transaction.NewTx[testObj, string](ctx, dbx, tableInfo, driver, debugFunc, nil)
	assert.NoError(t, err)

	c := &testCaseData{
		ctx:       ctx,
		ctxCancel: cancel,

		DB:        db,
		UnderTest: tx,
	}

	t.Cleanup(func() {
		c.TearDown(t)
	})

	return c, ctx
}

func (c *testCaseData) TearDown(t *testing.T) {
	t.Helper()

	_ = c.DB.Close()
	c.ctxCancel()
}

type testObj struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}
