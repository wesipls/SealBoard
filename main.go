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

var ContainersStatsCache = podmanStatsCache

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
	// Adapter to expose only the containers endpoint data for /stats



	api.NewStatsServer(allowedHTTPHosts, func() interface{} {
		// Serve latest cached Podman data per host
		result := make(map[string]interface{})
		for _, label := range hosts {
			data, ok := podmanStatsCache.Get(label.Name, "http://d/v4.0.0/libpod/containers/json?all=true")
			if !ok {
				continue
			}
			var parsed interface{}
			if err := json.Unmarshal(data, &parsed); err == nil {
				result[label.Name] = parsed
			} else {
				errmsg := util.FormatErrorMsg("Internal stats/cache error for %s: %v", label.Name, err)
								result[label.Name] = json.RawMessage(util.APIErrorArray(label.Name, errmsg)) // Candidates for direct HandleAPIError, but response write pattern differs in main.go
			}
		}
		return result
	}, ContainersStatsCache).Start()
	

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		polling.PollHosts(hosts, podmanStatsCache)
		<-ticker.C
	}
}
