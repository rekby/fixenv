package fixenv

import (
	"errors"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testMock struct {
	name string

	m        sync.Mutex
	cleanups []func()
	fatals   []struct {
		format string
		args   []interface{}
	}
}

func (e *EnvT) cloneWithTest(t T) *EnvT {
	e2 := newEnv(t, e.c, e.m, e.scopes)
	e2.onCreate()
	return e2
}

func newTestEnv(t T) *EnvT {
	e := newEnv(t, newCache(), &sync.Mutex{}, make(map[string]*scopeInfo))
	e.onCreate()
	return e
}

func (t *testMock) callCleanup() {
	for i := len(t.cleanups) - 1; i >= 0; i-- {
		t.cleanups[i]()
	}
}

func (t *testMock) Cleanup(f func()) {
	t.m.Lock()
	defer t.m.Unlock()

	t.cleanups = append(t.cleanups, f)
}

func (t *testMock) Fatalf(format string, args ...interface{}) {
	t.m.Lock()
	defer t.m.Unlock()

	t.fatals = append(t.fatals, struct {
		format string
		args   []interface{}
	}{format: format, args: args})

}

func (t *testMock) Name() string {
	if t.name == "" {
		return "mock"
	}
	return t.name
}

func Test_Env__NewEnv(t *testing.T) {
	t.Run("create_new_env", func(t *testing.T) {
		at := assert.New(t)

		tMock := &testMock{name: "mock"}
		defer tMock.callCleanup()

		e := NewEnv(tMock)
		at.Equal(tMock, e.t)
		at.Equal(globalCache, e.c)
		at.Equal(globalMutex, e.m)
		at.Equal(globalScopeInfo, e.scopes)
		at.Len(globalCache.store, 0)
		at.Len(globalScopeInfo, 1)
		at.Len(tMock.cleanups, 1)
	})

	t.Run("global_info_cleaned", func(t *testing.T) {
		at := assert.New(t)
		at.Len(globalCache.store, 0)
		at.Len(globalScopeInfo, 0)
	})

	t.Run("double_env_same_scope_same_time", func(t *testing.T) {
		at := assert.New(t)

		tMock := &testMock{name: "mock"}
		defer tMock.callCleanup()

		_ = NewEnv(tMock)
		at.Len(tMock.fatals, 0)

		_ = NewEnv(tMock)
		at.Len(tMock.fatals, 1)
	})

	t.Run("double_env_similar_scope_different_time", func(t *testing.T) {
		t.Run("test", func(t *testing.T) {
			_ = NewEnv(t)
		})
		t.Run("test", func(t *testing.T) {
			_ = NewEnv(t)
		})
	})
}

func Test_Env_Cache(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		at := assert.New(t)
		e := NewEnv(t)

		val := 0
		cntF := func() int {
			res := e.Cache(nil, nil, func() (res interface{}, err error) {
				val++
				return val, nil
			})
			return res.(int)
		}

		at.Equal(1, cntF())
		at.Equal(1, cntF())

		val = 2
		at.Equal(1, cntF())
	})

	t.Run("subtest_and_test_scope", func(t *testing.T) {
		at := assert.New(t)
		e := NewEnv(t)

		val := 0
		cntF := func(env Env) int {
			res := env.Cache(nil, &FixtureOptions{Scope: ScopeTest}, func() (res interface{}, err error) {
				val++
				return val, nil
			})
			return res.(int)
		}

		at.Equal(1, cntF(e))
		at.Equal(1, cntF(e))

		t.Run("subtest", func(t *testing.T) {
			at := assert.New(t)
			subEnv := NewEnv(t)
			at.Equal(2, cntF(subEnv))
			at.Equal(2, cntF(subEnv))
		})

		at.Equal(1, cntF(e))

	})

	t.Run("subtest_and_package_scope", func(t *testing.T) {
		at := assert.New(t)
		e := NewEnv(t)
		_, mainClose := CreateMainTestEnv(nil)
		defer mainClose()

		val := 0
		cntF := func(env Env) int {
			res := env.Cache(nil, &FixtureOptions{Scope: ScopePackage}, func() (res interface{}, err error) {
				val++
				return val, nil
			})
			return res.(int)
		}

		at.Equal(1, cntF(e))
		at.Equal(1, cntF(e))

		t.Run("subtest", func(t *testing.T) {
			at := assert.New(t)
			subEnv := NewEnv(t)
			at.Equal(1, cntF(subEnv))
			at.Equal(1, cntF(subEnv))
		})

		at.Equal(1, cntF(e))

	})

	t.Run("fail_on_fixture_err", func(t *testing.T) {
		at := assert.New(t)

		tMock := &testMock{name: "mock"}
		defer tMock.callCleanup()

		e := NewEnv(tMock)
		at.Len(tMock.fatals, 0)
		e.Cache(nil, nil, func() (res interface{}, err error) {
			return nil, errors.New("test")
		})
		at.Len(tMock.fatals, 1)
	})

	t.Run("not_serializable_param", func(t *testing.T) {
		at := assert.New(t)

		type paramT struct {
			F func() // can't serialize func to json
		}
		param := paramT{}
		tMock := &testMock{}
		defer tMock.callCleanup()
		e := NewEnv(tMock)
		e.Cache(param, nil, func() (res interface{}, err error) {
			return nil, nil
		})
		at.Len(tMock.fatals, 1)
	})

	t.Run("cache_by_caller_func", func(t *testing.T) {
		at := assert.New(t)
		tMock := &testMock{name: "mock"}
		e := newTestEnv(tMock)

		cnt := 0
		res := e.Cache(nil, &FixtureOptions{Scope: ScopeTest}, func() (res interface{}, err error) {
			cnt++
			return cnt, nil
		})
		at.Equal(1, res)

		res = e.Cache(nil, &FixtureOptions{Scope: ScopeTest}, func() (res interface{}, err error) {
			cnt++
			return cnt, nil
		})
		at.Equal(1, res)
	})

	t.Run("different_cache_for_diff_anonim_function", func(t *testing.T) {
		at := assert.New(t)
		tMock := &testMock{name: "mock"}
		e := newTestEnv(tMock)

		cnt := 0
		func() {
			res := e.Cache(nil, &FixtureOptions{Scope: ScopeTest}, func() (res interface{}, err error) {
				cnt++
				return cnt, nil
			})
			at.Equal(1, res)
		}()

		func() {
			res := e.Cache(nil, &FixtureOptions{Scope: ScopeTest}, func() (res interface{}, err error) {
				cnt++
				return cnt, nil
			})
			at.Equal(2, res)
		}()

	})
}

