package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Route represents a single route configuration
type Route struct {
	PathPrefix  string        `yaml:"path_prefix"`
	Target      string        `yaml:"target"`
	StripPrefix bool          `yaml:"strip_prefix"`
	Timeout     time.Duration `yaml:"timeout"`
}

// RoutesConfig holds all route configurations
type RoutesConfig struct {
	Routes []Route `yaml:"routes"`
}

// LoadRoutes reads route configuration from YAML file
func LoadRoutes(path string) (*RoutesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read routes file: %w", err)
	}

	var cfg RoutesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse routes file: %w", err)
	}

	return &cfg, nil
}

// MatchRoute finds a route that matches the given path
func (rc *RoutesConfig) MatchRoute(path string) *Route {
	for i := range rc.Routes {
		if strings.HasPrefix(path, rc.Routes[i].PathPrefix) {
			return &rc.Routes[i]
		}
	}
	return nil
}
