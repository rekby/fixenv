//go:build go1.18
// +build go1.18

package fixenv

func Cache[TRes any](env Env, params any, opt *FixtureOptions, f func() (TRes, error)) TRes {
	callbackResult := env.Cache(params, opt, func() (res interface{}, err error) {
		return f()
	})

	var res TRes
	if callbackResult != nil {
		res = callbackResult.(TRes)
	}
	return res
}

func CacheWithCleanup[TRes any](env Env, params any, opt *FixtureOptions, f func() (TRes, FixtureCleanupFunc, error)) TRes {
	callbackResult := env.CacheWithCleanup(params, opt, func() (res interface{}, cleanup FixtureCleanupFunc, err error) {
		return f()
	})

	var res TRes
	if callbackResult != nil {
		res = callbackResult.(TRes)
	}
	return res
}
