# Cleanup and ordering

Fixenv guarantees deterministic cleanup execution so resources are released when their scope ends.

## Registering cleanup callbacks

Return a result created with `fixenv.NewGenericResultWithCleanup` to register a callback. The callback runs exactly once when the fixture leaves scope.

```go
// requires import "os"
func temporaryFile(e fixenv.Env) string {
    return fixenv.CacheResult(e, func() (*fixenv.GenericResult[string], error) {
        file, err := os.CreateTemp("", "example-*.txt")
        if err != nil {
            return nil, err
        }

        cleanup := func() {
            file.Close()
            os.Remove(file.Name())
        }

        return fixenv.NewGenericResultWithCleanup(file.Name(), cleanup), nil
    })
}
```

Cleanups run even if the test fails or is skipped. They also run when the package-wide environment shuts down via `fixenv.RunTests`.

## Execution order

Fixenv mirrors the behaviour of `testing.T.Cleanup`: callbacks execute in **last-in, first-out** order. Nested fixtures therefore clean up from the inside out automatically.

```go
// requires imports "database/sql" and "github.com/rekby/fixenv"
func database(e fixenv.Env) *sql.DB {
    return fixenv.CacheResult(e, func() (*fixenv.GenericResult[*sql.DB], error) {
        db := openDatabase()
        return fixenv.NewGenericResultWithCleanup(db, func() { db.Close() }), nil
    })
}

// tableFixture depends on database and runs its cleanup before the database cleanup.
func tableFixture(e fixenv.Env, name string) string {
    key := fixenv.CacheOptions{CacheKey: name}
    return fixenv.CacheResult(e, func() (*fixenv.GenericResult[string], error) {
        db := database(e)
        createTable(db, name)
        cleanup := func() { dropTable(db, name) }
        return fixenv.NewGenericResultWithCleanup(name, cleanup), nil
    }, key)
}
```

When `tableFixture` leaves scope, its cleanup runs first, followed by the `database` cleanup. There is no extra API for manual ordering—the nesting of fixture calls already defines the order.

## Troubleshooting leaked resources

- Ensure every code path in the fixture returns a result with the appropriate cleanup.
- Add logging in cleanups (`e.T().Logf`) to confirm execution order.
- When debugging package-level fixtures, remember that cleanups run after `m.Run()` inside `TestMain` once `fixenv.RunTests` completes.
