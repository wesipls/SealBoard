package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"context"
)

// callPodmanAPIUnix queries the Podman API over Unix socket
func callPodmanAPIUnix(socketPath, apiPath, label string) {
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}
	client := &http.Client{Transport: transport}
	url := "http://d/v4.0.0/containers/json?all=true" // The host part is ignored for UNIX sockets
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Failed to request Podman API (unix socket) at %s: %v", label, err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read UNIX Podman API response at %s: %v", label, err)
		return
	}
	fmt.Printf("[%s] Podman stats raw response: %s\n", label, string(body))
}

