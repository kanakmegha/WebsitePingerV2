package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/alert"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/membership"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/monitor"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/monitorcheck"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/monitorcheckconfig"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/tenantsetting"
	"github.com/kanakmegha/WebsitePingerV2/internal/middleware"
)

type MonitorHandler struct {
	client *ent.Client
}

func NewMonitorHandler(client *ent.Client) *MonitorHandler {
	return &MonitorHandler{client: client}
}

type CheckConfigPayload struct {
	Type            string `json:"type"`
	IntervalSeconds int    `json:"interval_seconds"`
	TimeoutSeconds  int    `json:"timeout_seconds,omitempty"`
}

type CreateMonitorRequest struct {
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Checks []string `json:"checks"`
}

func (h *MonitorHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	monitors, err := h.client.Monitor.Query().
		Where(monitor.TenantID(tenantID)).
		WithCheckConfigs().
		WithChecks(func(q *ent.MonitorCheckQuery) {
			q.Order(ent.Desc(monitorcheck.FieldCheckedAt)).
				WithHTTPResult().
				WithSslResult().
				WithDomainResult().
				WithDNSResult()
		}).
		All(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch monitors"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(monitors)
}

func (h *MonitorHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := middleware.GetTenantID(r.Context())
	monitorIDStr := chi.URLParam(r, "id")
	monitorID, err := uuid.Parse(monitorIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid monitor id"})
		return
	}

	m, err := h.client.Monitor.Query().
		Where(
			monitor.ID(monitorID),
			monitor.TenantID(tenantID),
		).
		WithCheckConfigs().
		WithChecks(func(q *ent.MonitorCheckQuery) {
			q.Order(ent.Desc(monitorcheck.FieldCheckedAt)).
				WithHTTPResult().
				WithSslResult().
				WithDomainResult().
				WithDNSResult()
		}).
		Only(r.Context())

	if err != nil || m == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "monitor not found"})
		return
	}

	json.NewEncoder(w).Encode(m)
}

func (h *MonitorHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := middleware.GetTenantID(r.Context())

	var req CreateMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid payload format"})
		return
	}

	if req.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "URL is required"})
		return
	}

	if len(req.Checks) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "At least one check configuration must be specified"})
		return
	}

	ctx := r.Context()
	now := time.Now()

	// Fetch or create tenant settings for interval inheritance
	ts, err := h.client.TenantSetting.Query().
		Where(tenantsetting.TenantID(tenantID)).
		Only(ctx)
	if err != nil {
		ts, err = h.client.TenantSetting.Create().
			SetTenantID(tenantID).
			SetHTTPIntervalSeconds(60).
			SetDNSIntervalSeconds(300).
			SetSslIntervalSeconds(3600).
			SetDomainIntervalSeconds(86400).
			SetEmailAuthIntervalSeconds(300).
			Save(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to load tenant settings: " + err.Error()})
			return
		}
	}

	// Extract domain from URL
	parsedURL, err := url.Parse(req.URL)
	domain := req.URL
	if err == nil && parsedURL.Host != "" {
		domain = parsedURL.Host
	}
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	if req.Name == "" {
		req.Name = domain
	}

	// Transactional creation of Monitor & MonitorCheckConfig records
	tx, err := h.client.Tx(ctx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "database transaction error"})
		return
	}

	m, err := tx.Monitor.Create().
		SetTenantID(tenantID).
		SetName(req.Name).
		SetURL(req.URL).
		SetDomain(domain).
		SetType(monitor.TypeHTTP).
		SetIsActive(true).
		Save(ctx)

	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create monitor: " + err.Error()})
		return
	}

	for _, checkStr := range req.Checks {
		cType := strings.ToLower(checkStr)
		var interval int
		switch cType {
		case "http":
			interval = ts.HTTPIntervalSeconds
		case "dns":
			interval = ts.DNSIntervalSeconds
		case "ssl":
			interval = ts.SslIntervalSeconds
		case "domain":
			interval = ts.DomainIntervalSeconds
		case "email_auth":
			interval = ts.EmailAuthIntervalSeconds
		default:
			interval = 60
		}

		checkEnum := monitorcheckconfig.CheckType(cType)
		_, err := tx.MonitorCheckConfig.Create().
			SetMonitorID(m.ID).
			SetCheckType(checkEnum).
			SetIntervalSeconds(interval).
			SetTimeoutSeconds(10).
			SetIsEnabled(true).
			SetNextCheckAt(now).
			Save(ctx)

		if err != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to create check config for %s: %v", cType, err)})
			return
		}
	}

	// Create default Alert rules for the monitor
	defaultAlerts := []alert.Type{alert.TypeDown, alert.TypeSslExpiring, alert.TypeDNSChanged, alert.TypeDomainExpiring}
	for _, aType := range defaultAlerts {
		_, _ = tx.Alert.Create().
			SetTenantID(tenantID).
			SetMonitorID(m.ID).
			SetType(aType).
			SetThreshold(1).
			Save(ctx)
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to commit transaction"})
		return
	}

	// Query created monitor with check configs
	createdMon, _ := h.client.Monitor.Query().
		Where(monitor.ID(m.ID), monitor.TenantID(tenantID)).
		WithCheckConfigs().
		Only(ctx)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdMon)
}

