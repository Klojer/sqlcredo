package goquext

import (
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

func CreateDialectString(driver string) string {
	if driver == "pgx" {
		return "postgres"
	}
	return driver
}
