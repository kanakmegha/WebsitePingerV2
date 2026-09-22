package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/alertevent"
	"github.com/kanakmegha/WebsitePingerV2/internal/middleware"
)

type AlertHandler struct {
	client *ent.Client
}

func NewAlertHandler(client *ent.Client) *AlertHandler {
	return &AlertHandler{client: client}
}

type AlertResponse struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	MonitorID string    `json:"monitor_id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *AlertHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized: missing tenant context"})
		return
	}

	events, err := h.client.AlertEvent.Query().
		Where(alertevent.TenantID(tenantID)).
		Order(ent.Desc(alertevent.FieldCreatedAt)).
		Limit(50).
		All(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch alert events"})
		return
	}

	res := make([]AlertResponse, 0, len(events))
	for _, e := range events {
		res = append(res, AlertResponse{
			ID:        e.ID.String(),
			TenantID:  e.TenantID.String(),
			MonitorID: e.MonitorID.String(),
			Type:      string(e.Type),
			Message:   e.Message,
			Status:    string(e.Status),
			CreatedAt: e.CreatedAt,
		})
	}

	json.NewEncoder(w).Encode(res)
}

func (h *AlertHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized: missing tenant context"})
		return
	}

	alertIDStr := chi.URLParam(r, "id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid alert ID"})
		return
	}

	_, err = h.client.AlertEvent.Update().
		Where(
			alertevent.ID(alertID),
			alertevent.TenantID(tenantID),
		).
		SetStatus(alertevent.StatusRead).
		Save(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to mark alert as read"})
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *AlertHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized: missing tenant context"})
		return
	}

	_, err := h.client.AlertEvent.Update().
		Where(
			alertevent.TenantID(tenantID),
			alertevent.StatusEQ(alertevent.StatusUnread),
		).
		SetStatus(alertevent.StatusRead).
		Save(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to mark all alerts as read"})
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
