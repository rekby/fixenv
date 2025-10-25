# Example walkthroughs

The `examples/` directory contains runnable scenarios that demonstrate Fixenv features. Run them with:

```bash
go test ./examples/...
```

Each subdirectory is a self-contained module with descriptive tests.

## `simple`

A minimal demonstration of caching, cleanup, and scope selection.

- `random_fixture_test.go` shows how `CacheResult` ensures repeated calls reuse the cached value.
- `cleanup_test.go` illustrates `NewGenericResultWithCleanup` and verifies callbacks run in last-in, first-out order.
- `package_scope_test.go` demonstrates `ScopePackage` by caching a randomly generated value for the entire package.

To inspect the output of a single file:

```bash
go test ./examples/simple -run Random
```

## `custom_env`

Builds a custom environment wrapper that exposes project-specific helper methods.

- `env.go` embeds `*fixenv.EnvT` inside `Env`.
- `env_test.go` constructs the custom environment and reuses shared fixtures.

Use this pattern when you want to attach domain-specific helpers without rewriting existing fixtures.

## `sf_helpers`

Demonstrates how to benefit from the prebuilt fixtures in [`github.com/rekby/fixenv/sf`](../../sf) before writing your own helpers.

- `context_fixture_test.go` obtains a cancellable context that automatically terminates when the owning test finishes.
- `tempdir_fixture_test.go` provisions a temporary directory, shows cache reuse within a test, and verifies cleanup afterwards.
- `tcp_listener_fixture_test.go` uses a cached TCP listener and confirms that each scope gets an isolated instance.

These helpers make it easy to adopt Fixenv immediately—mix them with your own fixtures whenever you need extra customisation.

## `simple_main_test`

Illustrates how to run package-scoped fixtures by installing `TestMain`.

- `testmain_test.go` wraps the module's tests with `fixenv.RunTests` to create the package-level environment.
- `example_test.go` provides a generic fixture using Go 1.18 generics.

Run just this example with:

```bash
go test ./examples/simple_main_test -run TestFirst
```

## Extending the gallery

Have an idea for a new example—perhaps connecting to a message broker or seeding a temporary database? Contributions are welcome! Open an issue or pull request describing the scenario you would like to add.
