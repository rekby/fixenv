package fixenv

import (
	"errors"
	"github.com/rekby/fixenv/internal"
	"math/rand"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func Test_Env__NewEnv(t *testing.T) {
	t.Run("create_new_env", func(t *testing.T) {
		initGlobalState()
		at := assert.New(t)

		tMock := &internal.TestMock{TestName: "mock"}
		defer tMock.CallCleanup()

		e := New(tMock)
		at.Equal(tMock, e.t)
		at.Equal(globalCache, e.c)
		at.Equal(&globalMutex, e.m)
		at.Equal(globalScopeInfo, e.scopes)
		at.Len(globalCache.store, 0)
		at.Len(globalScopeInfo, 1)
		at.Len(tMock.Cleanups, 1)
	})

	t.Run("global_info_cleaned", func(t *testing.T) {
		at := assert.New(t)
		at.Len(globalCache.store, 0)
		at.Len(globalScopeInfo, 0)
	})

	t.Run("double_env_same_scope_same_time", func(t *testing.T) {
		at := assert.New(t)

		tMock := &internal.TestMock{TestName: "mock"}
		defer tMock.CallCleanup()

		_ = New(tMock)
		at.Len(tMock.Fatals, 0)

		runUntilFatal(func() {
			_ = New(tMock)
		})
		at.Len(tMock.Fatals, 1)
	})

	t.Run("double_env_similar_scope_different_time", func(t *testing.T) {
		t.Run("test", func(t *testing.T) {
			_ = New(t)
		})
		t.Run("test", func(t *testing.T) {
			_ = New(t)
		})
	})
}

func testFailedFixture(env Env) {
	env.Cache(nil, nil, func() (res interface{}, err error) {
		return nil, errors.New("test error")
	})
}

func Test_Env_Cache(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		at := assert.New(t)
		e := New(t)

		val := 0
		cntF := func() int {
			res := e.Cache(nil, nil, func() (res interface{}, err error) {
				val++
				e.T().Logf("val: ", val)
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
		e := New(t)

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
			subEnv := New(t)
			at.Equal(2, cntF(subEnv))
			at.Equal(2, cntF(subEnv))
		})

		at.Equal(1, cntF(e))

	})

	t.Run("subtest_and_package_scope", func(t *testing.T) {
		at := assert.New(t)
		e := New(t)
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
			subEnv := New(t)
			at.Equal(1, cntF(subEnv))
			at.Equal(1, cntF(subEnv))
		})

		at.Equal(1, cntF(e))

	})

	t.Run("fail_on_fixture_err", func(t *testing.T) {
		at := assert.New(t)

		tMock := &internal.TestMock{TestName: "mock"}
		defer tMock.CallCleanup()

		e := newTestEnv(tMock)
		at.Len(tMock.Fatals, 0)

		runUntilFatal(func() {
			testFailedFixture(e)
		})
		at.Len(tMock.Fatals, 1)

		// log message contains fixture name
		at.Contains(tMock.Fatals[0].ResultString, "testFailedFixture")
	})

	t.Run("not_serializable_param", func(t *testing.T) {
		at := assert.New(t)

		type paramT struct {
			F func() // can't serialize func to json
		}
		param := paramT{}
		tMock := &internal.TestMock{TestName: "mock"}
		defer tMock.CallCleanup()
		e := newTestEnv(tMock)
		runUntilFatal(func() {
			e.Cache(param, nil, func() (res interface{}, err error) {
				return nil, nil
			})
		})
		at.Len(tMock.Fatals, 1)
	})

	t.Run("cache_by_caller_func", func(t *testing.T) {
		at := assert.New(t)
		tMock := &internal.TestMock{TestName: "mock"}
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
		tMock := &internal.TestMock{TestName: "mock"}
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

	t.Run("check_unreachable_code", func(t *testing.T) {
		at := assert.New(t)

		tMock := &internal.TestMock{TestName: "mock", SkipGoexit: true}
		e := New(tMock)
		at.Panics(func() {
			e.Cache(nil, nil, func() (res interface{}, err error) {
				return nil, ErrSkipTest
			})
		})
		at.Equal(1, tMock.SkipCount)
	})
}

func Test_Env_CacheWithCleanup(t *testing.T) {
	t.Run("NilCleanup", func(t *testing.T) {
		tMock := &internal.TestMock{TestName: t.Name()}
		env := newTestEnv(tMock)

		callbackCalled := 0
		var callbackFunc FixtureCallbackWithCleanupFunc = func() (res interface{}, cleanup FixtureCleanupFunc, err error) {
			callbackCalled++
			return callbackCalled, nil, nil
		}

		res := env.CacheWithCleanup(nil, nil, callbackFunc)
		require.Equal(t, 1, res)
		require.Equal(t, 1, callbackCalled)

		// got value from cache
		res = env.CacheWithCleanup(nil, nil, callbackFunc)
		require.Equal(t, 1, res)
		require.Equal(t, 1, callbackCalled)
	})

	t.Run("WithCleanup", func(t *testing.T) {
		tMock := &internal.TestMock{TestName: t.Name()}
		env := newTestEnv(tMock)

		callbackCalled := 0
		cleanupCalled := 0
		var callbackFunc FixtureCallbackWithCleanupFunc = func() (res interface{}, cleanup FixtureCleanupFunc, err error) {
			callbackCalled++
			cleanup = func() {
				cleanupCalled++
			}
			return callbackCalled, cleanup, nil
		}

		res := env.CacheWithCleanup(nil, nil, callbackFunc)
		require.Equal(t, 1, res)
		require.Equal(t, 1, callbackCalled)
		require.Equal(t, cleanupCalled, 0)

		// got value from cache
		res = env.CacheWithCleanup(nil, nil, callbackFunc)
		require.Equal(t, 1, res)
		require.Equal(t, 1, callbackCalled)
		require.Equal(t, cleanupCalled, 0)

		tMock.CallCleanup()
		require.Equal(t, 1, callbackCalled)
		require.Equal(t, 1, cleanupCalled)
	})
}

func Test_Env_CacheResult(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		at := assert.New(t)
		tMock := &internal.TestMock{TestName: "mock", SkipGoexit: true}
		e := New(tMock)

		rndFix := func(e Env) int {
			return e.CacheResult(func() (*Result, error) {
				return NewResult(rand.Int()), nil
			}).(int)
		}
		first := rndFix(e)
		second := rndFix(e)

		at.Equal(first, second)
	})
	t.Run("Options", func(t *testing.T) {
		at := assert.New(t)
		tMock := &internal.TestMock{TestName: "mock", SkipGoexit: true}
		e := New(tMock)

		rndFix := func(e Env, name string) int {
			return e.CacheResult(func() (*Result, error) {
				return NewResult(rand.Int()), nil
			}, CacheOptions{CacheKey: name}).(int)
		}
		first1 := rndFix(e, "first")
		first2 := rndFix(e, "first")
		second1 := rndFix(e, "second")
		second2 := rndFix(e, "second")

		at.Equal(first1, first2)
		at.Equal(second1, second2)
		at.NotEqual(first1, second1)
	})
	t.Run("WithCleanup", func(t *testing.T) {
		tMock := &internal.TestMock{TestName: t.Name()}
		env := newTestEnv(tMock)

		callbackCalled := 0
		cleanupCalled := 0
		var callbackFunc FixtureFunction = func() (*Result, error) {
			callbackCalled++
			cleanup := func() {
				cleanupCalled++
			}
			return NewResultWithCleanup(callbackCalled, cleanup), nil
		}

		res := env.CacheResult(callbackFunc)
		require.Equal(t, 1, res)
		require.Equal(t, 1, callbackCalled)
		require.Equal(t, cleanupCalled, 0)

		// got value from cache
		res = env.CacheResult(callbackFunc)
		require.Equal(t, 1, res)
		require.Equal(t, 1, callbackCalled)
		require.Equal(t, cleanupCalled, 0)

		tMock.CallCleanup()
		require.Equal(t, 1, callbackCalled)
		require.Equal(t, 1, cleanupCalled)
	})
	t.Run("Panic", func(t *testing.T) {
		at := assert.New(t)
		tMock := &internal.TestMock{TestName: "mock", SkipGoexit: true}
		e := New(tMock)

		rndFix := func(e Env, name string) int {
			return e.CacheResult(func() (*Result, error) {
				return NewResult(rand.Int()), nil
			}, CacheOptions{CacheKey: name}, CacheOptions{CacheKey: name}).(int)
		}
		at.Panics(func() {
			rndFix(e, "first")
		})
	})
	t.Run("WithNilResult", func(t *testing.T) {
		at := assert.New(t)
		tMock := &internal.TestMock{TestName: "mock"}
		e := newTestEnv(tMock)

		testErr := errors.New("test err")

		failedFix := func(e Env) int {
			return e.CacheResult(func() (*Result, error) {
				return nil, testErr
			}).(int)
		}
		done := make(chan bool)
		go func() {
			defer close(done)
			failedFix(e)
		}()
		<-done
		at.Equal(1, len(tMock.Fatals))
	})
}

