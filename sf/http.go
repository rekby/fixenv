package sf

import (
	"github.com/rekby/fixenv"
	"net/http/httptest"
)

func HTTPServer(e fixenv.Env) *httptest.Server {
	f := func() (*fixenv.GenericResult[*httptest.Server], error) {
		server := httptest.NewServer(nil)
		return fixenv.NewGenericResultWithCleanup(server, server.Close), nil
	}
	return fixenv.CacheResult(e, f)
}
