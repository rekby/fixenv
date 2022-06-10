//go:build go1.18
// +build go1.18

package simple

import (
	"os"
	"testing"

	"github.com/rekby/fixenv"
)

var (
	packageCounterVal = 0
)

func TestMain(m *testing.M) {
	var exitCode int

	// initialize package env
	_, cancel := fixenv.CreateMainTestEnv(nil)
	defer func() {
		cancel()
		os.Exit(exitCode)
	}()

	exitCode = m.Run()
}

// packageCounter fixture will call without cache once only
func packageCounter(e fixenv.Env) int {
	return fixenv.Cache(e, "", &fixenv.FixtureOptions{Scope: fixenv.ScopePackage}, func() (res int, err error) {
		packageCounterVal++
		return packageCounterVal, nil
	})
}

func TestPackageFirst(t *testing.T) {
	e := fixenv.NewEnv(t)
	if packageCounter(e) != 1 {
		t.Error()
	}
}

func TestPackageSecond(t *testing.T) {
	e := fixenv.NewEnv(t)
	if packageCounter(e) != 1 {
		t.Error()
	}
}