func Test_FixtureWrapper(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		at := assert.New(t)

		tMock := &internal.TestMock{TestName: "mock"}
		defer tMock.CallCleanup()

		e := newTestEnv(tMock)
		key := cacheKey("asd")

		cnt := 0
		w := e.fixtureCallWrapper(key, func() (res interface{}, err error) {
			cnt++
			return cnt, errors.New("test")
		}, &FixtureOptions{})
		si := e.scopes[makeScopeName(tMock.Name(), ScopeTest)]
		at.Equal(0, cnt)
		at.Len(si.cacheKeys, 0)
		res1, err := w()
		at.Equal(1, res1)
		at.EqualError(err, "test")
		at.Equal(1, cnt)
		at.Equal([]cacheKey{key}, si.cacheKeys)

		cnt = 0
		key2 := cacheKey("asd")
		cleanupsLen := len(tMock.Cleanups)
		w = e.fixtureCallWrapper(key2, func() (res interface{}, err error) {
			cnt++
			return cnt, nil
		}, &FixtureOptions{cleanupFunc: func() {

		}})
		at.Len(tMock.Cleanups, cleanupsLen)
		_, _ = w()
		at.Equal([]cacheKey{key, key2}, si.cacheKeys)
		at.Len(tMock.Cleanups, cleanupsLen+1)
	})

	t.Run("unknown_scope_info", func(t *testing.T) {
		at := assert.New(t)

		tMock := &internal.TestMock{TestName: "mock"}
		defer tMock.CallCleanup()
		e := newTestEnv(tMock)

		tMock.TestName = "mock2"
		w := e.fixtureCallWrapper("asd", func() (res interface{}, err error) {
			return nil, nil
		}, &FixtureOptions{})
		runUntilFatal(func() {
			_, _ = w()
		})

		// revert test name for good cleanup
		tMock.TestName = "mock"
		at.Len(tMock.Fatals, 1)
	})
}

