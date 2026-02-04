package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadRoutes(t *testing.T) {
	// Create a temporary routes file
	content := `routes:
  - path_prefix: "/api"
    target: "http://localhost:8080"
    strip_prefix: true
    timeout: 10s
  - path_prefix: "/service"
    target: "http://localhost:9090"
    strip_prefix: false
`
	tmpFile, err := os.CreateTemp("", "routes-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Load routes
	routes, err := LoadRoutes(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadRoutes failed: %v", err)
	}

	// Verify routes loaded correctly
	if len(routes.Routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(routes.Routes))
	}

	// Check first route
	if routes.Routes[0].PathPrefix != "/api" {
		t.Errorf("Expected path_prefix '/api', got '%s'", routes.Routes[0].PathPrefix)
	}
	if routes.Routes[0].Target != "http://localhost:8080" {
		t.Errorf("Expected target 'http://localhost:8080', got '%s'", routes.Routes[0].Target)
	}
	if !routes.Routes[0].StripPrefix {
		t.Error("Expected strip_prefix to be true")
	}
	if routes.Routes[0].Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", routes.Routes[0].Timeout)
	}

	// Check second route
	if routes.Routes[1].StripPrefix {
		t.Error("Expected strip_prefix to be false")
	}
}

func TestLoadRoutesFileNotFound(t *testing.T) {
	_, err := LoadRoutes("/nonexistent/path/routes.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestMatchRoute(t *testing.T) {
	routes := &RoutesConfig{
		Routes: []Route{
			{PathPrefix: "/service-a", Target: "http://a:6000"},
			{PathPrefix: "/service-b", Target: "http://b:6001"},
			{PathPrefix: "/api/v1", Target: "http://api:8080"},
		},
	}

	tests := []struct {
		path     string
		expected string
	}{
		{"/service-a/hello", "http://a:6000"},
		{"/service-a", "http://a:6000"},
		{"/service-b/users/123", "http://b:6001"},
		{"/api/v1/resources", "http://api:8080"},
		{"/unknown", ""},
		{"/", ""},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			route := routes.MatchRoute(tc.path)
			if tc.expected == "" {
				if route != nil {
					t.Errorf("Expected no match for path '%s', got '%s'", tc.path, route.Target)
				}
			} else {
				if route == nil {
					t.Errorf("Expected match for path '%s', got nil", tc.path)
				} else if route.Target != tc.expected {
					t.Errorf("Expected target '%s' for path '%s', got '%s'", tc.expected, tc.path, route.Target)
				}
			}
		})
	}
}
