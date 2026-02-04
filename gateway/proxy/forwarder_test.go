package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/distributed-api-gateway/gateway/config"
)

func TestForwarderBasicProxy(t *testing.T) {
	// Create a test backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend-Response", "true")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"hello from backend"}`))
	}))
	defer backend.Close()

	// Create forwarder and route
	forwarder := NewForwarder()
	route := &config.Route{
		PathPrefix:  "/api",
		Target:      backend.URL,
		StripPrefix: true,
		Timeout:     5 * time.Second,
	}

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	// Forward request
	err := forwarder.Forward(rec, req, route)
	if err != nil {
		t.Fatalf("Forward failed: %v", err)
	}

	// Check response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "hello from backend") {
		t.Errorf("Expected backend response in body, got: %s", body)
	}

	if rec.Header().Get("X-Backend-Response") != "true" {
		t.Error("Expected backend header to be copied")
	}

	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("Expected X-Request-ID header to be set")
	}
}

func TestForwarderStripPrefix(t *testing.T) {
	var receivedPath string

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	forwarder := NewForwarder()

	// Test with strip_prefix: true
	route := &config.Route{
		PathPrefix:  "/service-a",
		Target:      backend.URL,
		StripPrefix: true,
		Timeout:     5 * time.Second,
	}

	req := httptest.NewRequest(http.MethodGet, "/service-a/hello/world", nil)
	rec := httptest.NewRecorder()

	forwarder.Forward(rec, req, route)

	if receivedPath != "/hello/world" {
		t.Errorf("Expected stripped path '/hello/world', got '%s'", receivedPath)
	}

	// Test with strip_prefix: false
	route.StripPrefix = false
	req = httptest.NewRequest(http.MethodGet, "/service-a/hello/world", nil)
	rec = httptest.NewRecorder()

	forwarder.Forward(rec, req, route)

	if receivedPath != "/service-a/hello/world" {
		t.Errorf("Expected full path '/service-a/hello/world', got '%s'", receivedPath)
	}
}

func TestForwarderCopiesHeaders(t *testing.T) {
	var receivedHeaders http.Header

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	forwarder := NewForwarder()
	route := &config.Route{
		PathPrefix:  "/api",
		Target:      backend.URL,
		StripPrefix: true,
		Timeout:     5 * time.Second,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("X-Custom-Header", "custom-value")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	forwarder.Forward(rec, req, route)

	if receivedHeaders.Get("X-Custom-Header") != "custom-value" {
		t.Error("Expected custom header to be forwarded")
	}

	if receivedHeaders.Get("X-Request-ID") == "" {
		t.Error("Expected X-Request-ID to be added")
	}

	if receivedHeaders.Get("X-Forwarded-For") == "" {
		t.Error("Expected X-Forwarded-For to be added")
	}
}

func TestForwarderRemovesAuthHeader(t *testing.T) {
	var receivedHeaders http.Header

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	forwarder := NewForwarder()
	route := &config.Route{
		PathPrefix:  "/api",
		Target:      backend.URL,
		StripPrefix: true,
		Timeout:     5 * time.Second,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()

	forwarder.Forward(rec, req, route)

	if receivedHeaders.Get("Authorization") != "" {
		t.Error("Expected Authorization header to be removed")
	}
}

func TestForwarderPreservesQueryParams(t *testing.T) {
	var receivedQuery string

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	forwarder := NewForwarder()
	route := &config.Route{
		PathPrefix:  "/api",
		Target:      backend.URL,
		StripPrefix: true,
		Timeout:     5 * time.Second,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/test?foo=bar&baz=qux", nil)
	rec := httptest.NewRecorder()

	forwarder.Forward(rec, req, route)

	if receivedQuery != "foo=bar&baz=qux" {
		t.Errorf("Expected query 'foo=bar&baz=qux', got '%s'", receivedQuery)
	}
}

func TestForwarderPOSTWithBody(t *testing.T) {
	var receivedBody string

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer backend.Close()

	forwarder := NewForwarder()
	route := &config.Route{
		PathPrefix:  "/api",
		Target:      backend.URL,
		StripPrefix: true,
		Timeout:     5 * time.Second,
	}

	body := `{"name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	forwarder.Forward(rec, req, route)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}

	if receivedBody != body {
		t.Errorf("Expected body '%s', got '%s'", body, receivedBody)
	}
}

func TestForwarderTimeout(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	forwarder := NewForwarder()
	route := &config.Route{
		PathPrefix:  "/api",
		Target:      backend.URL,
		StripPrefix: true,
		Timeout:     50 * time.Millisecond, // Very short timeout
	}

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	err := forwarder.Forward(rec, req, route)

	if err == nil {
		t.Error("Expected timeout error")
	}

	proxyErr, ok := err.(*ProxyError)
	if !ok {
		t.Fatalf("Expected ProxyError, got %T", err)
	}

	if proxyErr.Code != http.StatusGatewayTimeout {
		t.Errorf("Expected 504 Gateway Timeout, got %d", proxyErr.Code)
	}
}

func TestForwarderBackendUnreachable(t *testing.T) {
	forwarder := NewForwarder()
	route := &config.Route{
		PathPrefix:  "/api",
		Target:      "http://localhost:59999", // Unlikely to be running
		StripPrefix: true,
		Timeout:     1 * time.Second,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	err := forwarder.Forward(rec, req, route)

	if err == nil {
		t.Error("Expected error for unreachable backend")
	}

	proxyErr, ok := err.(*ProxyError)
	if !ok {
		t.Fatalf("Expected ProxyError, got %T", err)
	}

	if proxyErr.Code != http.StatusBadGateway {
		t.Errorf("Expected 502 Bad Gateway, got %d", proxyErr.Code)
	}
}
