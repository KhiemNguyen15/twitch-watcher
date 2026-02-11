package twitch

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const tokenURL = "https://id.twitch.tv/oauth2/token"

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// TokenManager manages a Twitch app-access token with proactive refresh.
type TokenManager struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client

	mu        sync.RWMutex
	token     string
	expiresAt time.Time
}

// NewTokenManager creates a TokenManager. Call Refresh once before use.
func NewTokenManager(clientID, clientSecret string) *TokenManager {
	return &TokenManager{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Token returns the current access token, refreshing if it expires within 5 minutes.
func (m *TokenManager) Token(ctx context.Context) (string, error) {
	m.mu.RLock()
	if m.token != "" && time.Until(m.expiresAt) > 5*time.Minute {
		t := m.token
		m.mu.RUnlock()
		return t, nil
	}
	m.mu.RUnlock()

	return m.Refresh(ctx)
}

// Refresh forces a new token fetch from Twitch.
func (m *TokenManager) Refresh(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	body := url.Values{
		"client_id":     {m.clientID},
		"client_secret": {m.clientSecret},
		"grant_type":    {"client_credentials"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}

	var tr tokenResponse
	if err := json.UnmarshalRead(resp.Body, &tr); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	m.token = tr.AccessToken
	m.expiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	return m.token, nil
}
