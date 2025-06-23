package goquext_test

import (
	"testing"

	"github.com/Klojer/sqlcredo/internal/goquext"

	"github.com/stretchr/testify/assert"
)

func TestCreateDialectString(t *testing.T) {
	tests := []struct {
		name   string
		driver string
		want   string
	}{
		{name: "Postgres driver", driver: "pgx", want: "postgres"},
		{name: "SQLite driver", driver: "sqlite3", want: "sqlite3"},
		{name: "MySQL driver", driver: "mysql", want: "mysql"},
		{name: "Unknown driver", driver: "unknown", want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := goquext.CreateDialectString(tt.driver)

			assert.Equal(t, tt.want, got)
		})
	}
}
