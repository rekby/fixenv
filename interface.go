package fixenv

import "errors"

// Env - fixture cache engine.
type Env interface {
	// T - return t object of current test/benchmark.
	T() T

	// CacheResult add result of call f to cache and return same result for all
	// calls for the same function and cache options within cache scope
	CacheResult(f FixtureFunction, options ...CacheOptions) interface{}

	// Cache result of f calls
	// f call exactly once for every combination of scope and params
	// params must be json serializable (deserialize not need)
	// Deprecated: Use CacheResult
	Cache(cacheKey interface{}, opt *FixtureOptions, f FixtureCallbackFunc) interface{}

	// CacheWithCleanup cache result of f calls
	// f call exactly once for every combination of scope and params
	// params must be json serializable (deserialize not need)
	// Deprecated: Use CacheResult
	CacheWithCleanup(cacheKey interface{}, opt *FixtureOptions, f FixtureCallbackWithCleanupFunc) interface{}
}

var (
	// ErrSkipTest - error for return from fixtures
	// return the error mean skip test and cache decision about skip test for feature fixtures call
	// as usual result/error cache.
	//
	// Use special error instead of detect of test.SkipNow() need for prevent run fixture in separate goroutine for
	// skip detecting
	ErrSkipTest = errors.New("skip test")
)

// CacheScope define life time of fixture value
// and allow use independent fixture values for different scopes, but share same value for
// one scope, which can be more then one test
type CacheScope int

const (
	// ScopeTest mean fixture function with same parameters called once per every test and subtests. Default value.
	// Second and more calls will use cached value.
	ScopeTest CacheScope = iota

	// ScopePackage mean fixture function with same parameters called once per package
	// for use the scope with TearDown function developer must initialize global handler and cleaner at TestMain.
	ScopePackage

	// ScopeTestAndSubtests mean fixture cached for top level test and subtests
	ScopeTestAndSubtests
)

// FixtureCallbackFunc - function, which result can cached
// res - result for cache.
// if err not nil - T().Fatalf() will called with error message
// if res exit without return (panic, GoExit, t.FailNow, ...)
// then cache error about unexpected exit
type FixtureCallbackFunc func() (res interface{}, err error)

// FixtureCallbackWithCleanupFunc - function, which result can cached
// res - result for cache.
// cleanup - if not nil - call on fixture cleanup. It called exactly once for every successfully call fixture
// if err not nil - T().Fatalf() will called with error message
// if res exit without return (panic, GoExit, t.FailNow, ...)
// then cache error about unexpected exit
type FixtureCallbackWithCleanupFunc func() (res interface{}, cleanup FixtureCleanupFunc, err error)

// FixtureCleanupFunc - callback function for cleanup after
// fixture value out from lifetime scope
// it called exactly once for every succesully call fixture
type FixtureCleanupFunc func()

// FixtureOptions options for fixenv engine
// for custom manage fixture
type FixtureOptions struct {
	// Scope for cache result
	Scope CacheScope

	additionlSkipExternalCalls int

	// cleanupFunc if not nil - called for cleanup fixture results
	// internal implementation details
	cleanupFunc FixtureCleanupFunc
}

// FixtureFunction - callback function with structured result
type FixtureFunction func() (*Result, error)

// Result of fixture callback
type Result struct {
	Value interface{}
	ResultAdditional
}

type ResultAdditional struct {
	Cleanup FixtureCleanupFunc
}

func NewResult(res interface{}) *Result {
	return &Result{Value: res}
}

func NewResultWithCleanup(res interface{}, cleanup FixtureCleanupFunc) *Result {
	return &Result{Value: res, ResultAdditional: ResultAdditional{Cleanup: cleanup}}
}

type CacheOptions struct {
	// Scope for cache result
	Scope CacheScope

	// Key for cache results, must be json serializable value
	CacheKey interface{}

	additionlSkipExternalCalls int
}

// T is subtype of testing.TB
type T interface {
	// Cleanup registers a function to be called when the test (or subtest) and all its subtests complete.
	// Cleanup functions will be called in last added, first called order.
	Cleanup(func())

	// Fatalf is equivalent to Logf followed by FailNow.
	//
	// Logf formats its arguments according to the format, analogous to Printf, and records the text in the error log.
	// A final newline is added if not provided. For tests, the text will be printed only if the test fails or the -test.v flag is set.
	// For benchmarks, the text is always printed to avoid having performance depend on the value of the -test.v flag.
	//
	// FailNow marks the function as having failed and stops its execution by calling runtime.Goexit
	// (which then runs all deferred calls in the current goroutine). Execution will continue at the next test or benchmark. FailNow must be called from the goroutine running the test or benchmark function, not from other goroutines created during the test. Calling FailNow does not stop those other goroutines.
	Fatalf(format string, args ...interface{})

	// Logf formats its arguments according to the format, analogous to Printf, and
	// records the text in the error log. A final newline is added if not provided. For
	// tests, the text will be printed only if the test fails or the -test.v flag is
	// set. For benchmarks, the text is always printed to avoid having performance
	// depend on the value of the -test.v flag.
	Logf(format string, args ...interface{})

	// Name returns the name of the running (sub-) test or benchmark.
	//
	// The name will include the name of the test along with the names
	// of any nested sub-tests. If two sibling sub-tests have the same name,
	// Name will append a suffix to guarantee the returned name is unique.
	Name() string

	// SkipNow is followed by testing.T.SkipNow().
	// Don't use SkipNow() for skip test from fixture - use special error ErrSkipTest for it.
	//
	// SkipNow marks the test as having been skipped and stops its execution
	// by calling runtime.Goexit.
	// If a test fails (see Error, Errorf, Fail) and is then skipped,
	// it is still considered to have failed.
	// Execution will continue at the next test or benchmark. See also FailNow.
	// SkipNow must be called from the goroutine running the test, not from
	// other goroutines created during the test. Calling SkipNow does not stop
	// those other goroutines.
	SkipNow()

	// Skipped reports whether the test was skipped.
	Skipped() bool
}
