package simple

import (
	"testing"

	"github.com/rekby/fixenv"
)

var (
	globalCounter               = 0
	globalTestAndSubtestCounter = 0
)

// counter fixture - increment globalCounter every non cached call
// and return new globalCounter value
func counter(e fixenv.Env) int {
	return e.Cache(nil, nil, func() (res interface{}, err error) {
		globalCounter++
		return globalCounter, nil
	}).(int)
}

func TestCounter(t *testing.T) {
	e := fixenv.NewEnv(t)

	r1 := counter(e)
	r2 := counter(e)
	if r1 != r2 {
		t.Error()
	}

	t.Run("subtest", func(t *testing.T) {
		e := fixenv.NewEnv(t)
		r3 := counter(e)
		if r3 == r1 {
			t.Error()
		}
	})
}

// counterTestAndSubtest increment counter every non cached call
// and cache result for top level test and all of subtests
func counterTestAndSubtest(e fixenv.Env) int {
	return e.Cache(nil, &fixenv.FixtureOptions{
		Scope: fixenv.ScopeTestAndSubtests,
	}, func() (res interface{}, err error) {
		globalTestAndSubtestCounter++
		return globalTestAndSubtestCounter, nil
	}).(int)
}

func TestTestAndSubtestCounter(t *testing.T) {
	e := fixenv.NewEnv(t)

	r1 := counterTestAndSubtest(e)
	r2 := counterTestAndSubtest(e)
	if r1 != r2 {
		t.Error()
	}

	t.Run("subtest", func(t *testing.T) {
		e := fixenv.NewEnv(t)
		r3 := counterTestAndSubtest(e)
		if r3 != r1 {
			t.Error()
		}
	})
}
