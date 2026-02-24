package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func pollHosts(hosts []HostConfig) {
	for _, host := range hosts {
		fmt.Printf("\n--- Connecting to %s ---\n", host.Name)
		if host.Address == "localhost" || strings.HasPrefix(host.Address, "127.") {
			if host.SocketPath != "" {
				sp := host.SocketPath
				if strings.Contains(sp, "${UID}") {
					uid := os.Getuid()
					sp = strings.ReplaceAll(sp, "${UID}", fmt.Sprintf("%d", uid))
				}
				callPodmanAPIUnix(sp, "/v4.0.0/containers/json?all=true", host.Name)
			}
			continue
		}
		key, err := os.ReadFile(os.ExpandEnv(host.PrivateKeyPath))
		if err != nil {
			log.Printf("Unable to read private key for %s: %v", host.Name, err)
			continue
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Printf("Unable to parse private key for %s: %v", host.Name, err)
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
			log.Printf("Failed to dial SSH for %s: %v", host.Name, err)
			continue
		}
		defer sshConn.Close()
		if host.LocalSocketPath != "" && host.RemoteSocketPath != "" {
			lsp := host.LocalSocketPath
			rsp := host.RemoteSocketPath
			if strings.Contains(rsp, "${UID}") {
				uid := os.Getuid()
				rsp = strings.ReplaceAll(rsp, "${UID}", fmt.Sprintf("%d", uid))
			}
			_ = os.Remove(lsp)
			listener, err := net.Listen("unix", lsp)
			if err != nil {
				log.Printf("Failed to listen on local unix socket for %s: %v", host.Name, err)
				continue
			}
			defer listener.Close()
			go func() {
				for {
					local, err := listener.Accept()
					if err != nil {
						if !strings.Contains(err.Error(), "use of closed network connection") {
							log.Println("Local unix tunnel error:", err)
						}
						continue
					}
					remote, err := sshConn.Dial("unix", rsp)
					if err != nil {
						log.Println("Remote unix tunnel error:", err)
						local.Close()
						continue
					}
					go proxyConn(local, remote)
				}
			}()
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
	log.Printf("Polling hosts every %d seconds. Press Ctrl+C to exit.", interval)

	// Start the lightweight HTTP stats server restricted to allowed hosts
	StartStatsServer(allowedHTTPHosts, func() interface{} {
		// TODO: Replace this with correct stats, for now just report host names
		return map[string]interface{}{
			"hosts": hosts,
		}
	})

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		pollHosts(hosts)
		<-ticker.C
	}
}
