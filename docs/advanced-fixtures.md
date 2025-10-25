# Advanced fixtures

This document explores patterns for building richer fixture ecosystems on top of Fixenv.

## Composing fixtures

Fixtures can depend on other fixtures simply by calling their helpers. Cached results are reused automatically, so downstream fixtures do not trigger redundant setup.

```go
// requires imports "database/sql" and "os"
func db(e fixenv.Env) *sql.DB {
    return fixenv.CacheResult(e, func() (*fixenv.GenericResult[*sql.DB], error) {
        dsn := os.Getenv("TEST_DATABASE_DSN")
        database := connect(dsn)
        cleanup := func() { database.Close() }
        return fixenv.NewGenericResultWithCleanup(database, cleanup), nil
    }, fixenv.CacheOptions{Scope: fixenv.ScopePackage})
}

func seededAccount(e fixenv.Env, name string) Account {
    return fixenv.CacheResult(e, func() (*fixenv.GenericResult[Account], error) {
        account := createAccount(db(e), name)
        cleanup := func() { deleteAccount(db(e), account.ID) }
        return fixenv.NewGenericResultWithCleanup(account, cleanup), nil
    }, fixenv.CacheOptions{CacheKey: name})
}
```

`seededAccount` reuses the cached `db` connection and ensures teardown is executed exactly once per account per scope.

## Using `ErrSkipTest`

Fixtures can abort the current test while signalling that the result should be skipped for the remainder of the scope:

```go
// requires import "os"
func optionalService(e fixenv.Env) Service {
    value := fixenv.CacheResult(e, func() (*fixenv.GenericResult[Service], error) {
        if os.Getenv("SERVICE_ENDPOINT") == "" {
            return nil, fixenv.ErrSkipTest
        }

        svc := connectService()
        cleanup := func() { svc.Close() }
        return fixenv.NewGenericResultWithCleanup(svc, cleanup), nil
    })

    if value == nil {
        e.T().Skip("service not configured")
    }

    return value
}
```

If the environment variable is missing, the fixture returns `ErrSkipTest`. Fixenv caches the skip decision and prevents future calls to the fixture within the scope.

## Building custom environments

The `EnvT` struct implements all Fixenv behaviour. Embed it in your own type to expose domain-specific helpers while preserving compatibility with existing fixtures.

```go
// requires import "github.com/rekby/fixenv"
type ProjectEnv struct {
    *fixenv.EnvT
}

func NewProjectEnv(t fixenv.T) *ProjectEnv {
    return &ProjectEnv{EnvT: fixenv.New(t)}
}

func (e *ProjectEnv) Customer(name string) Customer {
    return seededAccount(e, name)
}
```

Existing fixtures that accept `fixenv.Env` continue to work, and you can add new methods or interfaces tailored to your project.

## Leveraging generic helpers

Go 1.18+ users can adopt the generic utilities in `env_generic_sugar.go` to eliminate manual type assertions. The helpers fit naturally into regular fixtures:

```go
// requires import "math/rand"
func randomNumber(e fixenv.Env) int {
    return fixenv.CacheResult(e, func() (*fixenv.GenericResult[int], error) {
        return fixenv.NewGenericResult(rand.Int()), nil
    })
}
```

`CacheResult` infers the return type, so the test that calls `randomNumber(e)` receives a plain `int` value without needing casts. See [`env_generic_sugar.go`](../env_generic_sugar.go) for additional helpers.

## Observability and debugging

- Enable `testing -run` filters to focus on a specific fixture.
- Use `Env.T().Logf` inside fixtures to emit diagnostic messages when cache hits or cleanups occur.
- Pair Fixenv with structured logging to trace fixture dependencies in complex suites.

With these techniques, Fixenv scales from simple helper functions to a robust fixture platform for large integration suites.
