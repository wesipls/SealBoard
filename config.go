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
	SSHPort           int    `yaml:"ssh_port"`
	Interval          int    `yaml:"interval"`
}

// Config holds all HostConfigs and global interval
// Add global polling interval
// interval is the polling interval in seconds
// hosts is the list of host configs
type Config struct {
	Interval int          `yaml:"interval"`
	Hosts    []HostConfig `yaml:"hosts"`
}

// loadConfig reads YAML config file and returns HostConfigs and global interval
func loadConfig(path string) ([]HostConfig, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	var cfg Config
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, 0, err
	}
	return cfg.Hosts, cfg.Interval, nil
}

