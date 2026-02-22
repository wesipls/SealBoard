package main

import (
	"os"
	"gopkg.in/yaml.v3"
)

// HostConfig holds info for connecting to one host
type HostConfig struct {
	Name              string `yaml:"name"`
	Address           string `yaml:"address"`
	User              string `yaml:"user"`
	PrivateKeyPath    string `yaml:"private_key_path"`
	SocketPath        string `yaml:"socket_path"`
	RemoteSocketPath  string `yaml:"remote_socket_path"`
	LocalSocketPath   string `yaml:"local_socket_path"`
}

// Config holds all HostConfigs
type Config struct {
	Hosts []HostConfig `yaml:"hosts"`
}

// loadConfig reads YAML config file and returns HostConfigs
func loadConfig(path string) ([]HostConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Config
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return cfg.Hosts, nil
}

