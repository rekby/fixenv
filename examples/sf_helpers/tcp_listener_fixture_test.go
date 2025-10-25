package sf_helpers

import (
	"net"
	"testing"

	"github.com/rekby/fixenv"
	"github.com/rekby/fixenv/sf"
	"github.com/stretchr/testify/require"
)

func TestLocalTCPListenerLifecycle(t *testing.T) {
	var addr string

	t.Run("allocate listener", func(t *testing.T) {
		e := fixenv.New(t)

		listener := sf.LocalTCPListener(e)
		sameListener := sf.LocalTCPListener(e)
		require.Equal(t, listener, sameListener)

		addr = listener.Addr().String()

		t.Run("subtest gets new listener", func(t *testing.T) {
			e := fixenv.New(t)
			childListener := sf.LocalTCPListener(e)
			require.NotEqual(t, addr, childListener.Addr().String())
		})
	})

	ln, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	require.NoError(t, ln.Close())
}
