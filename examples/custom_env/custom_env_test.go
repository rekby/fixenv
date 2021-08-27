package custom_env

import (
	"context"
	"fixenv"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Env struct {
	Ctx context.Context
	fixenv.Env
	*assert.Assertions
}

func NewEnv(t *testing.T) (context.Context, *Env) {
	at := assert.New(t)
	fEnv := fixenv.NewEnv(t)
	ctx, ctxCancel := context.WithCancel(context.Background())
	fEnv.Cleanup(func() {
		ctxCancel()
	})
	res := &Env{
		Ctx:        ctx,
		Env:        fEnv,
		Assertions: at,
	}
	return ctx, res
}

func testServer(e fixenv.Env, response string) *httptest.Server {
	return e.Cache(response, nil, func() (res interface{}, err error) {
		resp := []byte(response)

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write(resp)
		}))
		e.T().(testing.TB).Logf("Http server start. %q url: %q", response, server.URL)
		e.T().Cleanup(func() {
			server.Close()
			e.T().(testing.TB).Logf("Http server stop. %q url: %q", response, server.URL)
		})
		return server, nil
	}).(*httptest.Server)
}

func TestHttpServerSelfEnv(t *testing.T) {
	ctx, e := NewEnv(t)
	s1 := testServer(e, "OK")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s1.URL, nil)
	e.NoError(err)

	resp, err := http.DefaultClient.Do(req)
	e.NoError(err)
	body, err := io.ReadAll(resp.Body)
	e.NoError(err)
	_ = resp.Body.Close()
	e.Equal("OK", string(body))
}
