//go:build go1.18
// +build go1.18

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
	f := func() (*fixenv.GenericResult[int], error) {
		globalCounter++
		e.T().Logf("increment globalCounter to: ")
		return fixenv.NewGenericResult(globalCounter), nil
	}

	return fixenv.CacheResult(e, f)
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
	return fixenv.CacheResult(e, func() (*fixenv.GenericResult[int], error) {
		globalTestAndSubtestCounter++
		e.T().Logf("increment globalTestAndSubtestCounter to: ")
		return fixenv.NewGenericResult(globalTestAndSubtestCounter), nil
	}, fixenv.CacheOptions{Scope: fixenv.ScopeTestAndSubtests})
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
