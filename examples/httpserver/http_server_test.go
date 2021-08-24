package httpserver

import (
	"fixenv"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

var (
	globalCounter = 1
)

// pingServer is simple fixture
func pingServer(env fixenv.Env) *httptest.Server {
	var server *httptest.Server

	return env.Cache(nil, func() (res interface{}, err error) {
		mux := &http.ServeMux{}
		mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write([]byte("OK" + strconv.Itoa(globalCounter)))
			globalCounter++
		})
		server = httptest.NewServer(mux)
		return server, nil
	}, &fixenv.FixtureOptions{TearDown: func() {
		server.Close()
	}}).(*httptest.Server)
}

func TestPingServer(t *testing.T) {
	e := fixenv.NewEnv(t)

	s := pingServer(e)

	res, err := http.Get(s.URL)
	if err != nil {
		t.Error()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error()
	}
	_ = res.Body.Close()
	if string(body) != "OK1" {
		t.Error()
	}

	var subTestServer *httptest.Server
	t.Run("subtest", func(t *testing.T) {
		e := fixenv.NewEnv(t)
		subTestServer = pingServer(e)

		res, err := http.Get(subTestServer.URL)
		if err != nil {
			t.Error()
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Error()
		}
		_ = res.Body.Close()
		if string(body) != "OK2" {
			t.Error()
		}
	})

	// server from subtest stopped by teardown fixture
	_, err = http.Get(subTestServer.URL)
	if err == nil {
		t.Error()
	}
}
