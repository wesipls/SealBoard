package api

import "strings"

// PodmanLibpodEndpoints holds common Podman libpod API endpoints for polling operations.
// Endpoints marked with placeholders like {id} require runtime interpolation.
var PodmanLibpodEndpoints = []string{
	// System/Image
	//	"http://d/v4.0.0/libpod/version",             // Version
	//	"http://d/v4.0.0/libpod/info",                // System info
	//	"http://d/v4.0.0/libpod/events?stream=false", // Events
	// Image endpoints
	//	"http://d/v4.0.0/libpod/images/json?all=true", // List all images
	//	"http://d/v4.0.0/libpod/images/prune",         // Prune images (POST)
	// Container endpoints
	"http://d/v4.0.0/libpod/containers/json?all=true", // List all containers
	//	"http://d/v4.0.0/libpod/containers/json?all=false",     // List running containers
	//	"http://d/v4.0.0/libpod/containers/stats?stream=false", // Stats all containers
	// Specific container endpoints (require {id} interpolation at runtime; see InterpolateEndpoint)
	//	"http://d/v4.0.0/libpod/containers/{id}/json",                                  // Inspect container {id}
	//	"http://d/v4.0.0/libpod/containers/{id}/logs?stderr=true&stdout=true&tail=100", // Logs for {id}
	"http://d/v4.0.0/libpod/containers/{id}/stats?stream=false", // Stats for {id}
	//	"http://d/v4.0.0/libpod/containers/{id}/top",                                   // Top for {id}
	// Pod endpoints
	"http://d/v4.0.0/libpod/pods/json", // List pods
	//	"http://d/v4.0.0/libpod/pods/stats?stream=false", // Pod stats
	//	"http://d/v4.0.0/libpod/pods/{id}/json",          // Inspect pod {id}
	//	"http://d/v4.0.0/libpod/pods/{id}/top",           // Pod top {id}
	// Network endpoints
	//	"http://d/v4.0.0/libpod/networks/json", // List networks
	//	"http://d/v4.0.0/libpod/networks/{id}", // Inspect network {id}
	// Volume endpoints
	//	"http://d/v4.0.0/libpod/volumes/json",   // List volumes
	//	"http://d/v4.0.0/libpod/volumes/{name}", // Inspect volume {name}
	// System prune endpoints
	//	"http://d/v4.0.0/libpod/system/prune", // System prune (POST)
	// Additional endpoints may be added here
}

// InterpolateEndpoint replaces placeholders like {id}, {name}, {host} in the endpoint string.
func InterpolateEndpoint(endpoint string, params map[string]string) string {
	for k, v := range params {
		endpoint = strings.ReplaceAll(endpoint, "{"+k+"}", v)
	}
	return endpoint
}


