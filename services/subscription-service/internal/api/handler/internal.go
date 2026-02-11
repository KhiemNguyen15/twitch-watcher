package handler

import (
	"encoding/json/v2"
	"net/http"

	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
	"github.com/khiemnguyen15/twitch-watcher/services/subscription-service/internal/service"
)

// InternalHandler handles internal endpoints consumed by stream-poller.
type InternalHandler struct {
	svc *service.SubscriptionService
}

// NewInternalHandler creates an InternalHandler.
func NewInternalHandler(svc *service.SubscriptionService) *InternalHandler {
	return &InternalHandler{svc: svc}
}

type activeSubscriptionsResponse struct {
	Subscriptions []models.Subscription `json:"subscriptions"`
	Total         int                   `json:"total"`
}

// ListActive handles GET /internal/subscriptions/active.
func (h *InternalHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	subs, err := h.svc.ListActive(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.MarshalWrite(w, errorResponse{Error: "internal error"})
		return
	}

	if subs == nil {
		subs = []models.Subscription{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.MarshalWrite(w, activeSubscriptionsResponse{
		Subscriptions: subs,
		Total:         len(subs),
	})
}
