//go:build go1.18
// +build go1.18

package simple

import (
	"testing"

	"github.com/rekby/fixenv"
)

// counter fixture - increment globalCounter every non cached call
// and return new globalCounter value
func counter(e fixenv.Env) int {
	return fixenv.CacheResult(e, nil, func() fixenv.GenericResult[int] {
		globalCounter++
		e.T().Logf("increment globalCounter to: ")
		return fixenv.GenericResult[int]{Result: globalCounter}
	})
}

func TestCounter(t *testing.T) {
	e := fixenv.New(t)

	r1 := counter(e)
	r2 := counter(e)
	if r1 != r2 {
		t.Error()
	}

	t.Run("subtest", func(t *testing.T) {
		e := fixenv.New(t)
		r3 := counter(e)
		if r3 == r1 {
			t.Error()
		}
	})
}

// counterTestAndSubtest increment counter every non cached call
// and cache result for top level test and all of subtests
func counterTestAndSubtest(e fixenv.Env) int {
	return fixenv.Cache(e, "", &fixenv.FixtureOptions{
		Scope: fixenv.ScopeTestAndSubtests,
	}, func() (res int, err error) {
		globalTestAndSubtestCounter++
		e.T().Logf("increment globalTestAndSubtestCounter to: ")
		return globalTestAndSubtestCounter, nil
	})
}

func TestTestAndSubtestCounter(t *testing.T) {
	e := fixenv.New(t)

	r1 := counterTestAndSubtest(e)
	r2 := counterTestAndSubtest(e)
	if r1 != r2 {
		t.Error()
	}

	t.Run("subtest", func(t *testing.T) {
		e := fixenv.New(t)
		r3 := counterTestAndSubtest(e)
		if r3 != r1 {
			t.Error()
		}
	})
}
