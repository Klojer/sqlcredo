
# SQLCredo

SQLCredo is a type-safe generic SQL CRUD operations wrapper for Go, built on top of [sqlx](https://github.com/jmoiron/sqlx) and [goqu](https://github.com/doug-martin/goqu).

The idea of the package to provide basic CRUD and pagination operations out of the box and simplify adding custom raw SQL-queries to extend functionality.

## Features

- Generic type-safe CRUD operations
- Built-in pagination support
- SQL query debugging capabilities
- Support for multiple SQL drivers (tested on sqlite3 and postgres (pgx))
- Transaction support
- Prepared statements by default

## Installation

```go
go get github.com/Klojer/sqlcredo
```

## Quick Start

```go
import (
 "context"
 "database/sql"

 "github/Klojer/sqlcredo"
 "github/Klojer/sqlcredo/pkg/model"
)

type User struct {
 ID   int    `db:"id"`
 Name string `db:"name"`
}

func main() {
 ctx := context.Background()

 // Create a new SQLCredo instance
 db, _ := sql.Open("sqlite3", "test.db")
 repo := sqlcredo.NewSQLCredo[User, int](db, "sqlite3", "users", "id")

 // Create
 user := User{ID: 1, Name: "John"}
 _, err := repo.Create(ctx, &user)
 orPanic(err)

 // Read
 users, err := repo.GetAll(ctx)
 orPanic(err)

 // Read with pagination
 page, err := repo.GetPage(ctx,
  model.WithPageNumber(0),
  model.WithPageSize(10),
  model.WithSortBy("name"))
 orPanic(err)

 // Update
 _, err = repo.Update(ctx, 1, &user)
 orPanic(err)

 // Delete
 _, err = repo.Delete(ctx, 1)
 orPanic(err)
}

func orPanic(err error) {
 if err == nil {
  return
 }
 panic(err)
}
```

See example of repository with custom query: [examples/users/users.go](https://github.com/Klojer/sqlcredo/blob/546f50239e1f399e8559534a8e3d02c748a89b09/examples/users/users.go)

## Debug Support

Enable SQL query debugging:

```go
repo.WithDebugFunc(func(sql string, args ...any) {
    log.Printf("SQL: %s Args: %v", sql, args)
})
```
