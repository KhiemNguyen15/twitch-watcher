package api

import (
	"net/http"

	"github.com/khiemnguyen15/twitch-watcher/services/subscription-service/internal/api/handler"
	"github.com/khiemnguyen15/twitch-watcher/services/subscription-service/internal/api/middleware"
	"github.com/khiemnguyen15/twitch-watcher/services/subscription-service/internal/service"
)

// NewRouter builds and returns the HTTP mux for the subscription service.
func NewRouter(svc *service.SubscriptionService, internalAPIKey string) http.Handler {
	mux := http.NewServeMux()

	subHandler := handler.NewSubscriptionHandler(svc)
	intHandler := handler.NewInternalHandler(svc)

	// Public routes
	mux.HandleFunc("POST /v1/subscriptions", subHandler.Create)
	mux.HandleFunc("GET /v1/subscriptions/{id}", subHandler.GetByID)
	mux.HandleFunc("DELETE /v1/subscriptions/{id}", subHandler.Delete)

	// Health check
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Internal routes (API-key protected)
	internalMux := http.NewServeMux()
	internalMux.HandleFunc("GET /internal/subscriptions/active", intHandler.ListActive)

	mux.Handle("/internal/", middleware.InternalAPIKey(internalAPIKey)(internalMux))

	return mux
}
