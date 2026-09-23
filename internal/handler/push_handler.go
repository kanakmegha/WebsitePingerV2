package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/pushsubscription"
	"github.com/kanakmegha/WebsitePingerV2/internal/middleware"
	"github.com/kanakmegha/WebsitePingerV2/internal/push"
)

type PushHandler struct {
	client  *ent.Client
	pushSvc *push.PushService
}

func NewPushHandler(client *ent.Client, pushSvc *push.PushService) *PushHandler {
	return &PushHandler{
		client:  client,
		pushSvc: pushSvc,
	}
}

type SubscribePayload struct {
	Endpoint string `json:"endpoint"`
	P256dh   string `json:"p256dh"`
	Auth     string `json:"auth"`
}

type UnsubscribePayload struct {
	Endpoint string `json:"endpoint"`
}

func (h *PushHandler) GetVAPIDKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pubKey := h.pushSvc.GetPublicKey()
	json.NewEncoder(w).Encode(map[string]string{
		"public_key": pubKey,
	})
}

func (h *PushHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	tenantID := middleware.GetTenantID(ctx)
	if userID == uuid.Nil || tenantID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized context"})
		return
	}

	var req SubscribePayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Endpoint == "" || req.P256dh == "" || req.Auth == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid push subscription payload"})
		return
	}

	// Delete existing subscription for this endpoint if present
	_, _ = h.client.PushSubscription.Delete().
		Where(
			pushsubscription.UserID(userID),
			pushsubscription.Endpoint(req.Endpoint),
		).
		Exec(ctx)

	// Save new PushSubscription record
	_, err := h.client.PushSubscription.Create().
		SetUserID(userID).
		SetTenantID(tenantID).
		SetEndpoint(req.Endpoint).
		SetP256dh(req.P256dh).
		SetAuth(req.Auth).
		Save(ctx)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save push subscription"})
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *PushHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized context"})
		return
	}

	var req UnsubscribePayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Endpoint == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Endpoint is required"})
		return
	}

	_, err := h.client.PushSubscription.Delete().
		Where(
			pushsubscription.UserID(userID),
			pushsubscription.Endpoint(req.Endpoint),
		).
		Exec(ctx)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to unsubscribe"})
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
