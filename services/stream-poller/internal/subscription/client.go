package subscription

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"time"

	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
)

// Client fetches active subscriptions from the subscription-service internal API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates a subscription service client.
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type activeResponse struct {
	Subscriptions []models.Subscription `json:"subscriptions"`
	Total         int                   `json:"total"`
}

// ListActive fetches all active subscriptions.
func (c *Client) ListActive(ctx context.Context) ([]models.Subscription, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/internal/subscriptions/active", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch active subscriptions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription-service returned %d", resp.StatusCode)
	}

	var ar activeResponse
	if err := json.UnmarshalRead(resp.Body, &ar); err != nil {
		return nil, fmt.Errorf("decode active subscriptions: %w", err)
	}

	return ar.Subscriptions, nil
}
