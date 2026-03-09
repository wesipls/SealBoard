package tunnel

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"strings"
	"net/http"
	"time"
	"context"

	"sealboard/internal/util"
	"sealboard/internal/config"
)

// proxyConn copies traffic bidirectionally between local and remote connections
func proxyConn(local net.Conn, remote net.Conn) {
	go func() { _, _ = io.Copy(local, remote); local.Close() }()
	go func() { _, _ = io.Copy(remote, local); remote.Close() }()
}

// EnsureTunnel checks if the Unix socket for the host exists and is responsive. If not, it recreates the SSH tunnel for the host.
func EnsureTunnel(host config.HostConfig) {
	if host.LocalSocketPath == "" || host.RemoteSocketPath == "" {
		return
	}
	lsp := host.LocalSocketPath
	// Check whether we can connect to the local socket: if not, recreate the tunnel.
	c, err := net.Dial("unix", lsp)
	if err == nil {
			// Try a minimal HTTP GET to ensure the tunnel and Podman API is actually alive
			c.Close() // close immediately to avoid resource leak
			ok := false
			func() {
				defer func() { recover() }()
				transport := &http.Transport{
					DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
						return net.Dial("unix", lsp)
					},
				}
				client := &http.Client{Transport: transport, Timeout: 2 * time.Second}
				rsp, err := client.Get("http://d/v4.0.0/libpod/info")
				if err == nil && rsp.StatusCode >= 200 && rsp.StatusCode < 300 {
					rsp.Body.Close()
					ok = true
				}
			}()
			if ok {
				return // Tunnel actually works for HTTP requests
			}
			// Otherwise fall through and reset tunnel
	}
	// If not, set up the tunnel as in SetupTunnels logic for this host only
	rsp := util.ExpandUIDVariable(host.RemoteSocketPath)
	_ = os.Remove(lsp) // Remove any old socket file
	key, err := os.ReadFile(os.ExpandEnv(host.PrivateKeyPath))
	if err != nil {
		util.LogError("Unable to read private key for %s: %v", host.Name, err)
		return
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		util.LogError("Unable to parse private key for %s: %v", host.Name, err)
		return
	}
	sshPort := 22
	if host.SSHPort != 0 {
		sshPort = host.SSHPort
	}
	sshConfig := &ssh.ClientConfig{
		User:            host.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.Address, sshPort), sshConfig)
	if err != nil {
		util.LogError("Failed to dial SSH for %s: %v", host.Name, err)
		return
	}
	listener, err := net.Listen("unix", lsp)
	if err != nil {
		util.LogError("Failed to listen on local unix socket for %s: %v", host.Name, err)
		return
	}
	go func(l, r string, listen net.Listener, sshConn *ssh.Client, hn string) {
		for {
			local, err := listen.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					util.LogError("Local unix tunnel error (%s): %v", hn, err)
				}
				continue
			}
			remote, err := sshConn.Dial("unix", r)
			if err != nil {
				util.LogError("Remote unix tunnel error (%s): %v", hn, err)
				local.Close()
				continue
			}
			go proxyConn(local, remote)
		}
	}(lsp, rsp, listener, sshConn, host.Name)
}

// SetupTunnels creates and runs all required SSH+unix tunnels for remote hosts at startup
func SetupTunnels(hosts []config.HostConfig) {
	for _, host := range hosts {
		if host.LocalSocketPath != "" && host.RemoteSocketPath != "" {
			EnsureTunnel(host)
		}
	}
}

