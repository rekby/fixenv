package sf

import (
	"context"
	"github.com/rekby/fixenv"
)

func Context(e fixenv.Env) context.Context {
	return fixenv.CacheWithCleanup(e, nil, nil, func() (context.Context, fixenv.FixtureCleanupFunc, error) {
		ctx, ctxCancel := context.WithCancel(context.Background())
		return ctx, fixenv.FixtureCleanupFunc(ctxCancel), nil
	})
}
