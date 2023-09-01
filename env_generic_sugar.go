//go:build go1.18
// +build go1.18

package fixenv

import "fmt"

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

func CacheResult[TRes any](env Env, f GenericFixtureFunction[TRes], options ...CacheOptions) TRes {
	var cacheOptions CacheOptions
	switch len(options) {
	case 0:
		cacheOptions = CacheOptions{}
	case 1:
		cacheOptions = options[0]
	default:
		panic(fmt.Errorf("max len of cache result cacheOptions is 1, given: %v", len(options)))
	}

	addSkipLevelCache(&cacheOptions)
	var oldStyleFunc FixtureFunction = func() (*Result, error) {
		res, err := f()

		var oldStyleRes *Result
		if res != nil {
			oldStyleRes = &Result{
				Value:            res.Value,
				ResultAdditional: res.ResultAdditional,
			}
		}
		return oldStyleRes, err
	}
	res := env.CacheResult(oldStyleFunc, cacheOptions)
	return res.(TRes)
}

// GenericFixtureFunction - callback function with structured result
type GenericFixtureFunction[ResT any] func() (*GenericResult[ResT], error)

// GenericResult of fixture callback
type GenericResult[ResT any] struct {
	Value ResT
	ResultAdditional
}

func NewGenericResult[ResT any](res ResT) (*GenericResult[ResT], error) {
	return &GenericResult[ResT]{Value: res}, nil
}

func NewGenericResultWithCleanup[ResT any](res ResT, cleanup FixtureCleanupFunc) (*GenericResult[ResT], error) {
	return &GenericResult[ResT]{Value: res, ResultAdditional: ResultAdditional{Cleanup: cleanup}}, nil
}

func addSkipLevel(optspp **FixtureOptions) {
	if *optspp == nil {
		*optspp = &FixtureOptions{}
	}
	(*optspp).additionlSkipExternalCalls++
}

func addSkipLevelCache(optspp *CacheOptions) {
	(*optspp).additionlSkipExternalCalls++
}
