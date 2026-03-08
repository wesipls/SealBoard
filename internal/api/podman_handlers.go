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
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(object + " stats not available"))
				return
			}
			var arr []map[string]interface{}
			json.Unmarshal(data, &arr)
			json.NewEncoder(w).Encode(arr)
			return
		}
		// Individual container
		if object == "container" && id != "" && typeSuffix != "" {
			cdata, ok := s.statsCache.Get(hostLabel, "http://d/v4.0.0/libpod/containers/json?all=true")
			if !ok { w.WriteHeader(http.StatusNotFound); w.Write([]byte("host stats not available")); return }
			var arr []map[string]interface{}
			json.Unmarshal(cdata, &arr)
			for _, c := range arr {
				if cid, ok := c["Id"]; ok && cid == id {
					json.NewEncoder(w).Encode(c)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("container data not available"))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unsupported or incomplete endpoint"))
	})
}

