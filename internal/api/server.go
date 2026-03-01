package api

import (
	"encoding/json"
	"net"
	"net/http"
	
		"sealboard/internal/util"
)

// StatsServer encapsulates HTTP stats serving logic and allowed hosts
 type StatsServer struct {
	allowedHosts []string
	statsFunc func() interface{}
 	statsCache interface{Get(label string) ([]byte, bool)}
}

// NewStatsServer initializes a StatsServer
func NewStatsServer(allowedHosts []string, statsFunc func() interface{}, statsCache interface{Get(label string) ([]byte, bool)}) *StatsServer {
	return &StatsServer{
		allowedHosts: allowedHosts,
		statsFunc: statsFunc,
		statsCache: statsCache,
	}
}

// Start launches the HTTP stats server and container endpoints
func (s *StatsServer) Start() {
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			util.LogError("Cannot parse remote address: %v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		allowed := false
		for _, host := range s.allowedHosts {
			if remoteIP == host || host == "*" {
				allowed = true
				break
			}
		}
		if !allowed {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("forbidden"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.statsFunc())
	})

	// New: container-specific endpoints
	http.HandleFunc("/api/host/", func(w http.ResponseWriter, r *http.Request) {
		remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil { w.WriteHeader(http.StatusForbidden); return }
		allowed := false
		for _, host := range s.allowedHosts {
			if remoteIP == host || host == "*" { allowed = true; break }
		}
		if !allowed { w.WriteHeader(http.StatusForbidden); w.Write([]byte("forbidden")); return }
		w.Header().Set("Content-Type", "application/json")

		// pattern: /api/host/{hostLabel}/container/{containerID}/{datatype}
		// simple parsing
		var hostLabel, containerID, dataType string
		path := r.URL.Path
		segments := util.SplitAndClean(path)
		if len(segments) < 6 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid container api path"))
			return
		}
		hostLabel = segments[2]
		containerID = segments[4]
		dataType = segments[5] // "stats", "logs", "config"
		// Implementation: serve actual data from cache (stats/config) or via streaming (logs/stats w/stream)
		var result interface{}
		if dataType == "stats" || dataType == "config" {
			// TODO: Replace podmanStatsCache with new container cache for per-container data
			// For demo, just fetch host-level stats and filter for matching container
			data, ok := s.statsCache.Get(hostLabel)
			if !ok { w.WriteHeader(http.StatusNotFound); w.Write([]byte("host stats not available")); return }
			var arr []map[string]interface{}
			json.Unmarshal(data, &arr)
			for _, c := range arr {
				if id, ok := c["Id"]; ok && id == containerID {
					result = c
					break
				}
			}
			if result == nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("container data not available"))
				return
			}
			json.NewEncoder(w).Encode(result)
			return
		}
		if dataType == "logs" {
			// Streaming TODO: use StreamPodmanContainerData from podman_stream.go
			result = map[string]string{"host":hostLabel, "container":containerID, "type":dataType, "message":"streaming logs not implemented yet"}
			json.NewEncoder(w).Encode(result)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unsupported data type"))
	})

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	go func() {
		util.LogInfo("Stats HTTP+static server listening on 127.0.0.1:8080 (allowed hosts: %v)", s.allowedHosts)
		http.ListenAndServe("127.0.0.1:8080", nil)
	}()
}

