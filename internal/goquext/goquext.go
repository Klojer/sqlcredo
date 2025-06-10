package goquext

import (
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

func CreateDialect(driver string) goqu.DialectWrapper {
	return goqu.Dialect(createDialectString(driver))
}

func createDialectString(driver string) string {
	if driver == "pgx" {
		return "postgres"
	}
	return driver
}
