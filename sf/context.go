package sf

import (
	"context"
	"github.com/rekby/fixenv"
)

func Context(e fixenv.Env) context.Context {
	return e.CacheResult(nil, func() fixenv.Result {
		ctx, ctxCancel := context.WithCancel(context.Background())
		return fixenv.Result{
			Result:  ctx,
			Error:   nil,
			Cleanup: fixenv.FixtureCleanupFunc(ctxCancel),
		}
	}).(context.Context)
}
