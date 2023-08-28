//go:build go1.18
// +build go1.18

package custom_env_with_shared_content

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rekby/fixenv"

	"github.com/stretchr/testify/assert"
)

type Env struct {
	fixenv.Env
	Resp string
}

func NewEnv(t *testing.T) *Env {
	return &Env{
		Env:  fixenv.New(t),
		Resp: "OK",
	}
}

func testServer(e *Env) *httptest.Server {
	return fixenv.CacheResult(e, nil, func() fixenv.GenericResult[*httptest.Server] {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write([]byte(e.Resp))
		}))
		e.T().(testing.TB).Logf("Http server start, url: %q", server.URL)
		cleanup := func() {
			server.Close()
			e.T().(testing.TB).Logf("Http server stop, url: %q", server.URL)
		}
		return fixenv.GenericResult[*httptest.Server]{Result: server, Cleanup: cleanup}
	})
}

func TestHttpServer(t *testing.T) {
	at := assert.New(t)
	e := NewEnv(t)

	s := testServer(e)

	resp, err := http.Get(s.URL)
	at.NoError(err)
	body, err := io.ReadAll(resp.Body)
	at.NoError(err)
	_ = resp.Body.Close()
	at.Equal("OK", string(body))

	e.Resp = "PONG"
	resp, err = http.Get(s.URL)
	at.NoError(err)
	body, err = io.ReadAll(resp.Body)
	at.NoError(err)
	_ = resp.Body.Close()
	at.Equal("PONG", string(body))
}
