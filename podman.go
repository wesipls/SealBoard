package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
)

// PodmanStats holds the latest stats/result per host
var (
	podmanStatsMu sync.RWMutex
	podmanStats   = map[string][]byte{} // key: host label, value: raw json
)

// callPodmanAPIUnix queries the Podman API over Unix socket
func callPodmanAPIUnix(socketPath, label string) {
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}
	client := &http.Client{Transport: transport}
	url := "http://d/v4.0.0/containers/json?all=true" // The host part is ignored for UNIX sockets
	resp, err := client.Get(url)
	if err != nil {
		LogError("Failed to request Podman API (unix socket) at %s: %v", label, err)
		podmanStatsMu.Lock()
		// Use errorutil.go to standardize error array
		podmanStats[label] = APIErrorArray(label, fmt.Sprintf("Failed to request Podman API (unix socket): %v", err))
		podmanStatsMu.Unlock()
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		LogError("Failed to read UNIX Podman API response at %s: %v", label, err)
		return
	}
	podmanStatsMu.Lock()
	podmanStats[label] = body
	podmanStatsMu.Unlock()
	LogInfo("[%s] Podman stats cached, %d bytes", label, len(body))
}
