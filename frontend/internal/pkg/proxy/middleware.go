package proxy

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ProxyHandler struct {
	logger     *slog.Logger
	backendURL string
	client     *http.Client
}

func NewProxyHandler(logger *slog.Logger, backendURL string) *ProxyHandler {
	return &ProxyHandler{
		logger:     logger,
		backendURL: strings.TrimSuffix(backendURL, "/"),
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxyRequest(w, r)
}

func (p *ProxyHandler) proxyRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	backendURL, err := url.Parse(p.backendURL + r.URL.Path)
	if err != nil {
		p.logger.Error("failed to parse backend URL", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	backendURL.RawQuery = r.URL.RawQuery

	req, err := http.NewRequestWithContext(r.Context(), r.Method, backendURL.String(), r.Body)
	if err != nil {
		p.logger.Error("failed to create proxy request", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		if strings.ToLower(key) == "host" {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Error("failed to proxy request", "error", err, "path", r.URL.Path)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	size, err := io.Copy(w, resp.Body)
	if err != nil {
		p.logger.Error("failed to copy response body", "error", err)
		return
	}

	duration := time.Since(start)
	p.logger.Info("proxy response",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery,
		"backend_url", backendURL.String(),
		"status", resp.StatusCode,
		"size", size,
		"duration_ms", duration.Milliseconds(),
		"content_type", resp.Header.Get("Content-Type"),
	)
}
