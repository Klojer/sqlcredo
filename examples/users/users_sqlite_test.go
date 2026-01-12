package users_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/require"
)

const (
	sqliteDriver = "sqlite3"
	sqliteDSN    = ":memory:?_foreign_keys=true"
	sqliteSchema = `
CREATE TABLE IF NOT EXISTS users (
    id TEXT NOT NULL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NULL,
    birth_date DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS articles (
    id TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    user_id TEXT NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id)
);
`
)

func TestSqlite(t *testing.T) {
	db, err := sql.Open(sqliteDriver, sqliteDSN)
	require.NoError(t, err)
	defer func() { require.NoError(t, db.Close()) }()

	testCases := []TestCaseDesc{
		{name: "create-user", run: CaseCreateUser},
		{name: "get-all-users", run: CaseGetAllUsers},
		{name: "get-user-by-id", run: CaseGetUserByID},
		{name: "get-users-by-ids", run: CaseGetUsersByIDs},
		{name: "delete-user", run: CaseDeleteUser},
		{name: "update-user", run: CaseUpdateUser},
		{name: "validate-page-request", run: CaseValidatePageRequest},
		{name: "get-page", run: CaseGetPage},
		{name: "get-page-custom-order", run: CaseGetPageCustomOrder},
		{name: "count-users", run: CaseCountUsers},
		{name: "count-by-last-name-exists", run: CaseCountByLastNameExists},
		{name: "count-by-last-name-exists-ctx-err", run: CaseCountByLastNameExistsCtxError},
		{name: "tx-commit", run: CaseTxCommit},
		{name: "tx-rollback", run: CaseTxRollback},
		{name: "get-user-with-articles", run: CaseGetUserWithArticles},
		{name: "delete-user-with-articles", run: CaseDeleteUserWithArticles},
		{name: "create-user-with-article-positive", run: CaseCreateUserWithArticlePositive},
		{name: "create-user-with-article-negative", run: CaseCreateUserWithArticleNegative},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			caze, ctx := NewTestCase(t, TestCaseParams{
				Schema: sqliteSchema,
				Driver: sqliteDriver,
				DB:     db,
			})

			tC.run(t, ctx, caze)

			_, err := caze.UnderTest.Articles().DeleteAll(ctx)
			require.NoError(t, err)
			_, err = caze.UnderTest.DeleteAll(ctx)
			require.NoError(t, err)
			caze.CtxCancel()
		})
	}
}
