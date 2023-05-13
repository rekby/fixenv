package sf

import (
	"github.com/rekby/fixenv"
	"io"
	"net/http"
	"testing"
)

func TestHttpServer(t *testing.T) {
	e := fixenv.New(t)
	server := HTTPServer(e)
	server.Config.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = io.WriteString(writer, "OK")
	})
	resp, err := http.Get(server.URL)
	mustNoErr(e, err, "failed to get url: %v", err)
	content, err := io.ReadAll(resp.Body)
	mustNoErr(e, err, "failed to read content: %v", err)
	if string(content) != "OK" {
		t.Fatal(string(content))
	}
}
