package api

// PodmanLibpodEndpoints holds common Podman libpod API endpoints for polling operations.
var PodmanLibpodEndpoints = []string{
	"http://d/v4.0.0/libpod/containers/json?all=true", // List all containers
	"http://d/v4.0.0/libpod/containers/stats?stream=false", // Container stats (all)
	"http://d/v4.0.0/libpod/info", // Podman info
	"http://d/v4.0.0/libpod/containers/json?all=false", // Running containers only
	// Add additional libpod endpoints here as needed
}

// PodmanContainerEndpointType enumerates supported per-container endpoint types
const (
	PodmanContainerInspect = "inspect"
	PodmanContainerLogs    = "logs"
	PodmanContainerStats   = "stats"
	PodmanContainerTop     = "top"
	PodmanContainerConfig  = "config"
)

// PodmanContainerEndpointURL generates the Podman Unix endpoint URL for a given container and endpoint type
func PodmanContainerEndpointURL(containerID, endpointType string) string {
	switch endpointType {
	case PodmanContainerInspect:
		return "http://d/v4.0.0/libpod/containers/" + containerID + "/json"
	case PodmanContainerLogs:
		return "http://d/v4.0.0/libpod/containers/" + containerID + "/logs?stderr=true&stdout=true&tail=100"
	case PodmanContainerStats:
		return "http://d/v4.0.0/libpod/containers/" + containerID + "/stats?stream=false"
	case PodmanContainerTop:
		return "http://d/v4.0.0/libpod/containers/" + containerID + "/top"
	case PodmanContainerConfig:
		return "http://d/v4.0.0/libpod/containers/" + containerID + "/mounts"
	default:
		return ""
	}
}

