package sqlcredo_test

import (
	"context"
	"fmt"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	_ "github.com/mattn/go-sqlite3"

	sc "gitlab.com/onrooh/sqlcredo"
)

type Identity string

type User struct {
	ID        Identity  `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  *string   `db:"last_name"`
	BirthDate time.Time `db:"birth_date"`
}

func (u *User) String() string {
	return fmt.Sprintf("%v", *u)
}

var (
	dsn       = ":memory:"
	driver    = "sqlite3"
	tableName = "user"
	idColumn  = "id"

	schema = `
CREATE TABLE IF NOT EXISTS user (
    id TEXT NOT NULL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NULL,
    birth_date DATETIME NOT NULL
);
`

	testUserValues = []User{
		{"u0", "John", ptr("Smith"), newTime("1989-03-05")},
		{"u1", "Carl", nil, newTime("1973-01-09")},
		{"u2", "Ann", ptr("Stone"), newTime("1985-08-01")},
	}

	testUserPtrs = wrapWithPtrs(testUserValues)
)

type UserRepo struct {
	sc.SQLCredo[User, Identity]
}

const CountByLastNameExistsQuery = `
SELECT 'with last_name' as category, COUNT(*) as cnt FROM user WHERE last_name IS NOT NULL
UNION
SELECT 'without last_name' as category, COUNT(*) as cnt FROM user WHERE last_name IS NULL;
`

type CountByLastNameExistsCategory struct {
	Name  string `db:"category"`
	Count int    `db:"cnt"`
}

func (r *UserRepo) CountByLastNameExists(ctx context.Context) (map[string]int, error) {
	var counters []CountByLastNameExistsCategory
	if err := r.SelectMany(ctx, &counters, CountByLastNameExistsQuery); err != nil {
		return nil, fmt.Errorf("failed to select entities: %w", err)
	}

	res := map[string]int{}
	for _, c := range counters {
		res[c.Name] = c.Count
	}

	return res, nil
}

var debugFunc = func(query string, args ...any) {
	g.GinkgoWriter.Printf("Query: [%s]; Args: %+v\n", query, args)
}

var _ = g.Describe("UserRepo", func() {
	var repo UserRepo
	ctx := context.Background()

	g.BeforeEach(func() {
		repo = UserRepo{
			SQLCredo: sc.NewSQLCredo[User, Identity](db, driver, tableName, idColumn).
				WithDebugFunc(debugFunc),
		}

		o.Expect(repo.InitSchema(ctx, schema)).NotTo(o.HaveOccurred())

		for _, u := range testUserPtrs {
			o.Expect(repo.Create(ctx, u)).NotTo(o.HaveOccurred())
		}

		cnt, err := repo.Count(ctx)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(int(cnt)).To(o.Equal(len(testUserPtrs)))
	})

	g.AfterEach(func() {
		repo.DeleteAll(ctx)
	})

	g.Context("base methods", func() {
		g.When("create user", func() {
			var user *User

			g.JustBeforeEach(func() {
				user = &User{"u4", "Gordon", ptr("Gibs"), newTime("1931-09-03")}
				o.Expect(repo.Create(ctx, user)).NotTo(o.HaveOccurred())
			})

			g.It("should be accessable by id", func() {
				got, err := repo.GetByID(ctx, user.ID)
				o.Expect(err).NotTo(o.HaveOccurred())
				o.Expect(got).To(o.Equal(*user))
			})
		})

		g.When("get all users", func() {
			g.It("should contain all record values", func() {
				o.Expect(repo.GetAll(ctx)).To(o.Equal(testUserValues))
			})

			g.When("get first page", func() {
				g.It("should contain first page records", func() {
					o.Expect(repo.GetPage(ctx, sc.WithOffset(0), sc.WithLimit(2), sc.WithOrderColumn("id"))).
						To(o.Equal(sc.Page[User]{
							Number:     0,
							Size:       2,
							Total:      3,
							TotalPages: 2,
							Content:    testUserValues[0:2],
						}))
				})
			})

			g.When("get second page", func() {
				g.It("should contain second page records", func() {
					o.Expect(repo.GetPage(ctx, sc.WithOffset(2), sc.WithLimit(2), sc.WithOrderColumn("id"))).
						To(o.Equal(sc.Page[User]{
							Number:     1,
							Size:       1,
							Total:      3,
							TotalPages: 2,
							Content:    testUserValues[2:],
						}))
				})
			})
		})

		g.When("get user by id", func() {
			g.It("should contain value of user", func() {
				o.Expect(repo.GetByID(ctx, testUserValues[2].ID)).
					To(o.Equal(testUserValues[2]))
			})
		})

		g.When("get users by ids", func() {
			g.It("should contain slice of values", func() {
				o.Expect(repo.GetByIDs(ctx, []Identity{testUserPtrs[1].ID, testUserPtrs[2].ID})).
					To(o.Equal([]User{testUserValues[1], testUserValues[2]}))
			})
		})

		g.When("delete user", func() {
			g.JustBeforeEach(func() {
				o.Expect(repo.Delete(ctx, testUserValues[1].ID)).NotTo(o.HaveOccurred())
			})

			g.It("should be absent in database", func() {
				got, err := repo.GetByID(ctx, testUserValues[1].ID)
				o.Expect(err).To(o.MatchError(sc.ErrRecordNotFound))
				o.Expect(got).To(o.Equal(User{}))
			})
		})

		g.When("update user", func() {
			var updated *User

			g.JustBeforeEach(func() {
				updated = testUserPtrs[1]
				updated.FirstName = updated.FirstName + "_updated"

				o.Expect(repo.Update(ctx, updated.ID, updated)).NotTo(o.HaveOccurred())
			})

			g.It("should be updated in database", func() {
				got, err := repo.GetByID(ctx, testUserValues[1].ID)
				o.Expect(err).NotTo(o.HaveOccurred())
				o.Expect(got).To(o.Equal(*updated))
			})
		})

		g.When("count users", func() {
			g.It("should contain actual number of users", func() {
				got, err := repo.Count(ctx)
				o.Expect(err).NotTo(o.HaveOccurred())
				o.Expect(int(got)).To(o.Equal(len(testUserPtrs)))
			})
		})
	})

	g.Context("custom methods", func() {
		g.It("count by last name exists", func() {
			o.Expect(repo.CountByLastNameExists(ctx)).
				To(o.Equal(map[string]int{
					"with last_name":    2,
					"without last_name": 1,
				}))
		})
	})
})

func newTime(input string) time.Time {
	result, err := time.Parse("2006-01-02", input)
	if err != nil {
		panic(err)
	}

	return result
}

func wrapWithPtrs[T comparable](input []T) []*T {
	result := make([]*T, 0, len(input))
	for _, i := range input {
		result = append(result, ptr[T](i))
	}
	return result
}

func ptr[T comparable](input T) *T {
	return &input
}
