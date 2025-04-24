package twitch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testClientID     = "test_client_id"
	testClientSecret = "test_client_secret"
	testAccessToken  = "test_access_token"
	testExpiresIn    = 3600 // seconds
)

func mockTokenServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	// Override the fetchNewToken URL temporarily for testing
	originalURL := fetchURL
	fetchURL = server.URL
	t.Cleanup(func() {
		server.Close()
		fetchURL = originalURL // Restore original URL after test completion.
	})
	return server
}

func TestNewTokenManager(t *testing.T) {
	tm := NewTokenManager(testClientID, testClientSecret)
	require.NotNil(t, tm)
	assert.Equal(t, testClientID, tm.clientID)
	assert.Equal(t, testClientSecret, tm.clientSecret)
	assert.Empty(t, tm.accessToken)
	assert.True(t, tm.expiry.IsZero())
}

func TestGetAccessToken_Success_FirstFetch(t *testing.T) {
	_ = mockTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		err := r.ParseForm()
		require.NoError(t, err)
		assert.Equal(t, testClientID, r.FormValue("client_id"))
		assert.Equal(t, testClientSecret, r.FormValue("client_secret"))
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: testAccessToken,
			ExpiresIn:   testExpiresIn,
		})
	})

	tm := NewTokenManager(testClientID, testClientSecret)
	token, err := tm.GetAccessToken()

	require.NoError(t, err)
	assert.Equal(t, testAccessToken, token)
	assert.Equal(t, testAccessToken, tm.accessToken)                                                             // Check internal state
	assert.WithinDuration(t, time.Now().Add(time.Duration(testExpiresIn)*time.Second), tm.expiry, 5*time.Second) // Allow some leeway
}

func TestGetAccessToken_Success_CachedToken(t *testing.T) {
	tm := NewTokenManager(testClientID, testClientSecret)
	// Pre-populate the token manager with a valid token
	tm.accessToken = testAccessToken
	tm.expiry = time.Now().Add(1 * time.Hour)

	token, err := tm.GetAccessToken()

	require.NoError(t, err)
	assert.Equal(t, testAccessToken, token)
}

func TestGetAccessToken_Success_ExpiredTokenRefetch(t *testing.T) {
	fetchCount := 0
	newToken := "new_test_access_token"
	newExpiry := 1800

	_ = mockTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: newToken,
			ExpiresIn:   newExpiry,
		})
	})

	tm := NewTokenManager(testClientID, testClientSecret)
	// Set an expired token
	tm.accessToken = "expired_token"
	tm.expiry = time.Now().Add(-1 * time.Minute)

	token, err := tm.GetAccessToken()

	require.NoError(t, err)
	assert.Equal(t, newToken, token)
	assert.Equal(t, 1, fetchCount, "Expected fetchNewToken to be called once")
	assert.Equal(t, newToken, tm.accessToken)
	assert.WithinDuration(t, time.Now().Add(time.Duration(newExpiry)*time.Second), tm.expiry, 5*time.Second)
}

func TestGetAccessToken_Error_FetchFailed_StatusNotOK(t *testing.T) {
	_ = mockTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
	})

	tm := NewTokenManager(testClientID, testClientSecret)
	token, err := tm.GetAccessToken()

	require.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "failed to fetch token: status 500")
	assert.Empty(t, tm.accessToken) // Ensure internal state wasn't updated
	assert.True(t, tm.expiry.IsZero())
}

func TestGetAccessToken_Error_FetchFailed_EmptyToken(t *testing.T) {
	_ = mockTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "",
			ExpiresIn:   testExpiresIn,
		})
	})

	tm := NewTokenManager(testClientID, testClientSecret)
	token, err := tm.GetAccessToken()

	require.Error(t, err)
	assert.Empty(t, token)
	assert.EqualError(t, err, "received empty access token")
	assert.Empty(t, tm.accessToken)
	assert.True(t, tm.expiry.IsZero())
}

func TestGetAccessToken_Error_FetchFailed_InvalidJson(t *testing.T) {
	_ = mockTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"access_token": "token", "expires_in": "not_an_int"}`) // Invalid expires_in type
	})

	tm := NewTokenManager(testClientID, testClientSecret)
	token, err := tm.GetAccessToken()

	require.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "failed to decode token response") // Check for decode error
	assert.Empty(t, tm.accessToken)
	assert.True(t, tm.expiry.IsZero())
}

func TestGetAccessToken_ConcurrentAccess(t *testing.T) {
	fetchCount := 0
	_ = mockTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Simulate network delay to increase chance of race condition without mutex
		time.Sleep(10 * time.Millisecond)
		fetchCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: testAccessToken,
			ExpiresIn:   testExpiresIn,
		})
	})

	tm := NewTokenManager(testClientID, testClientSecret)
	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Use a channel to signal all goroutines to start roughly simultaneously.
	startSignal := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			<-startSignal                     // Wait for the signal before proceeding.
			token, err := tm.GetAccessToken() // All goroutines call this concurrently.
			assert.NoError(t, err)
			assert.Equal(t, testAccessToken, token)
		}()
	}

	close(startSignal) // Signal all goroutines to start
	wg.Wait()

	// Verify that the mutex prevented multiple token fetches despite concurrent calls.
	assert.Equal(t, 1, fetchCount, "fetchNewToken should only be called once during concurrent access")
	assert.Equal(t, testAccessToken, tm.accessToken)
	assert.WithinDuration(t, time.Now().Add(time.Duration(testExpiresIn)*time.Second), tm.expiry, 5*time.Second)
}
