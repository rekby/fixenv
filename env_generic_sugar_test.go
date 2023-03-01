//go:build go1.18
// +build go1.18

package fixenv

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCacheGeneric(t *testing.T) {
	t.Run("PassParams", func(t *testing.T) {
		inParams := 123
		inOpt := &FixtureOptions{Scope: ScopeTest}

		env := envMock{onCache: func(params interface{}, opt *FixtureOptions, f FixtureCallbackFunc) interface{} {
			opt.additionlSkipExternalCalls--
			require.Equal(t, inParams, params)
			require.Equal(t, inOpt, opt)
			res, _ := f()
			return res
		}}

		res := Cache(env, inParams, inOpt, func() (int, error) {
			return 2, nil
		})
		require.Equal(t, 2, res)
	})
	t.Run("SkipAdditionalCache", func(t *testing.T) {
		test := &testMock{TestName: t.Name()}
		env := newTestEnv(test)

		f1 := func() int {
			return Cache(env, nil, nil, func() (int, error) {
				return 1, nil
			})
		}
		f2 := func() int {
			return Cache(env, nil, nil, func() (int, error) {
				return 2, nil
			})
		}

		require.Equal(t, 1, f1())
		require.Equal(t, 2, f2())
	})
}

func TestCacheWithCleanupGeneric(t *testing.T) {
	t.Run("PassParams", func(t *testing.T) {
		inParams := 123
		inOpt := &FixtureOptions{Scope: ScopeTest}

		cleanupCalledBack := 0

		env := envMock{onCacheWithCleanup: func(params interface{}, opt *FixtureOptions, f FixtureCallbackWithCleanupFunc) interface{} {
			opt.additionlSkipExternalCalls--
			require.Equal(t, inParams, params)
			require.Equal(t, inOpt, opt)
			res, _, _ := f()
			return res
		}}

		res := CacheWithCleanup(env, inParams, inOpt, func() (int, FixtureCleanupFunc, error) {
			cleanup := func() {
				cleanupCalledBack++
			}
			return 2, cleanup, nil
		})
		require.Equal(t, 2, res)
	})
	t.Run("SkipAdditionalCache", func(t *testing.T) {
		test := &testMock{TestName: t.Name()}
		env := newTestEnv(test)

		f1 := func() int {
			return CacheWithCleanup(env, nil, nil, func() (int, FixtureCleanupFunc, error) {
				return 1, nil, nil
			})
		}
		f2 := func() int {
			return CacheWithCleanup(env, nil, nil, func() (int, FixtureCleanupFunc, error) {
				return 2, nil, nil
			})
		}

		require.Equal(t, 1, f1())
		require.Equal(t, 2, f2())
	})

}

type envMock struct {
	onCache            func(params interface{}, opt *FixtureOptions, f FixtureCallbackFunc) interface{}
	onCacheWithCleanup func(params interface{}, opt *FixtureOptions, f FixtureCallbackWithCleanupFunc) interface{}
}

func (e envMock) T() T {
	panic("not implemented")
}

func (e envMock) Cache(params interface{}, opt *FixtureOptions, f FixtureCallbackFunc) interface{} {
	return e.onCache(params, opt, f)
}

func (e envMock) CacheWithCleanup(params interface{}, opt *FixtureOptions, f FixtureCallbackWithCleanupFunc) interface{} {
	return e.onCacheWithCleanup(params, opt, f)
}
