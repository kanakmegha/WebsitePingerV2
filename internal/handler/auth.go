package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/membership"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/tenant"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/user"
	"github.com/kanakmegha/WebsitePingerV2/internal/middleware"
)

type AuthHandler struct {
	client *ent.Client
}

func NewAuthHandler(client *ent.Client) *AuthHandler {
	return &AuthHandler{client: client}
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TenantDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Role string    `json:"role"`
}

type AuthResponse struct {
	Token    string      `json:"token"`
	UserID   uuid.UUID   `json:"user_id"`
	TenantID uuid.UUID   `json:"tenant_id"`
	Tenants  []TenantDTO `json:"tenants,omitempty"`
}

// Register creates Tenant, User, TenantSetting, and Membership atomically.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid payload"})
		return
	}

	ctx := r.Context()

	// Check if user already exists
	exists, err := h.client.User.Query().Where(user.Email(req.Email)).Exist(ctx)
	if err != nil || exists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "user already exists"})
		return
	}

	// Begin Atomic Transaction
	tx, err := h.client.Tx(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "transaction initialization failed"})
		return
	}
	defer tx.Rollback()

	// 1. CREATE TENANT
	t, err := tx.Tenant.Create().
		SetName(req.Email + "'s Organization").
		SetPlan("free").
		Save(ctx)
	if err != nil {
		log.Printf("[AUTH REGISTER] Step 1 Failed - Tenant creation error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create tenant"})
		return
	}

	// 2. CREATE USER
	u, err := tx.User.Create().
		SetEmail(req.Email).
		SetPasswordHash(req.Password).
		Save(ctx)
	if err != nil {
		log.Printf("[AUTH REGISTER] Step 2 Failed - User creation error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create user"})
		return
	}

	// 3. CREATE DEFAULT TENANT SETTINGS
	_, err = tx.TenantSetting.Create().
		SetTenantID(t.ID).
		SetHTTPIntervalSeconds(60).
		SetDNSIntervalSeconds(300).
		SetSslIntervalSeconds(3600).
		SetDomainIntervalSeconds(86400).
		SetEmailAuthIntervalSeconds(300).
		Save(ctx)
	if err != nil {
		log.Printf("[AUTH REGISTER] Step 3 Failed - TenantSetting creation error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create tenant settings"})
		return
	}

	// 4. CREATE USER-TENANT MEMBERSHIP (Role = Owner)
	_, err = tx.Membership.Create().
		SetUserID(u.ID).
		SetTenantID(t.ID).
		SetRole(membership.RoleOwner).
		Save(ctx)
	if err != nil {
		log.Printf("[AUTH REGISTER] Step 4 Failed - Membership creation error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create membership"})
		return
	}

	// Commit Transaction
	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to commit transaction"})
		return
	}

	// Issue JWT containing User ID
	tokenStr, _ := generateToken(u.ID, t.ID)

	json.NewEncoder(w).Encode(AuthResponse{
		Token:    tokenStr,
		UserID:   u.ID,
		TenantID: t.ID,
		Tenants: []TenantDTO{
			{ID: t.ID, Name: t.Name, Role: "owner"},
		},
	})
}

// Login retrieves existing user and all associated tenant memberships.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid payload"})
		return
	}

	ctx := r.Context()
	u, err := h.client.User.Query().
		Where(user.Email(req.Email)).
		Only(ctx)
	if err != nil || u.PasswordHash != req.Password {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
		return
	}

	// Retrieve ALL user tenant memberships
	memberships, err := h.client.Membership.Query().
		Where(membership.UserID(u.ID)).
		WithTenant().
		All(ctx)

	if err != nil || len(memberships) == 0 {
		// Auto-provision if user has no memberships
		t, errTenant := h.client.Tenant.Create().
			SetName(req.Email + "'s Organization").
			SetPlan("free").
			Save(ctx)
		if errTenant == nil {
			h.client.TenantSetting.Create().
				SetTenantID(t.ID).
				SetHTTPIntervalSeconds(60).
				SetDNSIntervalSeconds(300).
				SetSslIntervalSeconds(3600).
				SetDomainIntervalSeconds(86400).
				SetEmailAuthIntervalSeconds(300).
				Save(ctx)

			m, _ := h.client.Membership.Create().
				SetUserID(u.ID).
				SetTenantID(t.ID).
				SetRole(membership.RoleOwner).
				Save(ctx)

			m.Edges.Tenant = t
			memberships = []*ent.Membership{m}
		} else {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "user has no tenant membership"})
			return
		}
	}

	// Build list of DTOs and verify primary tenant
	tenantDTOs := make([]TenantDTO, 0, len(memberships))
	var primaryTenantID uuid.UUID

	for i, m := range memberships {
		if m.Edges.Tenant != nil {
			// Verify tenant still exists in DB
			tExists, _ := h.client.Tenant.Query().Where(tenant.IDEQ(m.TenantID)).Exist(ctx)
			if tExists {
				if i == 0 || primaryTenantID == uuid.Nil {
					primaryTenantID = m.TenantID
				}
				tenantDTOs = append(tenantDTOs, TenantDTO{
					ID:   m.TenantID,
					Name: m.Edges.Tenant.Name,
					Role: string(m.Role),
				})
			}
		}
	}

	if primaryTenantID == uuid.Nil {
		t, _ := h.client.Tenant.Create().
			SetName(req.Email + "'s Organization").
			SetPlan("free").
			Save(ctx)
		h.client.TenantSetting.Create().
			SetTenantID(t.ID).
			SetHTTPIntervalSeconds(60).
			SetDNSIntervalSeconds(300).
			SetSslIntervalSeconds(3600).
			SetDomainIntervalSeconds(86400).
			SetEmailAuthIntervalSeconds(300).
			Save(ctx)
		h.client.Membership.Create().
			SetUserID(u.ID).
			SetTenantID(t.ID).
			SetRole(membership.RoleOwner).
			Save(ctx)

		primaryTenantID = t.ID
		tenantDTOs = append(tenantDTOs, TenantDTO{
			ID:   t.ID,
			Name: t.Name,
			Role: "owner",
		})
	}

	tokenStr, _ := generateToken(u.ID, primaryTenantID)

	json.NewEncoder(w).Encode(AuthResponse{
		Token:    tokenStr,
		UserID:   u.ID,
		TenantID: primaryTenantID,
		Tenants:  tenantDTOs,
	})
}

// GetUserTenants handles GET /api/me/tenants to return available tenant memberships for tenant switcher.
func (h *AuthHandler) GetUserTenants(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r.Context())

	if userID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	memberships, err := h.client.Membership.Query().
		Where(membership.UserID(userID)).
		WithTenant().
		All(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to query tenant memberships"})
		return
	}

	tenantDTOs := make([]TenantDTO, 0, len(memberships))
	for _, m := range memberships {
		if m.Edges.Tenant != nil {
			tenantDTOs = append(tenantDTOs, TenantDTO{
				ID:   m.TenantID,
				Name: m.Edges.Tenant.Name,
				Role: string(m.Role),
			})
		}
	}

	json.NewEncoder(w).Encode(tenantDTOs)
}

func generateToken(userID uuid.UUID, tenantID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   userID.String(),
		"tenant_id": tenantID.String(), // Fallback claim for backward compatibility
		"exp":       time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(middleware.JWTSecret)
}
