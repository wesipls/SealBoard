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
				for _, endpoint := range api.PodmanLibpodEndpoints {
								api.CallPodmanAPIUnixEndpoint(sp, host.Name, endpoint, podmanStatsCache)
							}
				
			}
			continue
		}
		if host.LocalSocketPath != "" && host.RemoteSocketPath != "" {
			lsp := host.LocalSocketPath
			for _, endpoint := range api.PodmanLibpodEndpoints {
							api.CallPodmanAPIUnixEndpoint(lsp, host.Name, endpoint, podmanStatsCache)
						}
			
		}
	}
}

