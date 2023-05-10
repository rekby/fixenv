//go:build go1.18
// +build go1.18

package simple

import (
	"math/rand"
	"testing"

	"github.com/rekby/fixenv"
	"github.com/stretchr/testify/require"
)

// namedRandom return random number for new name args
// but return same value for all calls with same names
func namedRandom(e fixenv.Env, name string) int {
	return fixenv.Cache(e, name, nil, func() (res int, err error) {
		return rand.Int(), nil
	})
}

func TestNamedRandom(t *testing.T) {
	e := fixenv.New(t)
	first := namedRandom(e, "first")
	second := namedRandom(e, "second")
	require.NotEqual(t, first, second)
	require.Equal(t, first, namedRandom(e, "first"))
	require.Equal(t, second, namedRandom(e, "second"))
}
