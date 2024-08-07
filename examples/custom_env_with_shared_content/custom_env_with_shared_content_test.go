//go:build go1.18
// +build go1.18

package customenv2

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
	f := func() (*fixenv.GenericResult[*httptest.Server], error) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			_, _ = writer.Write([]byte(e.Resp))
		}))
		e.T().(testing.TB).Logf("Http server start, url: %q", server.URL)
		cleanup := func() {
			server.Close()
			e.T().(testing.TB).Logf("Http server stop, url: %q", server.URL)
		}
		return fixenv.NewGenericResultWithCleanup(server, cleanup), nil
	}

	return fixenv.CacheResult(e, f)
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
