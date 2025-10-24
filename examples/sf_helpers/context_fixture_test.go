package sf_helpers

import (
	"context"
	"testing"
	"time"

	"github.com/rekby/fixenv"
	"github.com/rekby/fixenv/sf"
	"github.com/stretchr/testify/require"
)

func TestContextIsCancelledAfterTest(t *testing.T) {
	var ctx context.Context

	t.Run("use context fixture", func(t *testing.T) {
		e := fixenv.New(t)
		ctx = sf.Context(e)
		require.NoError(t, ctx.Err())
	})

	select {
	case <-ctx.Done():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("context was not cancelled after the test finished")
	}
}
