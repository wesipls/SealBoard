package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// expandUIDVariable replaces ${UID} with the current user's UID
func expandUIDVariable(path string) string {
	if strings.Contains(path, "${UID}") {
		uid := os.Getuid()
		return strings.ReplaceAll(path, "${UID}", fmt.Sprintf("%d", uid))
	}
	return path
}

// setupTunnels creates and runs all required SSH+unix tunnels for remote hosts at startup
func setupTunnels(hosts []HostConfig) {
	for _, host := range hosts {
		if host.LocalSocketPath != "" && host.RemoteSocketPath != "" {
			lsp := host.LocalSocketPath
			rsp := host.RemoteSocketPath
			if strings.Contains(rsp, "${UID}") {
				uid := os.Getuid()
				rsp = strings.ReplaceAll(rsp, "${UID}", fmt.Sprintf("%d", uid))
			}
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

func pollHosts(hosts []HostConfig) {
	for _, host := range hosts {
		LogInfo("--- Connecting to %s ---", host.Name)
		if host.Address == "localhost" || strings.HasPrefix(host.Address, "127.") {
			if host.SocketPath != "" {
				sp := host.SocketPath
				sp = expandUIDVariable(sp)
				callPodmanAPIUnix(sp, "/v4.0.0/containers/json?all=true", host.Name)
			}
			continue
		}
		if host.LocalSocketPath != "" && host.RemoteSocketPath != "" {
			lsp := host.LocalSocketPath
			callPodmanAPIUnix(lsp, "/v4.0.0/containers/json?all=true", host.Name)
		}
	}
}

func main() {
	hosts, globalInterval, allowedHTTPHosts, err := loadConfig("seals.cfg")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	interval := globalInterval
	LogInfo("Polling hosts every %d seconds. Press Ctrl+C to exit.", interval)

	// Set up all required SSH+unix tunnels just once at startup
	setupTunnels(hosts)

	// Start the lightweight HTTP stats server restricted to allowed hosts
	StartStatsServer(allowedHTTPHosts, func() interface{} {
		// Serve latest cached Podman data per host
		podmanStatsMu.RLock()
		defer podmanStatsMu.RUnlock()
		result := make(map[string]interface{})
		for label, data := range podmanStats {
			var parsed interface{}
			if err := json.Unmarshal(data, &parsed); err == nil {
				result[label] = parsed
			} else {
				// If parsing fails, emit a standard error array for this label
				errmsg := FormatErrorMsg("Internal stats/cache error for %s: %v", label, err)
				result[label] = json.RawMessage(APIErrorArray(label, errmsg))
			}
		}
		return result
	})

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		pollHosts(hosts)
		<-ticker.C
	}
}
