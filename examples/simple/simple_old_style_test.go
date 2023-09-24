package simple

import (
	"testing"

	"github.com/rekby/fixenv"
)

var (
	globalCounter               = 0
	globalTestAndSubtestCounter = 0
)

// counterOldStyle fixture - increment globalCounter every non cached call
// and return new globalCounter value
func counterOldStyle(e fixenv.Env) int {
	f := func() (*fixenv.Result, error) {
		globalCounter++
		return fixenv.NewResult(globalCounter), nil
	}
	return e.CacheResult(f).(int)
}

func TestCounterOldStyle(t *testing.T) {
	e := fixenv.New(t)

	r1 := counterOldStyle(e)
	r2 := counterOldStyle(e)
	if r1 != r2 {
		t.Error()
	}

	t.Run("subtest", func(t *testing.T) {
		e := fixenv.New(t)
		r3 := counterOldStyle(e)
		if r3 == r1 {
			t.Error()
		}
	})
}

// counterTestAndSubtestOldStyle increment counterOldStyle every non cached call
// and cache result for top level test and all of subtests
func counterTestAndSubtestOldStyle(e fixenv.Env) int {
	return e.Cache(nil, &fixenv.FixtureOptions{
		Scope: fixenv.ScopeTestAndSubtests,
	}, func() (res interface{}, err error) {
		globalTestAndSubtestCounter++
		return globalTestAndSubtestCounter, nil
	}).(int)
}

func TestTestAndSubtestCounterOldStyle(t *testing.T) {
	e := fixenv.New(t)

	r1 := counterTestAndSubtestOldStyle(e)
	r2 := counterTestAndSubtestOldStyle(e)
	if r1 != r2 {
		t.Error()
	}

	t.Run("subtest", func(t *testing.T) {
		e := fixenv.New(t)
		r3 := counterTestAndSubtestOldStyle(e)
		if r3 != r1 {
			t.Error()
		}
	})
}
