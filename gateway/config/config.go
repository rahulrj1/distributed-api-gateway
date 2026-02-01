package config

import (
	"os"
	"strconv"
)

// Config holds the gateway configuration
type Config struct {
	Host string
	Port int
}

// Load reads configuration from environment variables with defaults
func Load() *Config {
	cfg := &Config{
		Host: getEnv("SERVER_HOST", "0.0.0.0"),
		Port: getEnvInt("SERVER_PORT", 5000),
	}
	return cfg
}

// Address returns the full address string for the server
func (c *Config) Address() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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
