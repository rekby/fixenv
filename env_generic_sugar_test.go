//go:build go1.18
// +build go1.18

package fixenv

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCacheGeneric(t *testing.T) {
	inParams := 123
	inOpt := &FixtureOptions{Scope: ScopeTest}

	env := envMock{onCache: func(params interface{}, opt *FixtureOptions, f FixtureCallbackFunc) interface{} {
		require.Equal(t, inParams, params)
		require.Equal(t, inOpt, opt)
		res, _ := f()
		return res
	}}

	res := Cache(env, inParams, inOpt, func() (int, error) {
		return 2, nil
	})
	require.Equal(t, 2, res)
}

func TestCacheWithCleanupGeneric(t *testing.T) {
	inParams := 123
	inOpt := &FixtureOptions{Scope: ScopeTest}

	cleanupCalledBack := 0

	env := envMock{onCacheWithCleanup: func(params interface{}, opt *FixtureOptions, f FixtureCallbackWithCleanupFunc) interface{} {
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
