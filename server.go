package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
)

var allowedHTTPHosts []string

func StartStatsServer(allowedHosts []string, statsFunc func() interface{}) {
	allowedHTTPHosts = allowedHosts
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Printf("Cannot parse remote address: %v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		allowed := false
		for _, host := range allowedHTTPHosts {
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
		json.NewEncoder(w).Encode(statsFunc())
	})

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	go func() {
		log.Printf("Stats HTTP+static server listening on 127.0.0.1:8080 (allowed hosts: %v)", allowedHosts)
		http.ListenAndServe("127.0.0.1:8080", nil)
	}()
}

