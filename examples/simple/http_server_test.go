//go:build go1.18
// +build go1.18

package simple

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rekby/fixenv"

	"github.com/stretchr/testify/assert"
)

func testServer(e fixenv.Env, response string) *httptest.Server {
	return e.CacheWithCleanup(response, nil, func() (res interface{}, cleanup fixenv.FixtureCleanupFunc, err error) {
		resp := []byte(response)

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write(resp)
		}))
		e.T().(testing.TB).Logf("Http server start. %q url: %q", response, server.URL)
		cleanup = func() {
			server.Close()
			e.T().(testing.TB).Logf("Http server stop. %q url: %q", response, server.URL)
		}
		return server, cleanup, nil
	}).(*httptest.Server)
}

func TestHttpServer(t *testing.T) {
	at := assert.New(t)
	e := fixenv.NewEnv(t)

	s1 := testServer(e, "OK")
	resp, err := http.Get(s1.URL)
	at.NoError(err)
	body, err := io.ReadAll(resp.Body)
	at.NoError(err)
	_ = resp.Body.Close()
	at.Equal("OK", string(body))

	s1Same := testServer(e, "OK")
	at.Equal(s1, s1Same)

	s2 := testServer(e, "PONG")
	at.NotEqual(s1, s2)
	resp, err = http.Get(s2.URL)
	at.NoError(err)
	body, err = io.ReadAll(resp.Body)
	at.NoError(err)
	_ = resp.Body.Close()
	at.Equal("PONG", string(body))
}
