package proxy

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/distributed-api-gateway/gateway/config"
	"github.com/google/uuid"
)

type Forwarder struct {
	client *http.Client
}

func NewForwarder() *Forwarder {
	return &Forwarder{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: config.DefaultConnectTimeout * time.Second,
				}).DialContext,
			},
		},
	}
}

// Forward proxies request to backend. Example: /service-a/users → http://service-a:6000/users
func (f *Forwarder) Forward(w http.ResponseWriter, r *http.Request, route *config.Route) error {
	targetURL := buildTargetURL(route, r.URL.Path, r.URL.RawQuery)
	requestID := getOrCreateRequestID(r)

	ctx, cancel := context.WithTimeout(r.Context(), route.Timeout)
	defer cancel()

	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, r.Body)
	if err != nil {
		return &ProxyError{Code: http.StatusBadGateway, Message: "failed to create request"}
	}

	copyHeaders(r.Header, proxyReq.Header)
	proxyReq.Header.Set("X-Request-ID", requestID)
	proxyReq.Header.Set("X-Forwarded-For", getClientIP(r))
	proxyReq.Header.Del("Authorization")

	resp, err := f.client.Do(proxyReq)
	if err != nil {
		return classifyError(ctx)
	}
	defer resp.Body.Close()

	copyHeaders(resp.Header, w.Header())
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	return nil
}

type ProxyError struct {
	Code    int
	Message string
}

func (e *ProxyError) Error() string { return e.Message }

// buildTargetURL: "/service-a/users" + strip → "http://service-a:6000/users"
func buildTargetURL(route *config.Route, path, query string) string {
	targetPath := path
	if route.StripPrefix {
		targetPath = strings.TrimPrefix(path, route.PathPrefix)
		if targetPath == "" {
			targetPath = "/"
		}
	}
	if query != "" {
		return route.Target + targetPath + "?" + query
	}
	return route.Target + targetPath
}

// classifyError: timeout → 504, connection failure → 502
func classifyError(ctx context.Context) *ProxyError {
	if ctx.Err() == context.DeadlineExceeded {
		return &ProxyError{Code: http.StatusGatewayTimeout, Message: "backend timeout"}
	}
	return &ProxyError{Code: http.StatusBadGateway, Message: "backend unreachable"}
}

func copyHeaders(src, dst http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// getOrCreateRequestID: use existing X-Request-ID or generate UUID
func getOrCreateRequestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	return uuid.New().String()
}

// getClientIP: X-Forwarded-For="1.2.3.4, 5.6.7.8" → "1.2.3.4"
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i != -1 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}
	return r.RemoteAddr
}
