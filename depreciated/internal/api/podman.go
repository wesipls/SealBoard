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
 	stats map[string]map[string][]byte // key: host label, value: map[endpoint]data
}

func NewPodmanStatsCache() *PodmanStatsCache {
	return &PodmanStatsCache{stats: make(map[string]map[string][]byte)}
}

func (c *PodmanStatsCache) Set(label string, endpoint string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stats[label] == nil {
		c.stats[label] = make(map[string][]byte)
	}
	c.stats[label][endpoint] = data
}

func (c *PodmanStatsCache) Get(label, endpoint string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	endpoints, ok := c.stats[label]
	if !ok {
		return nil, false
	}
	data, ok := endpoints[endpoint]
	return data, ok
}

func (c *PodmanStatsCache) Range(fn func(label, endpoint string, data []byte)) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for label, endpoints := range c.stats {
		for endpoint, data := range endpoints {
			fn(label, endpoint, data)
		}
	}
}

// CallPodmanAPIUnixEndpoint queries the Podman API over Unix socket at arbitrary endpoint
// client cache for persistent unix socket connections
var clientMu sync.Mutex
var unixClientMap = make(map[string]*http.Client)

func getUnixHttpClient(socketPath string) *http.Client {
	clientMu.Lock()
	defer clientMu.Unlock()
	if client, ok := unixClientMap[socketPath]; ok {
		return client
	}
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
			// Improve connection reuse and pool behavior:
			DisableKeepAlives: false,
			MaxIdleConns: 16,
			MaxIdleConnsPerHost: 16,
			IdleConnTimeout: 90 * 1e9, // 90s in nanoseconds, Go uses time.Duration
	}
	client := &http.Client{Transport: transport}
	unixClientMap[socketPath] = client
	return client
}

func CallPodmanAPIUnixEndpoint(socketPath, label, endpoint string, statsCache *PodmanStatsCache) {
	client := getUnixHttpClient(socketPath)
	url := endpoint // Fully customizable endpoint (must be complete, host part still ignored)
	resp, err := client.Get(url)
	if err != nil {
		util.LogError("Failed to request Podman API (unix socket) at %s endpoint %s: %v", label, endpoint, err)
		statsCache.Set(label, endpoint, util.APIErrorArray(label, fmt.Sprintf("Failed to request Podman API (unix socket): %v", err)))
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.LogError("Failed to read UNIX Podman API response at %s endpoint %s: %v", label, endpoint, err)
		return
	}
	statsCache.Set(label, endpoint, body)
	util.LogInfo("[%s] Podman API cached for endpoint %s, %d bytes", label, endpoint, len(body))
}

