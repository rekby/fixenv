package sf

import (
	"github.com/rekby/fixenv"
	"net"
)

func FreeLocalTCPAddress(e fixenv.Env) string {
	return FreeLocalTCPAddressNamed(e, "")
}

func FreeLocalTCPAddressNamed(e fixenv.Env, name string) string {
	f := func() (*fixenv.GenericResult[string], error) {
		listener := LocalTCPListenerNamed(e, "FreeLocalTCPAddressNamed-"+name)
		addr := listener.Addr().String()
		err := listener.Close()
		mustNoErr(e, err, "failed to close temp listener: %v", err)
		return fixenv.NewGenericResult(addr), nil
	}
	return fixenv.CacheResult(e, f, fixenv.CacheOptions{CacheKey: name})
}

func LocalTCPListener(e fixenv.Env) *net.TCPListener {
	return LocalTCPListenerNamed(e, "")
}

func LocalTCPListenerNamed(e fixenv.Env, name string) *net.TCPListener {
	f := func() (*fixenv.GenericResult[*net.TCPListener], error) {
		listener, err := net.Listen("tcp", "localhost:0")
		clean := func() {
			if listener != nil {
				_ = listener.Close()
			}
		}
		if err != nil {
			return nil, err
		}
		tcpListener := listener.(*net.TCPListener)
		return fixenv.NewGenericResultWithCleanup(tcpListener, clean), nil
	}
	return fixenv.CacheResult(e, f, fixenv.CacheOptions{CacheKey: name})
}
