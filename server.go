package main

import (
	"encoding/json"
	"net"
	"net/http"
)

// StatsServer encapsulates HTTP stats serving logic and allowed hosts
 type StatsServer struct {
	allowedHosts []string
	statsFunc func() interface{}
}

// NewStatsServer initializes a StatsServer
func NewStatsServer(allowedHosts []string, statsFunc func() interface{}) *StatsServer {
	return &StatsServer{
		allowedHosts: allowedHosts,
		statsFunc: statsFunc,
	}
}

// Start launches the HTTP stats server
func (s *StatsServer) Start() {
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			LogError("Cannot parse remote address: %v", err)
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

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	go func() {
		LogInfo("Stats HTTP+static server listening on 127.0.0.1:8080 (allowed hosts: %v)", s.allowedHosts)
		http.ListenAndServe("127.0.0.1:8080", nil)
	}()
}

