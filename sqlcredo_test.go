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
		{"u3", "Ann", ptr("Brick"), newTime("1987-03-02")},
		{"u4", "Antony", nil, newTime("1987-03-02")},
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

		_, err := repo.InitSchema(ctx, schema)
		o.Expect(err).NotTo(o.HaveOccurred())

		for _, u := range testUserPtrs {
			_, err := repo.Create(ctx, u)
			o.Expect(err).NotTo(o.HaveOccurred())
		}

		cnt, err := repo.Count(ctx)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(int(cnt)).To(o.Equal(len(testUserPtrs)))
	})

	g.AfterEach(func() {
		if _, err := repo.DeleteAll(ctx); err != nil {
			g.GinkgoLogr.Error(err, "unable to clear after case")
		}
	})

	g.Context("base methods", func() {
		g.When("create user", func() {
			var user *User

			g.JustBeforeEach(func() {
				user = &User{"u99", "Gordon", ptr("Gibs"), newTime("1931-09-03")}
				_, err := repo.Create(ctx, user)
				o.Expect(err).NotTo(o.HaveOccurred())
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
		})

		g.When("validate page request", func() {
			g.It("page size can't be 0", func() {
				_, err := repo.GetPage(ctx, sc.WithPageSize(0))
				o.Expect(err).To(o.MatchError(sc.ErrInvalidPageSize))
			})
		})

		g.When("get page of users", func() {
			g.When("get first page", func() {
				g.It("should contain first page records", func() {
					o.Expect(repo.GetPage(ctx, sc.WithPageNumber(0), sc.WithPageSize(2), sc.WithSort("id"))).
						To(o.Equal(sc.Page[User]{
							Number:     0,
							Size:       2,
							Total:      5,
							TotalPages: 3,
							Content:    testUserValues[0:2],
						}))
				})
			})

			g.When("get second page", func() {
				g.It("should contain second page records", func() {
					o.Expect(repo.GetPage(ctx, sc.WithPageNumber(1), sc.WithPageSize(2), sc.WithSort("id"))).
						To(o.Equal(sc.Page[User]{
							Number:     1,
							Size:       2,
							Total:      5,
							TotalPages: 3,
							Content:    testUserValues[2:4],
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
				_, err := repo.Delete(ctx, testUserValues[1].ID)
				o.Expect(err).NotTo(o.HaveOccurred())
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

				_, err := repo.Update(ctx, updated.ID, updated)
				o.Expect(err).NotTo(o.HaveOccurred())
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
					"with last_name":    3,
					"without last_name": 2,
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
		result = append(result, ptr(i))
	}
	return result
}

func ptr[T comparable](input T) *T {
	return &input
}
