package sqlcredo_test

import (
	"database/sql"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var db *sql.DB

func TestGolangSqlcredo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GolangSqlcredo Suite")
}

var _ = BeforeSuite(func() {
	var err error
	db, err = sql.Open(driver, dsn)
	DeferCleanup(db.Close)

	Expect(err).NotTo(HaveOccurred())
})
