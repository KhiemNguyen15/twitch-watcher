package twitch

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
)

const helixBase = "https://api.twitch.tv/helix"

// Client is a Twitch Helix API client.
type Client struct {
	clientID   string
	tokenMgr   *TokenManager
	httpClient *http.Client
}

// NewClient creates a Twitch Helix client.
func NewClient(clientID string, tokenMgr *TokenManager) *Client {
	return &Client{
		clientID: clientID,
		tokenMgr: tokenMgr,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type gamesResponse struct {
	Data []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"data"`
}

// GetGameIDs resolves game names to Twitch game IDs.
// Returns a map of name→id for names that were found.
func (c *Client) GetGameIDs(ctx context.Context, names []string) (map[string]string, error) {
	if len(names) == 0 {
		return nil, nil
	}

	params := url.Values{}
	for _, n := range names {
		params.Add("name", n)
	}

	body, err := c.get(ctx, "/games?"+params.Encode())
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var resp gamesResponse
	if err := json.UnmarshalRead(body, &resp); err != nil {
		return nil, fmt.Errorf("decode games response: %w", err)
	}

	result := make(map[string]string, len(resp.Data))
	for _, g := range resp.Data {
		result[g.Name] = g.ID
	}
	return result, nil
}

type streamsResponse struct {
	Data       []models.TwitchStream `json:"data"`
	Pagination struct {
		Cursor string `json:"cursor"`
	} `json:"pagination"`
}

// GetStreams fetches live streams matching the given game IDs and/or user logins.
// At most 100 game IDs and 100 user logins per call (Twitch API limit).
func (c *Client) GetStreams(ctx context.Context, gameIDs, userLogins []string) ([]models.TwitchStream, error) {
	params := url.Values{"first": {"100"}}
	for _, id := range gameIDs {
		params.Add("game_id", id)
	}
	for _, login := range userLogins {
		params.Add("user_login", login)
	}

	var all []models.TwitchStream
	cursor := ""

	for {
		if cursor != "" {
			params.Set("after", cursor)
		}

		body, err := c.get(ctx, "/streams?"+params.Encode())
		if err != nil {
			return nil, err
		}

		var resp streamsResponse
		if err := json.UnmarshalRead(body, &resp); err != nil {
			body.Close()
			return nil, fmt.Errorf("decode streams response: %w", err)
		}
		body.Close()

		all = append(all, resp.Data...)

		if resp.Pagination.Cursor == "" || len(resp.Data) == 0 {
			break
		}
		cursor = resp.Pagination.Cursor
	}

	return all, nil
}

// get performs a GET request against the Helix API, retrying once on 401.
func (c *Client) get(ctx context.Context, path string) (io.ReadCloser, error) {
	token, err := c.tokenMgr.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	resp, err := c.doGet(ctx, helixBase+path, token)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		token, err = c.tokenMgr.Refresh(ctx)
		if err != nil {
			return nil, fmt.Errorf("token refresh after 401: %w", err)
		}
		resp, err = c.doGet(ctx, helixBase+path, token)
		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("helix GET %s: status %d", path, resp.StatusCode)
	}

	return resp.Body, nil
}

func (c *Client) doGet(ctx context.Context, fullURL, token string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Client-ID", c.clientID)
	req.Header.Set("Authorization", "Bearer "+token)
	return c.httpClient.Do(req)
}
