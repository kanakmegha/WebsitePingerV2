package handler

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/monitor"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/monitorcheckconfig"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/tenant"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/tenantsetting"
	"github.com/kanakmegha/WebsitePingerV2/internal/middleware"
)

type SettingsHandler struct {
	client *ent.Client
}

func NewSettingsHandler(client *ent.Client) *SettingsHandler {
	return &SettingsHandler{client: client}
}

type UpdateSettingsRequest struct {
	HTTPIntervalSeconds      int `json:"http_interval_seconds"`
	DNSIntervalSeconds       int `json:"dns_interval_seconds"`
	SSLIntervalSeconds       int `json:"ssl_interval_seconds"`
	DomainIntervalSeconds    int `json:"domain_interval_seconds"`
	EmailAuthIntervalSeconds int `json:"email_auth_interval_seconds"`
	SSLMinExpiryDays         int `json:"ssl_min_expiry_days"`
	DomainMinExpiryDays      int `json:"domain_min_expiry_days"`
}

// GetSettings retrieves the tenant_settings for the authenticated tenant.
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := middleware.GetTenantID(r.Context())

	if tenantID == uuid.Nil {
		log.Println("[SETTINGS GET] Unauthorized: Nil tenant ID in context")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized: tenant ID missing from token"})
		return
	}

	s, err := h.getOrCreateTenantSettings(r.Context(), tenantID)
	if err != nil {
		log.Printf("[SETTINGS GET] Error fetching settings for tenant %s: %v\n", tenantID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch tenant settings: " + err.Error()})
		return
	}

	json.NewEncoder(w).Encode(s)
}

