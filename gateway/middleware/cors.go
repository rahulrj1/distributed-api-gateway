package middleware

import "net/http"

// CORS adds Cross-Origin Resource Sharing headers for the visualizer.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from visualizer
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Trace-ID, X-Request-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Trace-ID, X-RateLimit-Limit, X-RateLimit-Remaining, Retry-After")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
