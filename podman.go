package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"context"
	"sync"
	"encoding/json"
)


// PodmanStats holds the latest stats/result per host
var (
	podmanStatsMu sync.RWMutex
	podmanStats = map[string][]byte{} // key: host label, value: raw json
)

// callPodmanAPIUnix queries the Podman API over Unix socket
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
		LogError("Failed to request Podman API (unix socket) at %s: %v", label, err)
		podmanStatsMu.Lock()
		// Ensure an array so frontend always renders a row for the host
		errorArr := []map[string]interface{}{{
			"host": label,
			"status": "error",
			"error": fmt.Sprintf("Failed to request Podman API (unix socket): %v", err),
		}}
		errBytes, _ := json.Marshal(errorArr)
		podmanStats[label] = errBytes
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

