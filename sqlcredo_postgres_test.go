package sqlcredo_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	pgDriver = "pgx"
	pgSchema = `
CREATE TABLE IF NOT EXISTS "users" (
    id TEXT NOT NULL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NULL,
    birth_date TIMESTAMP NOT NULL
);
`
)

func TestPostgres(t *testing.T) {
	db := createDB(t)
	defer func() { require.NoError(t, db.Close()) }()

	params := TestCaseParams{
		Schema: pgSchema,
		Driver: pgDriver,
		DB:     db,
	}

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
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			tC.run(t, params)
		})
	}
}

func createDB(t *testing.T) *sql.DB {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:15.3-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open(pgDriver, connStr)
	require.NoError(t, err)

	return db
}
