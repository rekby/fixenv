package sf

import (
	"context"
	"github.com/rekby/fixenv"
	"github.com/rekby/fixenv/internal"
	"testing"
)

func TestContext(t *testing.T) {
	tm := &internal.TestMock{}

	e := fixenv.New(tm)
	ctx := Context(e)
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}

	tm.CallCleanup()
	if ctx.Err() != context.Canceled {
		t.Fatal(ctx.Err())
	}
}
