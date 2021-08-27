package custom_env_with_shared_content

import (
	"fixenv"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Env struct {
	fixenv.Env
	Resp string
}

func NewEnv(t *testing.T) *Env {
	return &Env{
		Env:  fixenv.NewEnv(t),
		Resp: "OK",
	}
}

func testServer(e *Env) *httptest.Server {
	return e.Cache(nil, nil, func() (res interface{}, err error) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write([]byte(e.Resp))
		}))
		e.T().(testing.TB).Logf("Http server start, url: %q", server.URL)
		e.T().Cleanup(func() {
			server.Close()
			e.T().(testing.TB).Logf("Http server stop, url: %q", server.URL)
		})
		return server, nil
	}).(*httptest.Server)
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
