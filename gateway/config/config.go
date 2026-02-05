package config

import (
	"os"
	"strconv"
)

// Default configuration values
const (
	DefaultPort           = 5000
	DefaultRoutesPath     = "config/routes.yaml"
	DefaultConnectTimeout = 1 // seconds - per HLD §9
	DefaultPublicKeyPath  = "keys/public.pem"
	DefaultRedisAddr      = "redis:6379"
	DefaultRateLimit      = 100 // requests per minute
)

// Config holds the gateway configuration
type Config struct {
	Port int
}

// Load reads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Port: getEnvInt("SERVER_PORT", DefaultPort),
	}
}

// Address returns the full address string for the server
func (c *Config) Address() string {
	return "0.0.0.0:" + strconv.Itoa(c.Port)
}

// getEnvInt returns the value of an environment variable as int or a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
