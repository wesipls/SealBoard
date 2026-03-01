package main

import (
	"encoding/json"
	"log"
	"time"

	"sealboard/internal/util"
	"sealboard/internal/api"
	"sealboard/internal/config"
	"sealboard/internal/polling"
	"sealboard/internal/tunnel"
)


var podmanStatsCache = api.NewPodmanStatsCache()

func main() {
	hosts, globalInterval, allowedHTTPHosts, err := config.LoadConfig("seals.cfg")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	interval := globalInterval
	util.LogInfo("Polling hosts every %d seconds. Press Ctrl+C to exit.", interval)

	// Set up all required SSH+unix tunnels just once at startup
	tunnel.SetupTunnels(hosts)

	// Start the lightweight HTTP stats server restricted to allowed hosts
	api.NewStatsServer(allowedHTTPHosts, func() interface{} {
			// Serve latest cached Podman data per host
			result := make(map[string]interface{})
			podmanStatsCache.Range(func(label string, data []byte) {
				var parsed interface{}
				if err := json.Unmarshal(data, &parsed); err == nil {
					result[label] = parsed
				} else {
					// If parsing fails, emit a standard error array for this label
					errmsg := util.FormatErrorMsg("Internal stats/cache error for %s: %v", label, err)
					result[label] = json.RawMessage(util.APIErrorArray(label, errmsg))
				}
			})
			return result
			}, podmanStatsCache).Start()
	

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		polling.PollHosts(hosts, podmanStatsCache)
		<-ticker.C
	}
}
