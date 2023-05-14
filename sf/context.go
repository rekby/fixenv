package sf

import (
	"context"
	"github.com/rekby/fixenv"
)

func Context(e fixenv.Env) context.Context {
	return e.CacheWithCleanup(nil, nil, func() (res interface{}, _ fixenv.FixtureCleanupFunc, _ error) {
		ctx, ctxCancel := context.WithCancel(context.Background())
		return ctx, fixenv.FixtureCleanupFunc(ctxCancel), nil
	}).(context.Context)
}
