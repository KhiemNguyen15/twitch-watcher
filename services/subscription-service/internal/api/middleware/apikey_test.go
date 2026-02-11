package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/khiemnguyen15/twitch-watcher/services/subscription-service/internal/api/middleware"
)

const testKey = "secret-key"

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestInternalAPIKey_ValidKey(t *testing.T) {
	handler := middleware.InternalAPIKey(testKey)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Internal-API-Key", testKey)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestInternalAPIKey_WrongKey(t *testing.T) {
	handler := middleware.InternalAPIKey(testKey)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Internal-API-Key", "wrong-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestInternalAPIKey_MissingHeader(t *testing.T) {
	handler := middleware.InternalAPIKey(testKey)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}
