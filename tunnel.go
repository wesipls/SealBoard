package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"strings"
)

// proxyConn copies traffic bidirectionally between local and remote connections
func proxyConn(local net.Conn, remote net.Conn) {
	go func() { _, _ = io.Copy(local, remote); local.Close() }()
	go func() { _, _ = io.Copy(remote, local); remote.Close() }()
}

// SetupTunnels creates and runs all required SSH+unix tunnels for remote hosts at startup

func SetupTunnels(hosts []HostConfig) {
	for _, host := range hosts {
		if host.LocalSocketPath != "" && host.RemoteSocketPath != "" {
			lsp := host.LocalSocketPath
			rsp := host.RemoteSocketPath
			rsp = expandUIDVariable(rsp)
			_ = os.Remove(lsp) // Remove any old socket file
			key, err := os.ReadFile(os.ExpandEnv(host.PrivateKeyPath))
			if err != nil {
				LogError("Unable to read private key for %s: %v", host.Name, err)
				continue
			}
			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				LogError("Unable to parse private key for %s: %v", host.Name, err)
				continue
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
				LogError("Failed to dial SSH for %s: %v", host.Name, err)
				continue
			}
			listener, err := net.Listen("unix", lsp)
			if err != nil {
				LogError("Failed to listen on local unix socket for %s: %v", host.Name, err)
				continue
			}
			go func(l, r string, listen net.Listener, sshConn *ssh.Client, hn string) {
				for {
					local, err := listen.Accept()
					if err != nil {
						if !strings.Contains(err.Error(), "use of closed network connection") {
							LogError("Local unix tunnel error (%s): %v", hn, err)
						}
						continue
					}
					remote, err := sshConn.Dial("unix", r)
					if err != nil {
						LogError("Remote unix tunnel error (%s): %v", hn, err)
						local.Close()
						continue
					}
					go proxyConn(local, remote)
				}
			}(lsp, rsp, listener, sshConn, host.Name)
		}
	}
}

