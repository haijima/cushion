# cushion

[![GitHub Release](https://img.shields.io/github/v/release/haijima/cushion)](https://github.com/haijima/cushion/releases)
[![CI](https://github.com/haijima/cushion/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/haijima/cushion/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/haijima/cushion.svg)](https://pkg.go.dev/github.com/haijima/cushion)
[![Go report](https://goreportcard.com/badge/github.com/haijima/cushion)](https://goreportcard.com/report/github.com/haijima/cushion)

Cushion is a simple read-through cache library for Go.

## Installation

You can install cushion using the following command:

``` sh
go get github.com/haijima/cushion@latest
```

## How to use

Example of using `cushion` with `database/sql`.

```go
var db *sql.DB
var userCache = cushion.New(
    // fetch function
    func(ctx context.Context, id int64) (user User, err error) {
        err = db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = ?", id)
        return
    },
    // expiration
    5*time.Minute,
)

func initialize() {
	// When to clear cache
	userCache.Clear()
}

func getUser(ctx context.Context) {
    // Get the value from the cache if exists and not expired.
    // If not exists or expired, it fetches the value from Database.
    user, err := userCache.Get(ctx, 1)
    // ...
}
```
