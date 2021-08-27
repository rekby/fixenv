package simple

import (
	"fixenv"
	"testing"
)

var (
	packageCounterVal = 0
)

func TestMain(m *testing.M) {
	// initialize package env
	_, cancel := fixenv.CreateMainTestEnv(nil)
	defer cancel()

	m.Run()
}

// packageCounter fixture will call without cache once only
func packageCounter(e fixenv.Env) int {
	return e.Cache(nil, &fixenv.FixtureOptions{Scope: fixenv.ScopePackage}, func() (res interface{}, err error) {
		packageCounterVal++
		return packageCounterVal, nil
	}).(int)
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
