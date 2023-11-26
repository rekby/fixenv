//go:build go1.18
// +build go1.18

package simple_main_test

import (
	"github.com/rekby/fixenv"
	"math/rand"
	"testing"
)

var global int = -1

func FSingleRandom(e fixenv.Env) int {
	var f fixenv.GenericFixtureFunction[int] = func() (*fixenv.GenericResult[int], error) {
		return fixenv.NewGenericResult(rand.Int()), nil
	}
	return fixenv.CacheResult(e, f, fixenv.CacheOptions{Scope: fixenv.ScopePackage})
}

func TestFirst(t *testing.T) {
	e := fixenv.New(t)
	if global == -1 {
		global = FSingleRandom(e)
	}

	if singleRnd := FSingleRandom(e); singleRnd != global {
		t.Fatalf("%v != %v", singleRnd, global)
	}
}

func TestSecond(t *testing.T) {
	e := fixenv.New(t)
	if global == -1 {
		global = FSingleRandom(e)
	}

	if singleRnd := FSingleRandom(e); singleRnd != global {
		t.Fatalf("%v != %v", singleRnd, global)
	}
}
