package main

import (
	"io"
	"net"
)

// proxyConn copies traffic bidirectionally between local and remote connections
func proxyConn(local net.Conn, remote net.Conn) {
	go func() { _, _ = io.Copy(local, remote); local.Close() }()
	go func() { _, _ = io.Copy(remote, local); remote.Close() }()
}

