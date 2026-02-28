package api

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	"sealboard/internal/util"
	
)

// PodmanStatsCache encapsulates stats and locking
 type PodmanStatsCache struct {
	mu    sync.RWMutex
	stats map[string][]byte // key: host label, value: raw json
}

func NewPodmanStatsCache() *PodmanStatsCache {
	return &PodmanStatsCache{stats: make(map[string][]byte)}
}

func (c *PodmanStatsCache) Set(label string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats[label] = data
}

func (c *PodmanStatsCache) Get(label string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.stats[label]
	return data, ok
}

func (c *PodmanStatsCache) Range(fn func(label string, data []byte)) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for label, data := range c.stats {
		fn(label, data)
	}
}

// callPodmanAPIUnix queries the Podman API over Unix socket
func CallPodmanAPIUnix(socketPath, label string, statsCache *PodmanStatsCache) {
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}
	client := &http.Client{Transport: transport}
	url := "http://d/v4.0.0/containers/json?all=true" // The host part is ignored for UNIX sockets
	resp, err := client.Get(url)
	if err != nil {
		util.LogError("Failed to request Podman API (unix socket) at %s: %v", label, err)
		statsCache.Set(label, util.APIErrorArray(label, fmt.Sprintf("Failed to request Podman API (unix socket): %v", err)))
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.LogError("Failed to read UNIX Podman API response at %s: %v", label, err)
		return
	}
	statsCache.Set(label, body)
	util.LogInfo("[%s] Podman stats cached, %d bytes", label, len(body))
}
