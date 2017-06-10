package main

import (
	"net"
	"net/url"
	"time"
)

func MustParseUrl(raw string) *url.URL {
	r, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}

	return r
}

type KeepAliveListener struct {
	*net.TCPListener

	KeepAlive time.Duration
}

func (listener KeepAliveListener) Accept() (c net.Conn, err error) {
	tcpConn, err := listener.AcceptTCP()
	if err != nil {
		return
	}

	tcpConn.SetKeepAlive(true)
	tcpConn.SetKeepAlivePeriod(listener.KeepAlive)
	return tcpConn, nil
}
