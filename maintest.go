package fixenv

import (
	"fmt"
	"sync"
)

type FatalfFunction func(format string, args ...interface{})

type CreateMainTestEnvOpts struct {
	// Fatalf equivalent of Fatalf in test.
	// Must write log, then exit from goroutine.
	// It may be panic.
	// Fatalf called if main envinment can't continue work
	Fatalf FatalfFunction
}

// CreateMainTestEnv called from TestMain for create global environment.
// It need only for use ScopePackage cache scope.
// If ScopePackage not used - no need to create main env.
func CreateMainTestEnv(opts *CreateMainTestEnvOpts) (env *EnvT, tearDown func()) {
	globalMutex.Lock()
	packageLevelVirtualTest := newVirtualTest(opts)
	globalMutex.Unlock()

	env = NewEnv(packageLevelVirtualTest) // register global test for env
	return env, packageLevelVirtualTest.cleanup
}

// virtualTest implement T interface for global env scope
type virtualTest struct {
	m        sync.Mutex
	fatalf   FatalfFunction
	cleanups []func()
}

func newVirtualTest(opts *CreateMainTestEnvOpts) *virtualTest {
	if opts == nil {
		opts = &CreateMainTestEnvOpts{}
	}
	t := &virtualTest{
		fatalf: opts.Fatalf,
	}

	if opts.Fatalf == nil {
		t.fatalf = func(format string, args ...interface{}) {
			panic(fmt.Sprintf(format, args...))
		}
	}

	return t
}

func (t *virtualTest) Cleanup(f func()) {
	t.m.Lock()
	defer t.m.Unlock()

	t.cleanups = append(t.cleanups, f)
}

func (t *virtualTest) Fatalf(format string, args ...interface{}) {
	t.fatalf(format, args...)
}

func (t *virtualTest) Name() string {
	return packageScopeName
}

func (t *virtualTest) cleanup() {
	for i := len(t.cleanups) - 1; i >= 0; i-- {
		t.cleanups[i]()
	}
}