func (h *MonitorHandler) GetChecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := middleware.GetTenantID(r.Context())
	monitorIDStr := chi.URLParam(r, "id")
	monitorID, err := uuid.Parse(monitorIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid monitor id"})
		return
	}

	// Validate monitor belongs to tenant first
	m, err := h.client.Monitor.Query().
		Where(monitor.ID(monitorID), monitor.TenantID(tenantID)).
		Only(r.Context())
	if err != nil || m == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "monitor not found"})
		return
	}

	// Optional check_type query filter
	checkTypeParam := r.URL.Query().Get("type")

	// Fetch fresh check records directly from PostgreSQL
	query := h.client.MonitorCheck.Query().
		Where(monitorcheck.MonitorID(m.ID)).
		WithHTTPResult().
		WithSslResult().
		WithDomainResult().
		WithDNSResult()

	if checkTypeParam != "" {
		query = query.Where(monitorcheck.CheckTypeEQ(monitorcheck.CheckType(checkTypeParam)))
	}

	checks, err := query.
		Order(ent.Desc(monitorcheck.FieldCheckedAt)).
		Limit(50).
		All(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch checks"})
		return
	}

	json.NewEncoder(w).Encode(checks)
}

// Delete deletes a single monitor by ID using PostgreSQL ON DELETE CASCADE.
func (h *MonitorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	userID := middleware.GetUserID(ctx)

	monitorIDStr := chi.URLParam(r, "id")
	monitorID, err := uuid.Parse(monitorIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid monitor id"})
		return
	}

	// Delete monitor safely
	if err := DeleteMonitor(ctx, h.client, userID, tenantID, monitorID); err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "monitor deleted successfully"})
}

// DeleteMonitor deletes a single monitor by ID safely.
func DeleteMonitor(ctx context.Context, client *ent.Client, userID uuid.UUID, tenantID uuid.UUID, monitorID uuid.UUID) error {
	hasPerm, err := client.Membership.Query().
		Where(
			membership.UserID(userID),
			membership.TenantID(tenantID),
		).
		Exist(ctx)

	if err != nil || !hasPerm {
		return fmt.Errorf("forbidden: user does not have membership access to this tenant")
	}

	m, err := client.Monitor.Query().
		Where(
			monitor.ID(monitorID),
			monitor.TenantID(tenantID),
		).
		Only(ctx)

	if err != nil || m == nil {
		return fmt.Errorf("monitor not found")
	}

	return client.Monitor.DeleteOneID(m.ID).Exec(ctx)
}

// DeleteTenant deletes an entire tenant using PostgreSQL ON DELETE CASCADE.
func DeleteTenant(ctx context.Context, client *ent.Client, userID uuid.UUID, tenantID uuid.UUID) error {
	mem, err := client.Membership.Query().
		Where(
			membership.UserID(userID),
			membership.TenantID(tenantID),
		).
		Only(ctx)

	if err != nil || mem.Role != membership.RoleOwner {
		return fmt.Errorf("forbidden: only tenant owner can delete tenant")
	}

	return client.Tenant.DeleteOneID(tenantID).Exec(ctx)
}

// DeleteUser deletes user memberships and the user entity safely.
func DeleteUser(ctx context.Context, client *ent.Client, userID uuid.UUID) error {
	return client.User.DeleteOneID(userID).Exec(ctx)
}
