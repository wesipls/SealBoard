package polling

import (
	"encoding/json"
	"sealboard/internal/api"
	"sealboard/internal/config"
	"sealboard/internal/tunnel"
	"sealboard/internal/util"
	"strings"
)

// pollHosts polls all hosts for Podman stats via Unix socket
func PollHosts(hosts []config.HostConfig, podmanStatsCache *api.PodmanStatsCache) {
	for _, host := range hosts {
		util.LogInfo("--- Connecting to %s ---", host.Name)
		if host.Address == "localhost" || strings.HasPrefix(host.Address, "127.") {
			if host.SocketPath != "" {
				sp := host.SocketPath
				sp = util.ExpandUIDVariable(sp)
				for _, endpoint := range api.PodmanLibpodEndpoints {
					if strings.Contains(endpoint, "{id}") {
						// Fetch the container list first
						cjsonEndpoint := "http://d/v4.0.0/libpod/containers/json?all=true"
						api.CallPodmanAPIUnixEndpoint(sp, host.Name, cjsonEndpoint, podmanStatsCache)
						cjson, ok := podmanStatsCache.Get(host.Name, cjsonEndpoint)
						if ok {
							var containers []map[string]any
							_ = json.Unmarshal(cjson, &containers)
							for _, c := range containers {
								if id, ok := c["Id"].(string); ok {
									fullEndpoint := api.InterpolateEndpoint(endpoint, map[string]string{"id": id})
									api.CallPodmanAPIUnixEndpoint(sp, host.Name, fullEndpoint, podmanStatsCache)
								}
							}
						}
					} else {
						api.CallPodmanAPIUnixEndpoint(sp, host.Name, endpoint, podmanStatsCache)
					}
				}

			}
			continue
		}
		if host.LocalSocketPath != "" && host.RemoteSocketPath != "" {
			// Ensure tunnel is alive before polling
			// (new: calls tunnel.EnsureTunnel)
			tunnel.EnsureTunnel(host)

			lsp := host.LocalSocketPath
			for _, endpoint := range api.PodmanLibpodEndpoints {
				if strings.Contains(endpoint, "{id}") {
					// Fetch container list from remote socket
					cjsonEndpoint := "http://d/v4.0.0/libpod/containers/json?all=true"
					api.CallPodmanAPIUnixEndpoint(lsp, host.Name, cjsonEndpoint, podmanStatsCache)
					cjson, ok := podmanStatsCache.Get(host.Name, cjsonEndpoint)
					if ok {
						var containers []map[string]any
						_ = json.Unmarshal(cjson, &containers)
						for _, c := range containers {
							if id, ok := c["Id"].(string); ok {
								fullEndpoint := api.InterpolateEndpoint(endpoint, map[string]string{"id": id})
								api.CallPodmanAPIUnixEndpoint(lsp, host.Name, fullEndpoint, podmanStatsCache)
							}
						}
					}
				} else {
					api.CallPodmanAPIUnixEndpoint(lsp, host.Name, endpoint, podmanStatsCache)
				}
			}
		}
	}
}