func Test_Env_Skip(t *testing.T) {
	at := assert.New(t)
	tm := &internal.TestMock{TestName: "mock"}
	tEnv := newTestEnv(tm)

	skipFixtureCallTimes := 0
	skipFixture := func() int {
		res := tEnv.Cache(nil, nil, func() (res interface{}, err error) {
			skipFixtureCallTimes++
			return nil, ErrSkipTest
		})
		return res.(int)
	}

	assertGoExit := func(callback func()) {
		var wg sync.WaitGroup
		wg.Add(1)

		// run in separate goroutine for prevent exit current goroutine
		go func() {
			callbackExited := true
			defer func() {
				at.True(callbackExited)
				panicValue := recover()

				// no panic value (go exit)
				at.Nil(panicValue)
				wg.Done()
			}()

			callback()
			callbackExited = false
		}()

		wg.Wait()
	}

	// skip first time - with call fixture
	executionStarted := false
	executionStopped := true
	assertGoExit(func() {
		executionStarted = true
		skipFixture()

		executionStopped = false
	})

	at.True(executionStarted)
	at.True(executionStopped)
	at.Equal(1, skipFixtureCallTimes)

	// skip second time, without call fixture - from cache
	executionStarted = false
	executionStopped = true
	assertGoExit(func() {
		executionStarted = true
		skipFixture()

		executionStopped = false
	})

	at.True(executionStarted)
	at.True(executionStopped)
	at.Equal(1, skipFixtureCallTimes)

}