func Test_FixtureWrapper(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		at := assert.New(t)

		tMock := &testMock{name: "mock"}
		defer tMock.callCleanup()

		e := newTestEnv(tMock)
		key := cacheKey("asd")

		cnt := 0
		w := e.fixtureCallWrapper(key, func() (res interface{}, err error) {
			cnt++
			return cnt, errors.New("test")
		}, &FixtureOptions{})
		si := e.scopes[scopeName(tMock.Name(), ScopeTest)]
		at.Equal(0, cnt)
		at.Len(si.cacheKeys, 0)
		res1, err := w()
		at.Equal(1, res1)
		at.EqualError(err, "test")
		at.Equal(1, cnt)
		at.Equal([]cacheKey{key}, si.cacheKeys)

		cnt = 0
		key2 := cacheKey("asd")
		cleanupsLen := len(tMock.cleanups)
		w = e.fixtureCallWrapper(key2, func() (res interface{}, err error) {
			cnt++
			return cnt, nil
		}, &FixtureOptions{CleanupFunc: func() {

		}})
		at.Len(tMock.cleanups, cleanupsLen)
		_, _ = w()
		at.Equal([]cacheKey{key, key2}, si.cacheKeys)
		at.Len(tMock.cleanups, cleanupsLen+1)
	})

	t.Run("unknown_scope_info", func(t *testing.T) {
		at := assert.New(t)

		tMock := &testMock{name: "mock"}
		defer tMock.callCleanup()
		e := newTestEnv(tMock)
		tMock.name = "mock2"

		w := e.fixtureCallWrapper("asd", func() (res interface{}, err error) {
			return nil, nil
		}, &FixtureOptions{})
		_, _ = w()
		at.Len(tMock.fatals, 1)

	})
}

func Test_Env_Cleanup(t *testing.T) {
	at := assert.New(t)
	tMock := &testMock{}

	e := newTestEnv(tMock)
	cleanups := len(tMock.cleanups)

	e.Cleanup(func() {})
	at.Len(tMock.cleanups, cleanups+1)
}

func Test_Env_T(t *testing.T) {
	at := assert.New(t)
	e := NewEnv(t)
	at.Equal(t, e.T())
}

