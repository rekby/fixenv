//go:build go1.18
// +build go1.18

package fixenv

import (
	"fmt"
	"github.com/rekby/fixenv/internal"
	"math/rand"
	"testing"
)

func TestCacheResultGeneric(t *testing.T) {
	t.Run("PassParams", func(t *testing.T) {
		inOpt := CacheOptions{
			CacheKey: 123,
			Scope:    ScopeTest,
		}

		cleanupCalledBack := 0

		env := envMock{onCacheResult: func(opt CacheOptions, f fixtureFunction) interface{} {
			opt.additionlSkipExternalCalls--
			requireEquals(t, inOpt, opt)
			res, _ := f()
			return res.Value
		}}

		f := func() (*GenericResult[int], error) {
			cleanup := func() {
				cleanupCalledBack++
			}
			return NewGenericResultWithCleanup(2, cleanup), nil
		}
		res := CacheResult(env, f, inOpt)
		requireEquals(t, 2, res)
	})
	t.Run("SkipAdditionalCache", func(t *testing.T) {
		test := &internal.TestMock{TestName: t.Name()}
		env := newTestEnv(test)

		f1 := func() int {
			return CacheResult(env, func() (*GenericResult[int], error) {
				return NewGenericResult(1), nil
			})
		}
		f2 := func() int {
			return CacheResult(env, func() (*GenericResult[int], error) {
				return NewGenericResult(2), nil
			})
		}

		requireEquals(t, 1, f1())
		requireEquals(t, 2, f2())
	})
	t.Run("NilResultReturnsZeroValue", func(t *testing.T) {
		tMock := &internal.TestMock{TestName: t.Name(), SkipGoexit: true}
		env := New(tMock)
		calls := 0
		fixture := func() (*GenericResult[*int], error) {
			calls++
			return nil, nil
		}
		res1 := CacheResult(env, fixture)
		res2 := CacheResult(env, fixture)
		requireNil(t, res1)
		requireNil(t, res2)
		requireEquals(t, 1, calls)
	})
}

func TestCacheResultPanic(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		tMock := &internal.TestMock{TestName: "mock", SkipGoexit: true}
		e := New(tMock)

		rndFix := func(e Env) int {
			return CacheResult(e, func() (*GenericResult[int], error) {
				return NewGenericResult(rand.Int()), nil
			})
		}
		first := rndFix(e)
		second := rndFix(e)

		requireEquals(t, first, second)
	})
	t.Run("Options", func(t *testing.T) {
		tMock := &internal.TestMock{TestName: "mock", SkipGoexit: true}
		e := New(tMock)

		rndFix := func(e Env, name string) int {
			return CacheResult(e, func() (*GenericResult[int], error) {
				return NewGenericResult(rand.Int()), nil
			}, CacheOptions{CacheKey: name})
		}
		first1 := rndFix(e, "first")
		first2 := rndFix(e, "first")
		second1 := rndFix(e, "second")
		second2 := rndFix(e, "second")

		requireEquals(t, first1, first2)
		requireEquals(t, second1, second2)
		requireNotEquals(t, first1, second1)
	})
	t.Run("Panic", func(t *testing.T) {
		tMock := &internal.TestMock{TestName: "mock", SkipGoexit: true}
		e := New(tMock)

		rndFix := func(e Env, name string) int {
			return CacheResult(e, func() (*GenericResult[int], error) {
				return NewGenericResult(rand.Int()), nil
			}, CacheOptions{CacheKey: name}, CacheOptions{CacheKey: name})
		}
		requirePanic(t, func() {
			rndFix(e, "first")
		})
	})
}

type envMock struct {
	onCacheResult func(opts CacheOptions, f fixtureFunction) interface{}
}

func (e envMock) T() T {
	panic("not implemented")
}

func (e envMock) cacheResult(f fixtureFunction, options ...CacheOptions) interface{} {
	var opts CacheOptions
	switch len(options) {
	case 0:
		// pass
	case 1:
		opts = options[0]
	default:
		panic(fmt.Errorf("max options len is 1, given: %v", len(options)))
	}
	return e.onCacheResult(opts, f)
}
