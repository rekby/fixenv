package sf

import (
	"github.com/rekby/fixenv"
	"net/http/httptest"
)

func HTTPServer(e fixenv.Env) *httptest.Server {
	return e.CacheResult(nil, func() fixenv.Result {
		server := httptest.NewServer(nil)
		return fixenv.Result{
			Result:  server,
			Cleanup: server.Close,
		}
	}).(*httptest.Server)
}
