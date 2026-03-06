package polling

import (
	"strings"
	"sealboard/internal/config"
	"sealboard/internal/util"
	"sealboard/internal/api"
)

// pollHosts polls all hosts for Podman stats via Unix socket
func PollHosts(hosts []config.HostConfig, podmanStatsCache *api.PodmanStatsCache) {
	for _, host := range hosts {
		util.LogInfo("--- Connecting to %s ---", host.Name)
		if host.Address == "localhost" || strings.HasPrefix(host.Address, "127.") {
			if host.SocketPath != "" {
				sp := host.SocketPath
				sp = util.ExpandUIDVariable(sp)
				api.CallPodmanAPIUnixEndpoint(sp, host.Name, "http://d/v4.0.0/libpod/containers/json?all=true", podmanStatsCache)
			}
			continue
		}
		if host.LocalSocketPath != "" && host.RemoteSocketPath != "" {
			lsp := host.LocalSocketPath
			api.CallPodmanAPIUnixEndpoint(lsp, host.Name, "http://d/v4.0.0/libpod/containers/json?all=true", podmanStatsCache)
		}
	}
}

