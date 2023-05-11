package sf

import (
	"github.com/rekby/fixenv"
	"os"
)

func TempDir(e fixenv.Env) string {
	return e.CacheWithCleanup(nil, nil, func() (res interface{}, cleanup fixenv.FixtureCleanupFunc, err error) {
		dir, err := os.MkdirTemp("", "")
		mustNoErr(e, err, "failed to create temp dir: %v", err)
		e.T().Logf("Temp dir created: %v", dir)
		clean := func() {
			_ = os.RemoveAll(dir)
			e.T().Logf("Temp dir removed: %v", dir)
		}
		return dir, clean, nil
	}).(string)
}
