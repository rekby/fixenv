//go:build go1.18
// +build go1.18

package simple

import (
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"

	"github.com/rekby/fixenv"
)

// namedRandom return random number for new name args
// but return same value for all calls with same names
func namedRandom(e fixenv.Env, name string) int {
	f := func() (*fixenv.GenericResult[int], error) {
		return fixenv.NewGenericResult(rand.Int()), nil
	}

	return fixenv.CacheResult(e, f, fixenv.CacheOptions{CacheKey: name})
}

func TestNamedRandom(t *testing.T) {
	e := fixenv.New(t)
	first := namedRandom(e, "first")
	second := namedRandom(e, "second")
	require.NotEqual(t, first, second)
	require.Equal(t, first, namedRandom(e, "first"))
	require.Equal(t, second, namedRandom(e, "second"))
}
