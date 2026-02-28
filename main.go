package main

import (
	"encoding/json"
	"log"
	"time"
)

// expandUIDVariable now handled by tunnel.go

// SetupTunnels is now handled in tunnel.go

// pollHosts now provided in poller.go

func main() {
	hosts, globalInterval, allowedHTTPHosts, err := loadConfig("seals.cfg")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	interval := globalInterval
	LogInfo("Polling hosts every %d seconds. Press Ctrl+C to exit.", interval)

	// Set up all required SSH+unix tunnels just once at startup
	SetupTunnels(hosts)

	// Start the lightweight HTTP stats server restricted to allowed hosts
	StartStatsServer(allowedHTTPHosts, func() interface{} {
		// Serve latest cached Podman data per host
		podmanStatsMu.RLock()
		defer podmanStatsMu.RUnlock()
		result := make(map[string]interface{})
		for label, data := range podmanStats {
			var parsed interface{}
			if err := json.Unmarshal(data, &parsed); err == nil {
				result[label] = parsed
			} else {
				// If parsing fails, emit a standard error array for this label
				errmsg := FormatErrorMsg("Internal stats/cache error for %s: %v", label, err)
				result[label] = json.RawMessage(APIErrorArray(label, errmsg))
			}
		}
		return result
	})

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		pollHosts(hosts)
		<-ticker.C
	}
}
