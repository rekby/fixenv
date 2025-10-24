package sf_helpers

import (
	"os"
	"testing"

	"github.com/rekby/fixenv"
	"github.com/rekby/fixenv/sf"
	"github.com/stretchr/testify/require"
)

func TestTempDirRemovedAfterScope(t *testing.T) {
	var dir string

	t.Run("create temp dir", func(t *testing.T) {
		e := fixenv.New(t)
		dir = sf.TempDir(e)
		require.DirExists(t, dir)

		sameDir := sf.TempDir(e)
		require.Equal(t, dir, sameDir)
	})

	_, err := os.Stat(dir)
	require.ErrorIs(t, err, os.ErrNotExist, "temporary directory should be removed after the scope")
}
