[![Go Reference](https://pkg.go.dev/badge/github.com/rekby/fixenv.svg)](https://pkg.go.dev/github.com/rekby/fixenv)
[![Coverage Status](https://coveralls.io/repos/github/rekby/fixenv/badge.svg?branch=master)](https://coveralls.io/github/rekby/fixenv?branch=master)
[![GoReportCard](https://goreportcard.com/badge/github.com/rekby/fixenv)](https://goreportcard.com/report/github.com/rekby/fixenv)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

# Fixenv

Pytest-inspired fixture caching for Go tests with precise scope control, automatic cleanup, and no external dependencies.

Fixenv helps Go developers describe repeatable test environments once and reuse them safely across an entire test suite. Fixtures are cached per configurable scope, dependencies between fixtures are tracked automatically, and cleanup hooks make resource lifecycle management explicit.

---

- [Features](#features)
- [Installation](#installation)
- [Quick start](#quick-start)
- [Fixture scopes & caching](#fixture-scopes--caching)
- [Cleanup & ordering guarantees](#cleanup--ordering-guarantees)
- [Documentation](#documentation)
- [Example gallery](#example-gallery)
- [Comparison with alternatives](#comparison-with-alternatives)
- [Project roadmap](#project-roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## Features

- **Scope-aware caching** – Run expensive setup only once per test, test tree, or package using [`CacheOptions.Scope`](https://pkg.go.dev/github.com/rekby/fixenv#CacheOptions).
- **Deterministic cleanup** – Register cleanup callbacks via [`NewGenericResultWithCleanup`](https://pkg.go.dev/github.com/rekby/fixenv#NewGenericResultWithCleanup); Fixenv invokes them when the owning scope ends.
- **Fixture dependency tree** – Compose fixtures freely; cached results are shared across the dependency tree to avoid duplicated work.
- **Skip-aware execution** – Return [`ErrSkipTest`](https://pkg.go.dev/github.com/rekby/fixenv#pkg-variables) from fixtures to short-circuit tests without leaking resources.
- **Extensible environment** – Embed [`EnvT`](https://pkg.go.dev/github.com/rekby/fixenv#EnvT) or implement the [`Env`](https://pkg.go.dev/github.com/rekby/fixenv#Env) interface to customise behaviour for your project.
- **Parallel-ready fixtures** – Caching and cleanup stay correct when tests call [`t.Parallel`](https://pkg.go.dev/testing#T.Parallel); each parallel test receives its own safe environment.
- **Zero dependencies** – Fixenv is a lightweight helper that plays well with the standard `testing` package.
- **Starter fixtures included** – Import [`github.com/rekby/fixenv/sf`](sf) for ready-to-use helpers like cancellable contexts, temporary directories, and TCP listeners.

## Installation

```bash
go get github.com/rekby/fixenv
```

Fixenv follows Go modules semantic import versioning. Re-run the command above to upgrade to the latest tag.

## Quick start

### Use bundled fixtures instantly

```go
package context_test

import (
    "testing"

    "github.com/rekby/fixenv"
    "github.com/rekby/fixenv/sf"
)

func TestAutoCanceledContext(t *testing.T) {
    e := fixenv.New(t)
    ctx := sf.Context(e)

    // Exercise your code under test with ctx.
    // When the test finishes, the context is cancelled automatically.
}
```

`sf.Context` ships with Fixenv and returns a cancellable context bound to the current test scope. When the test ends, the fixture automatically cancels the context through its registered cleanup—no manual teardown required.

### Write a custom fixture

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

This pattern—create an environment with `fixenv.New(t)`, wrap setup logic in `fixenv.CacheResult`, and optionally attach a cleanup—is the basis for all fixtures. Continue with the [Getting started guide](docs/getting-started.md) for step-by-step instructions and cleanup examples.

## Fixture scopes & caching

Fixtures default to `ScopeTest`: cached per `testing.T`. Subtests get their own scope unless you opt in to sharing. Override the scope through `CacheOptions` to share expensive setup more broadly:

| Scope | Lifetime | Typical use cases |
|-------|----------|-------------------|
| `ScopeTest` | Current `testing.T` only | Pure unit tests, short-lived resources |
| `ScopeTestAndSubtests` | Top-level test and all nested subtests | Reusing setup across a parent test and its subtests |
| `ScopePackage` | Entire package (requires `TestMain`) | Shared databases, external services, heavy caches |

Scope names combine automatically with the fixture's call site to produce a stable cache key. You can extend the cache key with serialisable parameters via `CacheOptions.CacheKey` when a fixture accepts arguments.

Learn more in [Scopes and lifetimes](docs/scopes-and-lifetimes.md).

## Cleanup & ordering guarantees

Each fixture may optionally return a cleanup callback. Fixenv registers the callback on the owning `testing.T` and guarantees **LIFO** execution when the scope ends. Cleanups run even if a test fails or is skipped via `ErrSkipTest`, making it safe to provision external services, create temporary files, or adjust global state.

See [Cleanup and ordering](docs/cleanup-and-ordering.md) for practical recipes.

## Documentation

The `docs/` directory provides extended guides:

- [Getting started](docs/getting-started.md) – installation, basic fixtures, and first tests.
- [Scopes and lifetimes](docs/scopes-and-lifetimes.md) – choosing the right cache level and structuring packages.
- [Advanced fixtures](docs/advanced-fixtures.md) – parameterised fixtures, dependency injection patterns, and custom environments.
- [Cleanup and ordering](docs/cleanup-and-ordering.md) – teardown techniques, deterministic ordering, and debugging tips.
- [Example walkthroughs](docs/examples/README.md) – narrative tours of the sample projects.

## Example gallery

Explore runnable examples under [`examples/`](examples):

| Example | Highlights |
| ------- | ---------- |
| [`simple`](examples/simple) | Minimal fixtures, scope defaults, and cleanup basics. |
| [`custom_env`](examples/custom_env) | Embedding `EnvT` in a domain-specific helper to expose project-specific fixtures. |
| [`sf_helpers`](examples/sf_helpers) | Using bundled fixtures from `sf` for contexts, temp directories, and local TCP listeners. |
| [`simple_main_test`](examples/simple_main_test) | Using `TestMain` to install package-level fixtures with cleanup control. |

Run the full set with:

```bash
go test ./examples/...
```

Detailed walkthroughs and expected outputs live in [docs/examples/README.md](docs/examples/README.md).

## Comparison with alternatives

| Feature / Project                | Fixenv | `testify/suite` | `dockertest` | `gotest.tools/fs` |
| -------------------------------- | :----: | :-------------: | :----------: | :---------------: |
| Scoped caching (test / package)  |   ✅   |    ⚪️ manual    |   ⚪️ manual  |    ⚪️ manual     |
| Declarative fixture tree         |   ✅   |       ⚪️       |      ⚪️      |        ⚪️        |
| Cleanup integration              | ✅ (LIFO with fixtures) | ⚪️ (`testing.T.Cleanup`) | ✅ (containers) | ⚪️ (`testing.T.Cleanup`) |
| Skip-aware setup                 | ✅ (`ErrSkipTest`) |       ⚪️       |      ⚪️      |        ⚪️        |
| Works with plain `testing.T`     |   ✅   |       ✅        |      ✅      |        ✅        |
| Extra runtime requirements       | Go stdlib only | Go stdlib only | Docker daemon | Go stdlib only |

Fixenv focuses on composing reusable fixtures with deterministic lifecycle control. It complements assertion libraries and environment provisioning tools—you can mix Fixenv with them instead of choosing only one approach.

## Project roadmap

- Additional built-in helpers for temporary directories and HTTP servers.
- Optional tracing hooks for fixture execution and cache hits.
- Community-contributed examples covering databases, message queues, and cloud resources.
- Automatic detection of scope-mixing mistakes (e.g. invoking a test-scoped fixture from a package-scoped one) to surface lifecycle issues early.

Interested in a feature? [Open an issue](https://github.com/rekby/fixenv/issues/new) or start a discussion.

## Contributing

Contributions are welcome! Please:

1. Fork the repository and create a feature branch.
2. Include thorough automated tests with each code change.
3. Run `go test ./...` before opening a pull request.
4. Describe your changes in detail and link to related issues.

Bug reports and documentation improvements are also appreciated.

## License

Licensed under the [MIT License](LICENSE.txt).
