package server

import (
	"net"
	"time"
)

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe, ListenAndServeTLS and
// ListenAndServeTLSEmbed so dead TCP connections (e.g. closing laptop mid-download)
// eventually go away.
type tcpKeepaliveListener struct {
	*net.TCPListener
	keepalivePeriod time.Duration
}

func (ln tcpKeepaliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	if ln.keepalivePeriod > 0 {
		tc.SetKeepAlivePeriod(ln.keepalivePeriod)
	}
	return tc, nil
}
