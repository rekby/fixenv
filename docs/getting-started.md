# Getting started

This guide walks through the basics of using Fixenv to share fixtures across Go tests while keeping resource lifecycles explicit.

## Installation

Fixenv is distributed as a Go module:

```bash
go get github.com/rekby/fixenv
```

Run the command again to upgrade when a new version is released.

For an instant start without writing custom fixtures, import [`github.com/rekby/fixenv/sf`](../sf) for prebuilt helpers such as cancellable contexts, temporary directories, HTTP servers, and TCP listeners. The examples below show how to extend these basics with your own project logic.

### First steps with bundled fixtures

The quickest way to experience Fixenv is to reuse the built-in fixtures. The snippet below shows how to obtain an automatically cancelled context that respects test scopes:

```go
package context_test

import (
    "testing"

    "github.com/rekby/fixenv"
    "github.com/rekby/fixenv/sf"
)

func TestAutoCanceledContext(t *testing.T) {
    t.Parallel()

    e := fixenv.New(t)
    ctx := sf.Context(e)

    // Use ctx in your test. When the test scope finishes, the context is
    // automatically cancelled by the fixture cleanup.
    select {
    case <-ctx.Done():
        t.Fatal("context cancelled too early")
    default:
    }
}
```

Each call to `sf.Context` is cached per test by default, so subtests or parallel invocations receive isolated cancellable contexts without extra setup.

### Your first custom fixture

Create a test file and import Fixenv:

```go
package counter_test

import (
    "testing"

    "github.com/rekby/fixenv"
)

var globalCounter int

func counter(e fixenv.Env) int {
    return fixenv.CacheResult(e, func() (*fixenv.GenericResult[int], error) {
        globalCounter++
        return fixenv.NewGenericResult(globalCounter), nil
    })
}

func TestCounter(t *testing.T) {
    e := fixenv.New(t)

    if counter(e) != counter(e) {
        t.Fatal("value must be cached within a single test scope")
    }
}
```

Key ideas:

1. Call `fixenv.New(t)` to create an environment bound to the test.
2. Wrap setup logic in `fixenv.CacheResult` and return a `*fixenv.GenericResult[T]`.
3. Store the value in the result; Fixenv caches the result and returns it from subsequent calls.

## Adding cleanup

When a fixture allocates resources, attach a cleanup callback:

```go
// requires import "os"
func tempDir(e fixenv.Env) string {
    return fixenv.CacheResult(e, func() (*fixenv.GenericResult[string], error) {
        dir, err := os.MkdirTemp("", "fixenv-example-")
        if err != nil {
            return nil, err
        }
        cleanup := func() { os.RemoveAll(dir) }
        return fixenv.NewGenericResultWithCleanup(dir, cleanup), nil
    })
}
```

Cleanups run in **last-in, first-out** order when the owning scope finishes, even if the test fails or is skipped.

## Choosing a scope

Fixenv defaults to `ScopeTest`, meaning a fixture is recomputed for every `testing.T`. Subtests therefore start with a fresh cache. Override the scope by passing `CacheOptions`:

```go
// requires import "database/sql"
func sharedDatabase(e fixenv.Env) *sql.DB {
    return fixenv.CacheResult(e, func() (*fixenv.GenericResult[*sql.DB], error) {
        db := openDatabase()
        cleanup := func() { db.Close() }
        return fixenv.NewGenericResultWithCleanup(db, cleanup), nil
    }, fixenv.CacheOptions{Scope: fixenv.ScopePackage})
}
```

With `ScopePackage`, the database is created once per package. Use `ScopeTestAndSubtests` to reuse a fixture across a parent test and all of its subtests.

Read more in [Scopes and lifetimes](scopes-and-lifetimes.md).

## Running the examples

The repository ships with runnable examples. Execute them with:

```bash
go test ./examples/...
```

Then follow the [example walkthroughs](examples/README.md) for context and expected output.
