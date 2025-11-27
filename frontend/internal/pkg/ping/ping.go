package ping

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"frontend/internal/pkg/httpclient"
)

const (
	defaultTimeout = 5 * time.Second
	pingURI        = "/ping"
)

type Client struct {
	client     *http.Client
	apiBaseURL string
}

func NewClient(apiBaseURL string) *Client {
	return &Client{
		client:     httpclient.New(&http.Client{Timeout: defaultTimeout}),
		apiBaseURL: apiBaseURL,
	}
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+pingURI, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed with status: %d", resp.StatusCode)
	}

	return nil
}
