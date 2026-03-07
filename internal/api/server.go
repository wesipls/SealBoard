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
 	statsCache interface{Get(label, endpoint string) ([]byte, bool)}
}

// NewStatsServer initializes a StatsServer
func NewStatsServer(allowedHosts []string, statsFunc func() interface{}, statsCache interface{Get(label, endpoint string) ([]byte, bool)}) *StatsServer {
	return &StatsServer{
		allowedHosts: allowedHosts,
		statsFunc: statsFunc,
		statsCache: statsCache,
	}
}

// handleAPI abstracts common http logic (IP check, content-type, error writing) for Podman endpoints.
func (s *StatsServer) handleAPI(pattern string, parsePath func(*http.Request) (map[string]string, int, string), handler func(http.ResponseWriter, *http.Request, map[string]string)) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
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

		vars := map[string]string{}
		msg := ""
		if parsePath != nil {
			var status int
			vars, status, msg = parsePath(r)
			if status != http.StatusOK {
				w.WriteHeader(status)
				w.Write([]byte(msg))
				return
			}
		}
		handler(w, r, vars)
	})
}

// Start launches the HTTP stats server and container endpoints
func (s *StatsServer) Start() {
	s.handleAPI("/stats", nil, func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		json.NewEncoder(w).Encode(s.statsFunc())
	})

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

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	go func() {
		util.LogInfo("Stats HTTP+static server listening on 127.0.0.1:8080 (allowed hosts: %v)", s.allowedHosts)
		http.ListenAndServe("127.0.0.1:8080", nil)
	}()
}

