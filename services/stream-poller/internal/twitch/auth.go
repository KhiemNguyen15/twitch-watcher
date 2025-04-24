package twitch

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type TokenManager struct {
	clientID     string
	clientSecret string
	accessToken  string
	expiry       time.Time
	mu           sync.Mutex
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func NewTokenManager(clientID, clientSecret string) *TokenManager {
	return &TokenManager{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (tm *TokenManager) GetAccessToken() (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if time.Now().Before(tm.expiry) && tm.accessToken != "" {
		return tm.accessToken, nil
	}

	token, expiry, err := tm.fetchNewToken()
	if err != nil {
		return "", err
	}

	tm.accessToken = token
	tm.expiry = time.Now().Add(expiry)

	return tm.accessToken, nil
}

func (tm *TokenManager) fetchNewToken() (string, time.Duration, error) {
	url := "https://id.twitch.tv/oauth2/token"

	data := fmt.Sprintf(
		"client_id=%s&client_secret=%s&grant_type=client_credentials",
		tm.clientID,
		tm.clientSecret,
	)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("failed to fetch token: status %d", resp.StatusCode)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", 0, err
	}

	if tr.AccessToken == "" {
		return "", 0, errors.New("received empty access token")
	}

	return tr.AccessToken, time.Duration(tr.ExpiresIn) * time.Second, nil
}
