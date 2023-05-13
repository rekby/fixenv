package sf

import (
	"github.com/rekby/fixenv"
	"net/http/httptest"
)

func HTTPServer(e fixenv.Env) *httptest.Server {
	return e.CacheWithCleanup(nil, nil, func() (res interface{}, cleanup fixenv.FixtureCleanupFunc, err error) {
		server := httptest.NewServer(nil)
		return server, server.Close, nil
	}).(*httptest.Server)
}
