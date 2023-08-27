//go:build go1.18
// +build go1.18

package fixenv

func Cache[TRes any](env Env, cacheKey any, opt *FixtureOptions, f func() (TRes, error)) TRes {
	addSkipLevel(&opt)
	callbackResult := env.Cache(cacheKey, opt, func() (res interface{}, err error) {
		return f()
	})

	var res TRes
	if callbackResult != nil {
		res = callbackResult.(TRes)
	}
	return res
}

func CacheWithCleanup[TRes any](env Env, cacheKey any, opt *FixtureOptions, f func() (TRes, FixtureCleanupFunc, error)) TRes {
	addSkipLevel(&opt)
	callbackResult := env.CacheWithCleanup(cacheKey, opt, func() (res interface{}, cleanup FixtureCleanupFunc, err error) {
		return f()
	})

	var res TRes
	if callbackResult != nil {
		res = callbackResult.(TRes)
	}
	return res
}

func CacheResult[TRes any](env Env, opts *CacheOptions, f GenericFixtureFunction[TRes]) TRes {
	addSkipLevelCache(&opts)
	var oldStyleFunc FixtureFunction = func() Result {
		res := f()
		return Result{
			Result:  res.Result,
			Error:   res.Error,
			Cleanup: res.Cleanup,
		}
	}
	res := env.CacheResult(opts, oldStyleFunc)
	return res.(TRes)
}

// GenericFixtureFunction - callback function with structured result
type GenericFixtureFunction[ResT any] func() GenericResult[ResT]

// GenericResult of fixture callback
type GenericResult[ResT any] struct {
	Result  ResT
	Error   error
	Cleanup FixtureCleanupFunc
}

func addSkipLevel(optspp **FixtureOptions) {
	if *optspp == nil {
		*optspp = &FixtureOptions{}
	}
	(*optspp).additionlSkipExternalCalls++
}

func addSkipLevelCache(optspp **CacheOptions) {
	if *optspp == nil {
		*optspp = &CacheOptions{}
	}
	(*optspp).additionlSkipExternalCalls++
}
