package api

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
)

// StreamPodmanContainerData initiates a streaming request to the Podman API for a container (e.g., stats or logs).
// Arguments:
// - socketPath: local unix socket to connect to Podman
// - containerID: container ID or name
// - endpoint: Podman endpoint to stream (e.g., "/stats?stream=true" or "/logs?follow=true")
// - handleChunk: callback invoked with each chunk of data received
// Returns error if connection/setup fails
func StreamPodmanContainerData(socketPath, containerID, endpoint string, handleChunk func([]byte)) error {
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}
	client := &http.Client{
		Transport: transport,
	}
	url := fmt.Sprintf("http://d/v4.0.0/containers/%s%s", containerID, endpoint)
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("Failed to start stream for %s: %w", containerID, err)
	}
	defer resp.Body.Close()
	buf := make([]byte, 4096)
	for {
		n,
		err := resp.Body.Read(buf)
		if n > 0 {
			// Deliver chunk to caller
			handleChunk(buf[:n])
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("Stream interrupted for %s: %w", containerID, err)
		}
	}
	return nil
}

