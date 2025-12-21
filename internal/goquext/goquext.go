// Package goquext provides utilities for working with goqu SQL query builder.
// It handles dialect string creation and normalizes database driver names
// to their corresponding goqu dialect identifiers.
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