func Test_Env_TearDown(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		at := assert.New(t)

		t1 := &testMock{name: "mock"}
		// defer t1.callCleanup - direct call e1.tearDown - for test

		e1 := newTestEnv(t1)
		at.Len(e1.scopes, 1)
		at.Len(e1.scopes[scopeName(t1.name, ScopeTest)].Keys(), 0)
		at.Len(e1.c.store, 0)

		e1.Cache(1, nil, func() (res interface{}, err error) {
			return nil, nil
		})
		e1.Cache(2, nil, func() (res interface{}, err error) {
			return nil, nil
		})
		at.Len(e1.scopes, 1)
		at.Len(e1.scopes[scopeName(t1.name, ScopeTest)].Keys(), 2)
		at.Len(e1.c.store, 2)

		t2 := &testMock{name: "mock2"}
		// defer t2.callCleanup - direct call e2.tearDown - for test

		e2 := e1.cloneWithTest(t2)
		at.Len(e1.scopes, 2)
		at.Len(e1.scopes[scopeName(t1.name, ScopeTest)].Keys(), 2)
		at.Len(e1.scopes[scopeName(t2.name, ScopeTest)].Keys(), 0)
		at.Len(e1.c.store, 2)

		e2.Cache(1, nil, func() (res interface{}, err error) {
			return nil, nil
		})

		at.Len(e1.scopes, 2)
		at.Len(e1.scopes[scopeName(t1.name, ScopeTest)].Keys(), 2)
		at.Len(e1.scopes[scopeName(t2.name, ScopeTest)].Keys(), 1)
		at.Len(e1.c.store, 3)

		// finish first test and tearDown e1
		e1.tearDown()
		at.Len(e1.scopes, 1)
		at.Len(e1.scopes[scopeName(t2.name, ScopeTest)].Keys(), 1)
		at.Len(e1.c.store, 1)

		e2.tearDown()
		at.Len(e1.scopes, 0)
		at.Len(e1.c.store, 0)
	})

	t.Run("tearDown on unexisted scope", func(t *testing.T) {
		at := assert.New(t)
		tMock := &testMock{name: "mock"}
		// defer tMock.callCleanups. e.tearDown will call directly for test
		e := NewEnv(tMock)

		for key := range e.scopes {
			delete(e.scopes, key)
		}

		e.tearDown()
		at.Len(tMock.fatals, 1)
	})
}

func Test_MakeCacheKey(t *testing.T) {
	at := assert.New(t)

	var res cacheKey
	var err error

	envFunc := func() {
		res, err = makeCacheKey("asdf", 222, globalEmptyFixtureOptions, true)
	}
	envFunc()
	at.NoError(err)

	expected := cacheKey(`{"func":"github.com/rekby/fixenv.Test_MakeCacheKey","fname":".../env_test.go","scope":0,"scope_name":"asdf","params":222}`)
	at.JSONEq(string(expected), string(res))
}

func Test_MakeCacheKeyFromFrame(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		at := assert.New(t)

		key, err := makeCacheKeyFromFrame(123, ScopeTest, runtime.Frame{
			PC:       0,
			Function: "func_name",
			File:     "/asd/file_name.go",
		}, "scope-name", false)
		at.NoError(err)
		at.JSONEq(`{
	"scope": 0,
	"scope_name": "scope-name",
	"func": "func_name",
	"fname": "/asd/file_name.go",
	"params": 123
}`, string(key))
	})

	t.Run("test_call", func(t *testing.T) {
		at := assert.New(t)

		key, err := makeCacheKeyFromFrame(123, ScopeTest, runtime.Frame{
			PC:       0,
			Function: "func_name",
			File:     "/asd/file_name.go",
		}, "scope-name", true)
		at.NoError(err)
		at.JSONEq(`{
	"scope": 0,
	"scope_name": "scope-name",
	"func": "func_name",
	"fname": ".../file_name.go",
	"params": 123
}`, string(key))
	})

	t.Run("no_func_name", func(t *testing.T) {
		at := assert.New(t)

		_, err := makeCacheKeyFromFrame(123, ScopeTest, runtime.Frame{
			PC:       0,
			Function: "",
			File:     "/asd/file_name.go",
		}, "scope-name", true)
		at.Error(err)
	})

	t.Run("no_file_name", func(t *testing.T) {
		at := assert.New(t)

		_, err := makeCacheKeyFromFrame(123, ScopeTest, runtime.Frame{
			PC:       0,
			Function: "func_name",
			File:     "",
		}, "scope-name", true)
		at.Error(err)
	})

	t.Run("not_serializable_param", func(t *testing.T) {
		at := assert.New(t)

		type TStruct struct {
			F func()
		}

		_, err := makeCacheKeyFromFrame(TStruct{}, ScopeTest, runtime.Frame{
			PC:       0,
			Function: "func_name",
			File:     "/asd/file_name.go",
		}, "scope-name", true)
		at.Error(err)
	})
}

func Test_ScopeName(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		table := []struct {
			name     string
			testName string
			scope    CacheScope
			result   string
		}{
			{
				name:     "package",
				testName: "Test",
				scope:    ScopePackage,
				result:   packageScopeName,
			},
			{
				name:     "simple",
				testName: "Test",
				scope:    ScopeTest,
				result:   "Test",
			},
			{
				name:     "subtest",
				testName: "Test/subtest",
				scope:    ScopeTest,
				result:   "Test/subtest",
			},
			{
				name:     "subtests with TestAndSubtests level",
				testName: "Test/subtest",
				scope:    ScopeTestAndSubtests,
				result:   "Test",
			},
		}

		for _, c := range table {
			t.Run(c.name, func(t *testing.T) {
				at := assert.New(t)
				at.Equal(c.result, scopeName(c.testName, c.scope))
			})
		}
	})

	t.Run("unexpected_scope", func(t *testing.T) {
		at := assert.New(t)
		at.Panics(func() {
			scopeName("asd", -1)
		})
	})
}
