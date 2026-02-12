package handler

import (
	"encoding/json/v2"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
	"github.com/khiemnguyen15/twitch-watcher/services/subscription-service/internal/service"
)

// SubscriptionHandler handles public subscription endpoints.
type SubscriptionHandler struct {
	svc *service.SubscriptionService
}

// NewSubscriptionHandler creates a SubscriptionHandler.
func NewSubscriptionHandler(svc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

type createRequest struct {
	DiscordWebhook string           `json:"discord_webhook"`
	WatchType      models.WatchType `json:"watch_type"`
	WatchTarget    string           `json:"watch_target"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Create handles POST /v1/subscriptions.
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	sub, err := h.svc.Create(r.Context(), req.DiscordWebhook, req.WatchType, req.WatchTarget)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidWebhook):
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid Discord webhook URL"})
		case errors.Is(err, service.ErrDuplicate):
			writeJSON(w, http.StatusConflict, errorResponse{Error: "subscription already exists"})
		default:
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, sub)
}

// GetByID handles GET /v1/subscriptions/{id}.
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid subscription ID"})
		return
	}

	sub, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "subscription not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, sub)
}

// Delete handles DELETE /v1/subscriptions/{id}.
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid subscription ID"})
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "subscription not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.MarshalWrite(w, v)
}
