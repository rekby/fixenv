package fixenv

type CacheScope int

const (
	// ScopeTest mean fixture function with same parameters called once per every test and subtests. Default value.
	// Second and more calls will use cached value.
	ScopeTest CacheScope = iota
)

type FixtureOptions struct {
	// Scope for cache result
	Scope CacheScope

	// TearDown if not nil - called for cleanup fixture results
	// TearDown called exactly once for every succesully call fixture
	TearDown func()
}