// UpdateSettings validates and updates the tenant_settings, verifying tenant existence prior to mutation.
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)

	log.Println("==== SETTINGS UPDATE START ====")
	log.Println("[SETTINGS UPDATE] Context Tenant ID:", tenantID)

	if tenantID == uuid.Nil {
		log.Println("[SETTINGS UPDATE] Validation Error: Nil/Missing Tenant ID")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized: invalid or missing tenant ID"})
		return
	}

	// 1. VERIFY TENANT EXISTS BEFORE PROCEEDING
	tenantExists, err := h.client.Tenant.Query().Where(tenant.ID(tenantID)).Exist(ctx)
	if err != nil {
		log.Printf("[SETTINGS UPDATE] Error checking tenant existence (%s): %v\n", tenantID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error validating tenant existence"})
		return
	}
	if !tenantExists {
		log.Printf("[SETTINGS UPDATE] Tenant ID %s does not exist in tenants table. Failing request gracefully.\n", tenantID)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "tenant record does not exist"})
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[SETTINGS UPDATE] Failed to read request body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to read request body"})
		return
	}

	var req UpdateSettingsRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Printf("[SETTINGS UPDATE] JSON Decode Error: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid payload format"})
		return
	}

	log.Printf("[SETTINGS UPDATE] Parsed Request Payload: %+v\n", req)

	// Validation rules
	if req.HTTPIntervalSeconds < 30 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "HTTP interval must be at least 30 seconds"})
		return
	}
	if req.DNSIntervalSeconds < 300 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "DNS interval must be at least 300 seconds"})
		return
	}
	if req.SSLIntervalSeconds < 3600 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "SSL interval must be at least 3600 seconds"})
		return
	}
	if req.DomainIntervalSeconds < 86400 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Domain interval must be at least 86400 seconds"})
		return
	}
	if req.EmailAuthIntervalSeconds < 300 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email Auth interval must be at least 300 seconds"})
		return
	}
	if req.SSLMinExpiryDays < 1 {
		req.SSLMinExpiryDays = 30
	}
	if req.DomainMinExpiryDays < 1 {
		req.DomainMinExpiryDays = 30
	}

	tx, err := h.client.Tx(ctx)
	if err != nil {
		log.Printf("[SETTINGS UPDATE] DB Tx start error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database transaction failed: " + err.Error()})
		return
	}
	defer tx.Rollback()

	// Safe Upsert TenantSetting via Transaction
	existingSettings, err := tx.TenantSetting.Query().
		Where(tenantsetting.TenantIDEQ(tenantID)).
		Only(ctx)

	var ts *ent.TenantSetting
	if ent.IsNotFound(err) {
		log.Printf("[SETTINGS UPDATE] Creating new TenantSetting row for tenant %s\n", tenantID)
		ts, err = tx.TenantSetting.Create().
			SetTenantID(tenantID).
			SetHTTPIntervalSeconds(req.HTTPIntervalSeconds).
			SetDNSIntervalSeconds(req.DNSIntervalSeconds).
			SetSslIntervalSeconds(req.SSLIntervalSeconds).
			SetDomainIntervalSeconds(req.DomainIntervalSeconds).
			SetEmailAuthIntervalSeconds(req.EmailAuthIntervalSeconds).
			SetSslMinExpiryDays(req.SSLMinExpiryDays).
			SetDomainMinExpiryDays(req.DomainMinExpiryDays).
			SetUpdatedAt(time.Now()).
			Save(ctx)
	} else if err == nil {
		log.Printf("[SETTINGS UPDATE] Updating existing TenantSetting row ID %s for tenant %s\n", existingSettings.ID, tenantID)
		ts, err = tx.TenantSetting.UpdateOneID(existingSettings.ID).
			SetHTTPIntervalSeconds(req.HTTPIntervalSeconds).
			SetDNSIntervalSeconds(req.DNSIntervalSeconds).
			SetSslIntervalSeconds(req.SSLIntervalSeconds).
			SetDomainIntervalSeconds(req.DomainIntervalSeconds).
			SetEmailAuthIntervalSeconds(req.EmailAuthIntervalSeconds).
			SetSslMinExpiryDays(req.SSLMinExpiryDays).
			SetDomainMinExpiryDays(req.DomainMinExpiryDays).
			SetUpdatedAt(time.Now()).
			Save(ctx)
	}

	if err != nil {
		log.Printf("[SETTINGS UPDATE] Failed to save/update TenantSetting for tenant %s: %v\n", tenantID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to persist tenant settings: " + err.Error()})
		return
	}

	// Retroactively update all existing monitor_check_configs for this tenant's monitors
	monitors, err := tx.Monitor.Query().
		Where(monitor.TenantID(tenantID)).
		Select(monitor.FieldID).
		All(ctx)

	if err == nil && len(monitors) > 0 {
		monitorIDs := make([]uuid.UUID, len(monitors))
		for i, m := range monitors {
			monitorIDs[i] = m.ID
		}

		tx.MonitorCheckConfig.Update().
			Where(monitorcheckconfig.MonitorIDIn(monitorIDs...), monitorcheckconfig.CheckTypeEQ(monitorcheckconfig.CheckTypeHTTP)).
			SetIntervalSeconds(req.HTTPIntervalSeconds).
			Exec(ctx)

		tx.MonitorCheckConfig.Update().
			Where(monitorcheckconfig.MonitorIDIn(monitorIDs...), monitorcheckconfig.CheckTypeEQ(monitorcheckconfig.CheckTypeDNS)).
			SetIntervalSeconds(req.DNSIntervalSeconds).
			Exec(ctx)

		tx.MonitorCheckConfig.Update().
			Where(monitorcheckconfig.MonitorIDIn(monitorIDs...), monitorcheckconfig.CheckTypeEQ(monitorcheckconfig.CheckTypeSsl)).
			SetIntervalSeconds(req.SSLIntervalSeconds).
			Exec(ctx)

		tx.MonitorCheckConfig.Update().
			Where(monitorcheckconfig.MonitorIDIn(monitorIDs...), monitorcheckconfig.CheckTypeEQ(monitorcheckconfig.CheckTypeDomain)).
			SetIntervalSeconds(req.DomainIntervalSeconds).
			Exec(ctx)

		tx.MonitorCheckConfig.Update().
			Where(monitorcheckconfig.MonitorIDIn(monitorIDs...), monitorcheckconfig.CheckTypeEQ(monitorcheckconfig.CheckTypeEmailAuth)).
			SetIntervalSeconds(req.EmailAuthIntervalSeconds).
			Exec(ctx)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[SETTINGS UPDATE] Failed to commit transaction for tenant %s: %v\n", tenantID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to commit settings update"})
		return
	}

	log.Printf("[SETTINGS UPDATE] Successfully updated settings for tenant %s\n", tenantID)
	json.NewEncoder(w).Encode(ts)
}

func (h *SettingsHandler) getOrCreateTenantSettings(ctx context.Context, tenantID uuid.UUID) (*ent.TenantSetting, error) {
	s, err := h.client.TenantSetting.Query().
		Where(tenantsetting.TenantID(tenantID)).
		Only(ctx)
	if err == nil {
		return s, nil
	}

	// Create default tenant settings
	return h.client.TenantSetting.Create().
		SetTenantID(tenantID).
		SetHTTPIntervalSeconds(60).
		SetDNSIntervalSeconds(300).
		SetSslIntervalSeconds(3600).
		SetDomainIntervalSeconds(86400).
		SetEmailAuthIntervalSeconds(300).
		SetSslMinExpiryDays(30).
		SetDomainMinExpiryDays(30).
		Save(ctx)
}
