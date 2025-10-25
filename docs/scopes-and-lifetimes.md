# Scopes and lifetimes

Fixenv caches fixture results so expensive setup logic only runs when needed. Choosing the right scope keeps tests fast while preserving isolation.

## Scope overview

| Scope | Lifetime | Notes |
|-------|----------|-------|
| `ScopeTest` | Default. Cache is unique to each `testing.T`. | Safest option for unit tests. No extra setup required. |
| `ScopeTestAndSubtests` | Cache is shared between a top-level test and its descendants. | Let a parent test build data once and share it with its subtests. |
| `ScopePackage` | Cache is shared for the entire package. Requires `TestMain` to manage cleanups. | Ideal for costly resources such as databases or external services. |

All scopes respect parameterised cache keys. If a fixture accepts arguments, supply a serialisable key via `CacheOptions.CacheKey` to differentiate results.

For example, a fixture that creates a bank account with `ScopeTest` will provision a fresh record for each test, so parallel tests using the same customer name do not clash. Switching that fixture to `ScopePackage` would instead create one shared account that stays alive until the package finishes.

## Switching scopes

Pass `CacheOptions` as the second argument to `fixenv.CacheResult`:

```go
func suiteLevelConfig(e fixenv.Env) Config {
    return fixenv.CacheResult(e, func() (*fixenv.GenericResult[Config], error) {
        cfg := loadConfig()
        cleanup := func() { resetConfig() }
        return fixenv.NewGenericResultWithCleanup(cfg, cleanup), nil
    }, fixenv.CacheOptions{Scope: fixenv.ScopeTestAndSubtests})
}
```

For package-wide resources, register cleanups via `TestMain`:

```go
// requires imports "os" and "testing"
func TestMain(m *testing.M) {
    // requires import "github.com/rekby/fixenv"
    os.Exit(fixenv.RunTests(m))
}
```

Inside fixtures you can now cache with `ScopePackage` and be confident cleanups run once the process shuts down. `fixenv.RunTests` creates a package-level environment and wires cleanups to run after `m.Run()`.

## Parameterised fixtures

Fixtures can depend on arguments while still using scoped caching. Provide a JSON-serialisable cache key to distinguish values:

```go
func userAccount(e fixenv.Env, name string) Account {
    return fixenv.CacheResult(e, func() (*fixenv.GenericResult[Account], error) {
        acc := createAccount(name)
        cleanup := func() { deleteAccount(acc.ID) }
        return fixenv.NewGenericResultWithCleanup(acc, cleanup), nil
    }, fixenv.CacheOptions{CacheKey: name})
}
```

When one test calls `userAccount(e, "alice")` several times, the same account object is reused and its cleanup runs once. Another test—even if it runs in parallel—receives a separate account because it holds a different `testing.T` and therefore a different cache.

## When to choose each scope

- **`ScopeTest`** – default for unit tests, or when fixture outputs are mutated.
- **`ScopeTestAndSubtests`** – parent test performs setup and subtests make assertions against shared state.
- **`ScopePackage`** – provisioning external services, expensive database migrations, or large datasets.

If in doubt, start with `ScopeTest` and promote individual fixtures to broader scopes as performance bottlenecks appear.
