package fixenv

import (
	"errors"
	"runtime"
	"sync"
	"testing"
)

func TestCreateMainTestEnv(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		e, cancel := CreateMainTestEnv(nil)
		e.T().Logf("env created")

		requireEquals(t, packageScopeName, e.t.Name())
		requireNotNil(t, globalScopeInfo[packageScopeName])
		cancel()
		requireNil(t, globalScopeInfo[packageScopeName])
	})

	t.Run("fatal_as_panic", func(t *testing.T) {
		e, cancel := CreateMainTestEnv(nil)
		defer cancel()

		requirePanic(t, func() {
			e.T().Fatalf("asd")
		})
	})

	t.Run("opt_fatal", func(t *testing.T) {
		var fFormat string
		var fArgs []interface{}
		e, cancel := CreateMainTestEnv(&CreateMainTestEnvOpts{Fatalf: func(format string, args ...interface{}) {
			fFormat = format
			fArgs = args
		}})
		defer cancel()

		e.T().Fatalf("asd", 1, 2, 3)
		requireEquals(t, "asd", fFormat)
		requireEquals(t, []interface{}{1, 2, 3}, fArgs)
	})

	t.Run("skip_now", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			e, cancel := CreateMainTestEnv(nil)
			defer cancel()

			requirePanic(t, func() {
				e.T().SkipNow()
			})
			requireTrue(t, e.T().Skipped())
		})
		t.Run("opt", func(t *testing.T) {
			skipCalled := 0
			e, cancel := CreateMainTestEnv(&CreateMainTestEnvOpts{SkipNow: func() {
				skipCalled++
				runtime.Goexit()
			}})
			defer cancel()

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				e.T().SkipNow()
			}()
			wg.Wait()
			requireEquals(t, 1, skipCalled)
			requireTrue(t, e.T().Skipped())
		})
	})
}

func TestRunTests(t *testing.T) {
	expectedReturnCode := 123

	checkInitialized := func(t *testing.T) {
		t.Helper()

		globalMutex.Lock()
		defer globalMutex.Unlock()

		if _, ok := globalScopeInfo[packageScopeName]; !ok {
			t.Fatal()
		}
	}
	cleanGlobalState := func() {
		globalMutex.Lock()
		defer globalMutex.Unlock()

		delete(globalScopeInfo, packageScopeName)
	}

	t.Run("without options", func(t *testing.T) {
		m := &mTestsMock{
			returnCode: expectedReturnCode,
			run: func() {
				checkInitialized(t)
			},
		}

		if res := RunTests(m); res != expectedReturnCode {
			t.Fatalf("%v != %v", res, expectedReturnCode)
		}
		cleanGlobalState()
	})
	t.Run("with options", func(t *testing.T) {
		m := &mTestsMock{
			returnCode: expectedReturnCode,
			run: func() {
				checkInitialized(t)
				lastPackageLevelVirtualTest.SkipNow()
			},
		}

		called := false
		RunTests(m, CreateMainTestEnvOpts{SkipNow: func() {
			called = true
		}})
		if !called {
			t.Fatal()
		}
		cleanGlobalState()
	})
	t.Run("with two options", func(t *testing.T) {
		defer func() {
			cleanGlobalState()

			rec := recover()
			if !errors.Is(rec.(error), errTooManyOptionalArgs) {
				t.Fatal(rec)
			}
		}()
		m := &mTestsMock{
			run: func() {
				checkInitialized(t)
			},
		}
		RunTests(m, CreateMainTestEnvOpts{}, CreateMainTestEnvOpts{})
	})
}

type mTestsMock struct {
	runCalled  bool
	returnCode int
	run        func()
}

func (r *mTestsMock) Run() (code int) {
	r.runCalled = true
	if r.run != nil {
		r.run()
	}
	return r.returnCode
}

// check interface implementation
var _ RunTestsI = &mTestsMock{}
