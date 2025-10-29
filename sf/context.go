package sf

import (
	"context"
	"github.com/rekby/fixenv"
)

func Context(e fixenv.Env) context.Context {
	f := func() (*fixenv.GenericResult[context.Context], error) {
		ctx, ctxCancel := context.WithCancel(context.Background())
		return fixenv.NewGenericResultWithCleanup(ctx, fixenv.FixtureCleanupFunc(ctxCancel)), nil
	}
	return fixenv.CacheResult(e, f)
}
