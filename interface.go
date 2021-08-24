package fixenv

// FixtureInternalFunc - function, which result can cached
// res - result for cache.
// if err not nil - T().Fatalf() will called with error message
// if res exit without return (panic, GoExit, t.FailNow, ...)
// then cache error about unexpected exit
type FixtureInternalFunc func() (res interface{}, err error)

// Env - fixture cache engine.
type Env interface {
	// T - return t object of current test/benchmark.
	T() T

	Cache(params interface{}, f FixtureInternalFunc, opt *FixtureOptions) interface{}
}

// T is subtime of testing.TB
// it can extended from time to time, but include only public methods from testing.TB
type T interface {
	Cleanup(func())
	Fatalf(format string, args ...interface{})
	Name() string
}
