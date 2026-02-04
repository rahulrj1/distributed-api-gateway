package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/distributed-api-gateway/gateway/config"
	"github.com/distributed-api-gateway/gateway/proxy"
)

// ErrorResponse represents the standard error response format per LLD
type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ProxyHandler creates a handler for proxying requests to backend services
func ProxyHandler(routes *config.RoutesConfig, forwarder *proxy.Forwarder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Match route
		route := routes.MatchRoute(r.URL.Path)
		if route == nil {
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Route not found")
			return
		}

		// Forward request
		if err := forwarder.Forward(w, r, route); err != nil {
			if proxyErr, ok := err.(*proxy.ProxyError); ok {
				code := "BAD_GATEWAY"
				if proxyErr.Code == http.StatusGatewayTimeout {
					code = "GATEWAY_TIMEOUT"
				}
				writeError(w, r, proxyErr.Code, code, proxyErr.Message)
				return
			}
			log.Printf("Unexpected proxy error: %v", err)
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
	}
}

// writeError writes a standard error response per LLD format
func writeError(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
		RequestID: r.Header.Get("X-Request-ID"), // From client if provided
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}