func Test_Env_T(t *testing.T) {
	at := assert.New(t)
	e := New(t)
	at.Equal(t, e.T())
}

func Test_Env_TearDown(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		at := assert.New(t)

		t1 := &internal.TestMock{TestName: "mock"}
		// defer t1.callCleanup - direct call e1.tearDown - for test

		e1 := newTestEnv(t1)
		at.Len(e1.scopes, 1)
		at.Len(e1.scopes[makeScopeName(t1.TestName, ScopeTest)].Keys(), 0)
		at.Len(e1.c.store, 0)

		e1.Cache(1, nil, func() (res interface{}, err error) {
			return nil, nil
		})
		e1.Cache(2, nil, func() (res interface{}, err error) {
			return nil, nil
		})
		at.Len(e1.scopes, 1)
		at.Len(e1.scopes[makeScopeName(t1.TestName, ScopeTest)].Keys(), 2)
		at.Len(e1.c.store, 2)

		t2 := &internal.TestMock{TestName: "mock2"}
		// defer t2.callCleanup - direct call e2.tearDown - for test

		e2 := e1.cloneWithTest(t2)
		at.Len(e1.scopes, 2)
		at.Len(e1.scopes[makeScopeName(t1.TestName, ScopeTest)].Keys(), 2)
		at.Len(e1.scopes[makeScopeName(t2.TestName, ScopeTest)].Keys(), 0)
		at.Len(e1.c.store, 2)

		e2.Cache(1, nil, func() (res interface{}, err error) {
			return nil, nil
		})

		at.Len(e1.scopes, 2)
		at.Len(e1.scopes[makeScopeName(t1.TestName, ScopeTest)].Keys(), 2)
		at.Len(e1.scopes[makeScopeName(t2.TestName, ScopeTest)].Keys(), 1)
		at.Len(e1.c.store, 3)

		// finish first test and tearDown e1
		e1.tearDown()
		at.Len(e1.scopes, 1)
		at.Len(e1.scopes[makeScopeName(t2.TestName, ScopeTest)].Keys(), 1)
		at.Len(e1.c.store, 1)

		e2.tearDown()
		at.Len(e1.scopes, 0)
		at.Len(e1.c.store, 0)
	})

	t.Run("tearDown on unexisted scope", func(t *testing.T) {
		at := assert.New(t)
		tMock := &internal.TestMock{TestName: "mock"}
		// defer tMock.callCleanups. e.tearDown will call directly for test
		e := newTestEnv(tMock)

		for key := range e.scopes {
			delete(e.scopes, key)
		}

		runUntilFatal(e.tearDown)

		at.Len(tMock.Fatals, 1)
	})
}

func Test_MakeCacheKey(t *testing.T) {
	at := assert.New(t)

	var res cacheKey
	var err error

	privateEnvFunc := func() {
		res, err = makeCacheKey("asdf", 222, globalEmptyFixtureOptions, true)
	}

	publicEnvFunc := func() {
		privateEnvFunc()
	}
	publicEnvFunc() // external caller
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
				at.Equal(c.result, makeScopeName(c.testName, c.scope))
			})
		}
	})

	t.Run("unexpected_scope", func(t *testing.T) {
		at := assert.New(t)
		at.Panics(func() {
			makeScopeName("asd", -1)
		})
	})
}

func TestNewEnv(t *testing.T) {
	tm := &internal.TestMock{}
	tm.SkipGoexit = true
	New(tm)

	NewEnv(tm)
	if len(tm.Fatals) == 0 {
		t.Fatal("bad double login between new and NewEnv")
	}
}

func runUntilFatal(f func()) {
	stopped := make(chan bool)
	go func() {
		defer func() {
			_ = recover()
			close(stopped)
		}()
		f()
	}()
	<-stopped
}
