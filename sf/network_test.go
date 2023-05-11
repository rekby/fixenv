package sf

import (
	"github.com/rekby/fixenv"
	"net"
	"testing"
)

func TestFreeLocalTcpAddress(t *testing.T) {
	e := fixenv.New(t)
	addr := FreeLocalTCPAddress(e)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	_ = listener.Close()
}

func TestLocalTcpListener(t *testing.T) {
	e := fixenv.New(t)
	listener := LocalTCPListener(e)
	conn, err := net.Dial("tcp", listener.Addr().String())
	mustNoErr(e, err, "failed connect to listener: %v", err)
	_ = conn.Close()
}
