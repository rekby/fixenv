package fixenv

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMainTestEnv(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		at := assert.New(t)
		e, cancel := CreateMainTestEnv(nil)
		at.Equal(packageScopeName, e.t.Name())
		at.NotNil(globalScopeInfo[packageScopeName])
		cancel()
		at.Nil(globalScopeInfo[packageScopeName])
	})

	t.Run("fatal_as_panic", func(t *testing.T) {
		at := assert.New(t)
		e, cancel := CreateMainTestEnv(nil)
		defer cancel()

		at.Panics(func() {
			e.T().Fatalf("asd")
		})
	})

	t.Run("opt_fatal", func(t *testing.T) {
		at := assert.New(t)
		var fFormat string
		var fArgs []interface{}
		e, cancel := CreateMainTestEnv(&CreateMainTestEnvOpts{Fatalf: func(format string, args ...interface{}) {
			fFormat = format
			fArgs = args
		}})
		defer cancel()

		e.T().Fatalf("asd", 1, 2, 3)
		at.Equal("asd", fFormat)
		at.Equal([]interface{}{1, 2, 3}, fArgs)
	})

	t.Run("skip_now", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			at := assert.New(t)
			e, cancel := CreateMainTestEnv(nil)
			defer cancel()

			at.Panics(func() {
				e.T().SkipNow()
			})
			at.True(e.T().Skipped())
		})
		t.Run("opt", func(t *testing.T) {
			at := assert.New(t)
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
			at.Equal(1, skipCalled)
			at.True(e.T().Skipped())
		})
	})
}
