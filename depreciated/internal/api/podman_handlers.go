package api

import (
	"encoding/json"
	"net/http"
	"sealboard/internal/util"
)

// RegisterPodmanAPIHandlers sets up Podman-related API endpoints on the provided StatsServer.
func (s *StatsServer) RegisterPodmanAPIHandlers() {
	s.handleAPI("/api/host/", func(r *http.Request) (map[string]string, int, string) {
		segments := util.SplitAndClean(r.URL.Path)
		if len(segments) < 3 {
			return nil, http.StatusBadRequest, "invalid api path"
		}
		vars := map[string]string{"hostLabel": segments[2]}
		if len(segments) > 3 {
			vars["object"] = segments[3]
		}
		if len(segments) > 4 {
			vars["id"] = segments[4]
		}
		if len(segments) > 5 {
			vars["type"] = segments[5]
		}
		return vars, http.StatusOK, ""
	}, func(w http.ResponseWriter, r *http.Request, vars map[string]string) {
		hostLabel := vars["hostLabel"]
		object := vars["object"]
		id := vars["id"]
		typeSuffix := vars["type"]
		// Choose correct endpoint for each object
		var endpoint string
		switch object {
		case "containers":
			endpoint = "http://d/v4.0.0/libpod/containers/json?all=true"
		case "pods":
			endpoint = "http://d/v4.0.0/libpod/pods/json"
		}
		if endpoint != "" {
			data, ok := s.statsCache.Get(hostLabel, endpoint)
			if !ok {
							util.HandleAPIError(w, http.StatusNotFound, hostLabel, object+" stats not available")
				return
			}
			var arr []map[string]interface{}
			json.Unmarshal(data, &arr)
			json.NewEncoder(w).Encode(arr)
			return
		}
		// Individual container, all supported endpoints with {id} substitution
		if object == "container" && id != "" && typeSuffix != "" {
			// Map logical types to podman endpoint templates
			var endpoint string
			switch typeSuffix {
			case "inspect":
				endpoint = "http://d/v4.0.0/libpod/containers/{id}/json"
			case "logs":
				endpoint = "http://d/v4.0.0/libpod/containers/{id}/logs?stderr=true&stdout=true&tail=100"
			case "stats":
				endpoint = "http://d/v4.0.0/libpod/containers/{id}/stats?stream=false"
			case "top":
				endpoint = "http://d/v4.0.0/libpod/containers/{id}/top"
			default:
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("unsupported container endpoint type"))
				return
			}
			endpoint = InterpolateEndpoint(endpoint, map[string]string{"id": id})
			data, ok := s.statsCache.Get(hostLabel, endpoint)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("container endpoint data not available"))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unsupported or incomplete endpoint"))
	})
}

