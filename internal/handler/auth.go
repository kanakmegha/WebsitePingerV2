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
	Email      string `json:"email"`
	Password   string `json:"password"`
	TenantName string `json:"tenant_name,omitempty"`
}

type TenantDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Role string    `json:"role"`
}

type AuthResponse struct {
	Token    string      `json:"token"`
	UserID   uuid.UUID   `json:"user_id"`
	TenantID uuid.UUID   `json:"tenant_id,omitempty"`
	Tenants  []TenantDTO `json:"tenants,omitempty"`
}

// Register creates Tenant, User, TenantSetting, and Membership atomically in a single database transaction.
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

	// Begin Atomic Database Transaction
	tx, err := h.client.Tx(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "transaction initialization failed"})
		return
	}
	defer tx.Rollback()

	// 1. Create User
	u, err := tx.User.Create().
		SetEmail(req.Email).
		SetPasswordHash(req.Password).
		Save(ctx)
	if err != nil {
		log.Printf("[AUTH REGISTER] Step 1 Failed - User creation error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create user"})
		return
	}

	// Determine organization name (use provided name or default to Email's Organization)
	orgName := req.TenantName
	if orgName == "" {
		orgName = req.Email + "'s Organization"
	}

	// 2. Create Tenant
	t, err := tx.Tenant.Create().
		SetName(orgName).
		SetPlan("free").
		Save(ctx)
	if err != nil {
		log.Printf("[AUTH REGISTER] Step 2 Failed - Tenant creation error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create organization"})
		return
	}

	// 3. Create Default Tenant Settings
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
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create organization settings"})
		return
	}

	// 4. Create User-Tenant Membership (Role = Owner)
	_, err = tx.Membership.Create().
		SetUserID(u.ID).
		SetTenantID(t.ID).
		SetRole(membership.RoleOwner).
		Save(ctx)
	if err != nil {
		log.Printf("[AUTH REGISTER] Step 4 Failed - Membership creation error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create organization membership"})
		return
	}

	// Commit Transaction atomically
	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to commit atomic registration transaction"})
		return
	}

	// Issue JWT containing User ID and primary Tenant ID
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

// Login retrieves existing user and all associated tenant memberships (without auto-provisioning).
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

	// Retrieve user tenant memberships
	memberships, _ := h.client.Membership.Query().
		Where(membership.UserID(u.ID)).
		WithTenant().
		All(ctx)

	tenantDTOs := make([]TenantDTO, 0, len(memberships))
	var primaryTenantID uuid.UUID

	for i, m := range memberships {
		if m.Edges.Tenant != nil {
			tExists, _ := h.client.Tenant.Query().Where(tenant.IDEQ(m.TenantID)).Exist(ctx)
			if tExists {
				if i == 0 {
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

	tokenStr, _ := generateToken(u.ID, primaryTenantID)

	json.NewEncoder(w).Encode(AuthResponse{
		Token:    tokenStr,
		UserID:   u.ID,
		TenantID: primaryTenantID,
		Tenants:  tenantDTOs,
	})
}

type CreateOrgRequest struct {
	Name string `json:"name"`
}

type JoinOrgRequest struct {
	TenantID string `json:"tenant_id"`
}

// CreateOrganization creates a new Tenant, TenantSetting, and sets the current User as Owner.
func (h *AuthHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r.Context())

	if userID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	var req CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "organization name is required"})
		return
	}

	ctx := r.Context()
	tx, err := h.client.Tx(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	t, err := tx.Tenant.Create().
		SetName(req.Name).
		SetPlan("free").
		Save(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create tenant"})
		return
	}

	_, err = tx.TenantSetting.Create().
		SetTenantID(t.ID).
		SetHTTPIntervalSeconds(60).
		SetDNSIntervalSeconds(300).
		SetSslIntervalSeconds(3600).
		SetDomainIntervalSeconds(86400).
		SetEmailAuthIntervalSeconds(300).
		Save(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create tenant settings"})
		return
	}

	_, err = tx.Membership.Create().
		SetUserID(userID).
		SetTenantID(t.ID).
		SetRole(membership.RoleOwner).
		Save(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create membership"})
		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to commit transaction"})
		return
	}

	dto := TenantDTO{
		ID:   t.ID,
		Name: t.Name,
		Role: "owner",
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto)
}

// JoinOrganization binds a user to an existing organization ID with member role.
func (h *AuthHandler) JoinOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r.Context())

	if userID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	var req JoinOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TenantID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "valid tenant_id is required"})
		return
	}

	targetTenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid tenant_id UUID format"})
		return
	}

	ctx := r.Context()
	t, err := h.client.Tenant.Query().Where(tenant.IDEQ(targetTenantID)).Only(ctx)
	if err != nil || t == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "organization not found"})
		return
	}

	// Check if membership already exists
	mExists, _ := h.client.Membership.Query().
		Where(membership.UserID(userID), membership.TenantID(targetTenantID)).
		Exist(ctx)
	if mExists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "already a member of this organization"})
		return
	}

	m, err := h.client.Membership.Create().
		SetUserID(userID).
		SetTenantID(targetTenantID).
		SetRole(membership.RoleMember).
		Save(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to join organization"})
		return
	}

	dto := TenantDTO{
		ID:   t.ID,
		Name: t.Name,
		Role: string(m.Role),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto)
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
