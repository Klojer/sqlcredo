
# SQLCredo

[![GoDoc](https://godoc.org/github.com/Klojer/sqlcredo?status.svg)](https://godoc.org/github.com/Klojer/sqlcredo)
[![Go report](https://goreportcard.com/badge/github.com/Klojer/sqlcredo)](https://goreportcard.com/badge/github.com/Klojer/sqlcredo)

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
package main

import (
 "context"
 "database/sql"
 "fmt"

 _ "github.com/mattn/go-sqlite3"

 sc "github.com/Klojer/sqlcredo"
 scapi "github.com/Klojer/sqlcredo/pkg/api"
)

type User struct {
 ID   int    `db:"id"`
 Name string `db:"name"`
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL
);
`

func main() {
 ctx := context.Background()

 // Create a new SQLCredo instance
 db, _ := sql.Open("sqlite3", "test.db")
 repo := sc.NewSQLCredo[User, int](db, "sqlite3", "users", "id")

 _, err := repo.InitSchema(ctx, schema)
 orPanic(err)

 // Create
 user := User{ID: 1, Name: "John"}
 _, err = repo.Create(ctx, &user)
 orPanic(err)

 // Read
 users, err := repo.GetAll(ctx)
 orPanic(err)
 fmt.Printf("Users: %+v\n", users)

 // Read with pagination
 page, err := repo.GetPage(ctx,
  scapi.WithPageNumber(0),
  scapi.WithPageSize(10),
  scapi.WithSortBy("name"))
 orPanic(err)
 fmt.Printf("Page: %+v\n", page)

 // Update
 user.Name = "Johnnnn"
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

See example of repository with custom query: [examples/users/users.go](https://github.com/Klojer/sqlcredo/blob/main/examples/users/users.go)

## Debug Support

Enable SQL query debugging:

```go
repo.WithDebugFunc(func(sql string, args ...any) {
    log.Printf("SQL: %s Args: %v", sql, args)
})
```
