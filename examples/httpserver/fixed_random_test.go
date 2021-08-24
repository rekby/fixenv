package httpserver

import (
	"fixenv"
	"math/rand"
	"testing"
)

func fixedRandom(e fixenv.Env) int {
	return e.Cache(nil, func() (res interface{}, err error) {
		return rand.Int(), nil
	}, nil).(int)
}

func TestFixedRandom(t *testing.T) {
	e := fixenv.NewEnv(t)
	num1 := fixedRandom(e)
	num2 := fixedRandom(e)
	if num1 != num2 {
		t.Error()
	}

	t.Run("other_cache_scope", func(t *testing.T) {
		e := fixenv.NewEnv(t)
		numSub1 := fixedRandom(e)
		numSub2 := fixedRandom(e)
		if num1 == numSub1 {
			t.Error()
		}
		if numSub1 != numSub2 {
			t.Error()
		}
	})
}
