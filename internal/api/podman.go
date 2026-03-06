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

// CallPodmanAPIUnixEndpoint queries the Podman API over Unix socket at arbitrary endpoint
func CallPodmanAPIUnixEndpoint(socketPath, label, endpoint string, statsCache *PodmanStatsCache) {
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}
	client := &http.Client{Transport: transport}
	url := endpoint // Fully customizable endpoint (must be complete, host part still ignored)
	resp, err := client.Get(url)
	if err != nil {
		util.LogError("Failed to request Podman API (unix socket) at %s endpoint %s: %v", label, endpoint, err)
		statsCache.Set(label, util.APIErrorArray(label, fmt.Sprintf("Failed to request Podman API (unix socket): %v", err)))
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.LogError("Failed to read UNIX Podman API response at %s endpoint %s: %v", label, endpoint, err)
		return
	}
	statsCache.Set(label, body)
	util.LogInfo("[%s] Podman API cached for endpoint %s, %d bytes", label, endpoint, len(body))
}

