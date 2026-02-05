package middleware

import (
	"net/http"
	"strings"

	"github.com/distributed-api-gateway/gateway/pkg/jwt"
)

// Auth returns middleware that validates JWT tokens.
// On success, adds X-User-ID and X-Client-ID headers.
func Auth(validator *jwt.Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			token := extractToken(r)
			if token == "" {
				writeAuthError(w, "missing authorization token")
				return
			}

			// Validate token
			claims, err := validator.Validate(token)
			if err != nil {
				writeAuthError(w, err.Error())
				return
			}

			// Add user info to request headers for downstream
			r.Header.Set("X-User-ID", claims.Sub)
			if claims.ClientID != "" {
				r.Header.Set("X-Client-ID", claims.ClientID)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractToken gets token from "Authorization: Bearer <token>"
func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"` + message + `"}}`))
}
