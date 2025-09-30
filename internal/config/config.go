package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// TenantConfig represents configuration for a single tenant
type TenantConfig struct {
	Prefix         string `yaml:"prefix"`
	Password       string `yaml:"password,omitempty"`        // Plain password (not recommended for production)
	PasswordHash   string `yaml:"password_hash,omitempty"`   // Bcrypt hash of password (recommended)
	RateLimit      int    `yaml:"rate_limit,omitempty"`      // Optional rate limiting
	MaxConnections int    `yaml:"max_connections,omitempty"` // Max connections per tenant
}

// Config holds the proxy configuration
type Config struct {
	// Proxy settings
	ListenAddr string `yaml:"listen_addr"`

	// Redis Cluster settings
	ClusterNodes []string `yaml:"cluster_nodes"`
	// Optional cluster authentication
	ClusterUser     string `yaml:"cluster_user,omitempty"`
	ClusterPassword string `yaml:"cluster_password,omitempty"`

	// Authentication settings
	AuthEnabled bool                    `yaml:"auth_enabled"`
	Tenants     map[string]TenantConfig `yaml:"tenants"`

	// Performance settings
	MaxConnections int `yaml:"max_connections"`

	// Logging settings
	LogLevel string `yaml:"log_level"`
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
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
		Tenants:        make(map[string]TenantConfig),
	}

	// Parse tenant config from environment
	// Format: TENANTS=tenant1:prefix:password,tenant2:prefix:passwordhash:hash
	tenantsStr := getEnvOrDefault("TENANTS", "")
	if tenantsStr != "" {
		tenants := strings.Split(tenantsStr, ",")
		for _, tenantStr := range tenants {
			parts := strings.Split(tenantStr, ":")
			if len(parts) >= 2 {
				username := parts[0]
				prefix := parts[1]
				tenant := TenantConfig{
					Prefix: prefix + ":",
				}

				// Check if password is provided
				if len(parts) >= 3 {
					// If 4th part exists and is "hash", it's a password hash
					if len(parts) >= 4 && parts[3] == "hash" {
						tenant.PasswordHash = parts[2]
					} else {
						tenant.Password = parts[2]
					}
				}

				config.Tenants[username] = tenant
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

	if config.AuthEnabled && len(config.Tenants) == 0 {
		return fmt.Errorf("authentication is enabled but no tenants are configured")
	}

	return nil
}
