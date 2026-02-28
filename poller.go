package main

import (
	"strings"
)

// pollHosts polls all hosts for Podman stats via Unix socket
func pollHosts(hosts []HostConfig) {
	for _, host := range hosts {
		LogInfo("--- Connecting to %s ---", host.Name)
		if host.Address == "localhost" || strings.HasPrefix(host.Address, "127.") {
			if host.SocketPath != "" {
				sp := host.SocketPath
				sp = expandUIDVariable(sp)
				callPodmanAPIUnix(sp, host.Name, podmanStatsCache)
			}
			continue
		}
		if host.LocalSocketPath != "" && host.RemoteSocketPath != "" {
			lsp := host.LocalSocketPath
			callPodmanAPIUnix(lsp, host.Name, podmanStatsCache)
		}
	}
}

