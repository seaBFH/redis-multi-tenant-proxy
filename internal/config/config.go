package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config holds the proxy configuration
type Config struct {
	// Proxy settings
	ListenAddr string `yaml:"listen_addr"`

	// Redis Cluster settings
	ClusterNodes []string `yaml:"cluster_nodes"`

	// Authentication settings
	AuthEnabled bool              `yaml:"auth_enabled"`
	TenantMap   map[string]string `yaml:"tenant_map"`

	// Performance settings
	MaxConnections int `yaml:"max_connections"`

	// Logging settings
	LogLevel string `yaml:"log_level"`
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, validateConfig(&config)
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	config := &Config{
		ListenAddr:     getEnvOrDefault("PROXY_LISTEN_ADDR", ":6380"),
		ClusterNodes:   strings.Split(getEnvOrDefault("REDIS_CLUSTER_NODES", "localhost:6379"), ","),
		AuthEnabled:    getEnvOrDefault("AUTH_ENABLED", "true") == "true",
		MaxConnections: 1000, // Default value
		LogLevel:       getEnvOrDefault("LOG_LEVEL", "info"),
	}

	// Parse tenant map from environment
	tenantMapStr := getEnvOrDefault("TENANT_MAP", "")
	if tenantMapStr != "" {
		config.TenantMap = make(map[string]string)
		pairs := strings.Split(tenantMapStr, ",")
		for _, pair := range pairs {
			parts := strings.Split(pair, ":")
			if len(parts) == 2 {
				config.TenantMap[parts[0]] = parts[1] + ":"
			}
		}
	}

	return config, validateConfig(config)
}

// Helper to get environment variable with default
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Validate configuration
func validateConfig(config *Config) error {
	if config.ListenAddr == "" {
		return fmt.Errorf("listen_addr is required")
	}

	if len(config.ClusterNodes) == 0 {
		return fmt.Errorf("at least one cluster node is required")
	}

	return nil
}
