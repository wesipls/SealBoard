package main

import (
	"fmt"
	"io"
	"os"
	"log"
	"net"
	"net/http"
	"strings"
	"context"
	"gopkg.in/yaml.v3"
	"golang.org/x/crypto/ssh"
)

type HostConfig struct {
	Name              string `yaml:"name"`
	Address           string `yaml:"address"`
	User              string `yaml:"user"`
	PrivateKeyPath    string `yaml:"private_key_path"`
	SocketPath        string `yaml:"socket_path"`
	RemoteSocketPath  string `yaml:"remote_socket_path"`
	LocalSocketPath   string `yaml:"local_socket_path"`
}

type Config struct {
	Hosts []HostConfig `yaml:"hosts"`
}

func loadConfig(path string) ([]HostConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Config
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return cfg.Hosts, nil
}

func main() {
	hosts, err := loadConfig("seals.cnf")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
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
		sshConfig := &ssh.ClientConfig{
			User: host.User,
			Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		sshConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host.Address), sshConfig)
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
			// Remove any stale local socket
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


func callPodmanAPIUnix(socketPath, apiPath, label string) {
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}
	client := &http.Client{Transport: transport}
	url := "http://d/v4.0.0/containers/json?all=true" // The host part is ignored for UNIX sockets
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Failed to request Podman API (unix socket) at %s: %v", label, err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read UNIX Podman API response at %s: %v", label, err)
		return
	}
	fmt.Printf("[%s] Podman stats raw response: %s\n", label, string(body))
}

func proxyConn(local net.Conn, remote net.Conn) {
	go func() { _, _ = io.Copy(local, remote); local.Close() }()
	go func() { _, _ = io.Copy(remote, local); remote.Close() }()
}

