package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kanakmegha/WebsitePingerV2/internal/email"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/invite"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/membership"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/user"
	"github.com/kanakmegha/WebsitePingerV2/internal/middleware"
)

type OrgHandler struct {
	client   *ent.Client
	emailSvc *email.Service
}

func NewOrgHandler(client *ent.Client, emailSvc *email.Service) *OrgHandler {
	return &OrgHandler{
		client:   client,
		emailSvc: emailSvc,
	}
}

type InviteUserRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type AcceptInviteRequest struct {
	Token    string `json:"token"`
	Password string `json:"password,omitempty"`
}

type InviteDetailsDTO struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantName string    `json:"tenant_name"`
	Role       string    `json:"role"`
	Status     string    `json:"status"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type InviteResponseDTO struct {
	ID        uuid.UUID     `json:"id"`
	TenantID  uuid.UUID     `json:"tenant_id"`
	Email     string        `json:"email"`
	Role      invite.Role   `json:"role"`
	Status    invite.Status `json:"status"`
	ExpiresAt time.Time     `json:"expires_at"`
	CreatedAt time.Time     `json:"created_at"`
}

func generateSecureToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return uuid.New().String()
	}
	return hex.EncodeToString(b)
}

func resolveBaseURL() string {
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}
	// Trim trailing slash to prevent double slash in generated invite URLs
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	return baseURL
}

// InviteUser handles POST /api/orgs/invite
// Creates an invite in DB with 24-hour expiry and dispatches async email with dynamic APP_BASE_URL link.
func (h *OrgHandler) InviteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	tenantID := middleware.GetTenantID(ctx)

	if userID == uuid.Nil || tenantID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized or missing tenant context"})
		return
	}

	// Verify requester role in tenant (Must be owner or admin)
	m, err := h.client.Membership.Query().
		Where(membership.UserID(userID), membership.TenantID(tenantID)).
		Only(ctx)
	if err != nil || (m.Role != membership.RoleOwner && m.Role != membership.RoleAdmin) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "only tenant owners and admins can invite members"})
		return
	}

	var req InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "valid email is required"})
		return
	}

	role := invite.RoleMember
	if req.Role == "admin" {
		role = invite.RoleAdmin
	} else if req.Role == "owner" {
		role = invite.RoleOwner
	}

	// Check if target user is already a member of this tenant
	existingUser, err := h.client.User.Query().Where(user.Email(req.Email)).Only(ctx)
	if err == nil && existingUser != nil {
		alreadyMember, _ := h.client.Membership.Query().
			Where(membership.UserID(existingUser.ID), membership.TenantID(tenantID)).
			Exist(ctx)
		if alreadyMember {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "user is already a member of this organization"})
			return
		}
	}

	// Invalidate/delete previous pending invites for the same tenant + email
	deletedCount, err := h.client.Invite.Delete().
		Where(invite.TenantID(tenantID), invite.Email(req.Email), invite.StatusEQ(invite.StatusPending)).
		Exec(ctx)
	if err != nil {
		log.Printf("[ORG INVITE] Warning: failed to invalidate previous pending invites for %s: %v\n", req.Email, err)
	} else if deletedCount > 0 {
		log.Printf("[ORG INVITE] Invalidated %d previous pending invite(s) for email: %s (Tenant: %s)\n", deletedCount, req.Email, tenantID)
	}

	// Generate secure 32-byte hex token with strict 24-hour expiration
	token := generateSecureToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	// Fetch tenant details for email presentation
	t, err := h.client.Tenant.Get(ctx, tenantID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch tenant info"})
		return
	}

	// Create invite record
	inv, err := h.client.Invite.Create().
		SetTenantID(tenantID).
		SetEmail(req.Email).
		SetRole(role).
		SetToken(token).
		SetStatus(invite.StatusPending).
		SetExpiresAt(expiresAt).
		Save(ctx)

	if err != nil {
		log.Printf("[ORG INVITE] Failed to create invite record: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create invitation"})
		return
	}

	// Build dynamic environment-driven invite URL (APP_BASE_URL/invite/<token>)
	inviteURL := resolveBaseURL() + "/invite/" + token
	shortToken := token
	if len(token) >= 8 {
		shortToken = token[:8]
	}
	log.Printf("[ORG INVITE] Generated fresh invite link: %s for email: %s (Tenant: %s, Ref: %s)\n", inviteURL, inv.Email, t.Name, shortToken)

	// Dispatch async email notification
	if h.emailSvc != nil {
		go func(recipientEmail, orgName, roleStr, link, sToken string) {
			if err := h.emailSvc.SendInviteEmail(recipientEmail, orgName, roleStr, link, sToken); err != nil {
				log.Printf("[ORG INVITE] Failed to send invite email to %s: %v\n", recipientEmail, err)
			}
		}(inv.Email, t.Name, string(role), inviteURL, shortToken)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(InviteResponseDTO{
		ID:        inv.ID,
		TenantID:  inv.TenantID,
		Email:     inv.Email,
		Role:      inv.Role,
		Status:    inv.Status,
		ExpiresAt: inv.ExpiresAt,
		CreatedAt: inv.CreatedAt,
	})
}

// GetInviteDetails handles GET /api/orgs/invite/{token}
// Returns non-sensitive invitation details (email, tenant_name, role) for frontend landing page rendering.
func (h *OrgHandler) GetInviteDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	token := chi.URLParam(r, "token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "token is required"})
		return
	}

	inv, err := h.client.Invite.Query().
		Where(invite.Token(token)).
		WithTenant().
		Only(ctx)
	if err != nil || inv == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "invitation not found or invalid"})
		return
	}

	if inv.Status != invite.StatusPending {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invitation has already been processed or cancelled"})
		return
	}

	if time.Now().After(inv.ExpiresAt) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invitation token has expired"})
		return
	}

	tenantName := "Organization"
	if inv.Edges.Tenant != nil {
		tenantName = inv.Edges.Tenant.Name
	}

	json.NewEncoder(w).Encode(InviteDetailsDTO{
		ID:         inv.ID,
		Email:      inv.Email,
		TenantID:   inv.TenantID,
		TenantName: tenantName,
		Role:       string(inv.Role),
		Status:     string(inv.Status),
		ExpiresAt:  inv.ExpiresAt,
	})
}

// AcceptInvite handles POST /api/orgs/accept and POST /api/orgs/accept-invite
// Public endpoint supporting both authenticated users and unauthenticated new user creation with onboarding password.
func (h *OrgHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	var req AcceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invitation token is required"})
		return
	}

	// Retrieve invite by token
	inv, err := h.client.Invite.Query().
		Where(invite.Token(req.Token)).
		WithTenant().
		Only(ctx)
	if err != nil || inv == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid or non-existent invitation token"})
		return
	}

	if inv.Status != invite.StatusPending {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invitation has already been processed or cancelled"})
		return
	}

	if time.Now().After(inv.ExpiresAt) {
		// Mark invite as expired
		_, _ = h.client.Invite.UpdateOne(inv).SetStatus(invite.StatusExpired).Save(ctx)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invitation token has expired"})
		return
	}

	// Dual mode user identification:
	// Check if Authorization header is provided for existing logged-in user flow
	var userID uuid.UUID
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return middleware.JWTSecret, nil
		})
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized: invalid or expired authorization token"})
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userIDStr, _ := claims["user_id"].(string); userIDStr != "" {
				if parsedID, uErr := uuid.Parse(userIDStr); uErr == nil {
					userID = parsedID
				}
			}
		}
		if userID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized: invalid user ID in token"})
			return
		}
	} else {
		// Fallback check from middleware context if present
		userID = middleware.GetUserID(ctx)
	}

	var targetUserID uuid.UUID

	if userID != uuid.Nil {
		// Scenario A: Existing User Flow (Logged In)
		u, err := h.client.User.Get(ctx, userID)
		if err != nil || u == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid user session"})
			return
		}

		// Strict Email Binding Check
		if strings.ToLower(u.Email) != strings.ToLower(inv.Email) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("This invitation was sent to %s, but you are signed in as %s. Please sign in with the invited email address to accept.", inv.Email, u.Email),
			})
			return
		}
		targetUserID = u.ID
	} else {
		// Scenario B: New User Flow (Unauthenticated)
		if req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "password is required to set up your account"})
			return
		}
		if len(req.Password) < 6 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "password must be at least 6 characters long"})
			return
		}

		// Check if user account with invite email already exists
		existingUser, err := h.client.User.Query().
			Where(user.Email(inv.Email)).
			Only(ctx)
		if err == nil && existingUser != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("An account with email %s already exists. Please log in first to accept this invitation.", inv.Email),
			})
			return
		}
	}

	// Begin Atomic Database Transaction
	tx, err := h.client.Tx(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to initialize transaction"})
		return
	}
	defer tx.Rollback()

	// If new user, create User record inside transaction
	if targetUserID == uuid.Nil {
		newUser, err := tx.User.Create().
			SetEmail(inv.Email).
			SetPasswordHash(req.Password).
			Save(ctx)
		if err != nil {
			log.Printf("[ORG ACCEPT] Failed to create user: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to create user account"})
			return
		}
		targetUserID = newUser.ID
	}

	// Convert invite role to membership role
	mRole := membership.RoleMember
	if inv.Role == invite.RoleAdmin {
		mRole = membership.RoleAdmin
	} else if inv.Role == invite.RoleOwner {
		mRole = membership.RoleOwner
	}

	// Check if membership already exists
	alreadyMember, _ := tx.Membership.Query().
		Where(membership.UserID(targetUserID), membership.TenantID(inv.TenantID)).
		Exist(ctx)

	if !alreadyMember {
		_, err = tx.Membership.Create().
			SetUserID(targetUserID).
			SetTenantID(inv.TenantID).
			SetRole(mRole).
			Save(ctx)
		if err != nil {
			log.Printf("[ORG ACCEPT] Failed to create membership: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to join tenant membership"})
			return
		}
	}

	// Delete used invite record from DB (enforce single-use token)
	if err := tx.Invite.DeleteOne(inv).Exec(ctx); err != nil {
		log.Printf("[ORG ACCEPT] Failed to delete invite record: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to process invitation"})
		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to commit transaction"})
		return
	}

	// Generate JWT authentication token for user & tenant
	tokenStr, err := generateToken(targetUserID, inv.TenantID)
	if err != nil {
		log.Printf("[ORG ACCEPT] Failed to generate token: %v\n", err)
	}

	// Fetch tenant memberships for frontend tenant list update
	memberships, _ := h.client.Membership.Query().
		Where(membership.UserID(targetUserID)).
		WithTenant().
		All(ctx)

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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "successfully accepted invitation",
		"token":     tokenStr,
		"user_id":   targetUserID,
		"tenant_id": inv.TenantID,
		"email":     inv.Email,
		"user": map[string]string{
			"id":    targetUserID.String(),
			"email": inv.Email,
		},
		"tenants": tenantDTOs,
	})
}

type TenantMemberDTO struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ListMembers handles GET /api/orgs/members
// Returns all active members of the current selected tenant.
func (h *OrgHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	if tenantID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized or missing tenant context"})
		return
	}

	memberships, err := h.client.Membership.Query().
		Where(membership.TenantID(tenantID)).
		WithUser().
		All(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to query tenant members"})
		return
	}

	dtos := make([]TenantMemberDTO, 0, len(memberships))
	for _, m := range memberships {
		emailStr := "User"
		if m.Edges.User != nil {
			emailStr = m.Edges.User.Email
		}
		dtos = append(dtos, TenantMemberDTO{
			ID:        m.ID,
			UserID:    m.UserID,
			Email:     emailStr,
			Role:      string(m.Role),
			CreatedAt: m.CreatedAt,
		})
	}

	json.NewEncoder(w).Encode(dtos)
}

// ListInvites handles GET /api/orgs/invites
// Returns all pending invitations for the current selected tenant.
func (h *OrgHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	if tenantID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized or missing tenant context"})
		return
	}

	invites, err := h.client.Invite.Query().
		Where(invite.TenantID(tenantID), invite.StatusEQ(invite.StatusPending)).
		All(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to query pending invitations"})
		return
	}

	dtos := make([]InviteResponseDTO, 0, len(invites))
	for _, inv := range invites {
		dtos = append(dtos, InviteResponseDTO{
			ID:        inv.ID,
			TenantID:  inv.TenantID,
			Email:     inv.Email,
			Role:      inv.Role,
			Status:    inv.Status,
			ExpiresAt: inv.ExpiresAt,
			CreatedAt: inv.CreatedAt,
		})
	}

	json.NewEncoder(w).Encode(dtos)
}

// CancelInvite handles DELETE /api/orgs/invites/{id}
// Cancels/deletes a pending invitation.
func (h *OrgHandler) CancelInvite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	tenantID := middleware.GetTenantID(ctx)
	userID := middleware.GetUserID(ctx)
	if tenantID == uuid.Nil || userID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized or missing tenant context"})
		return
	}

	inviteIDStr := chi.URLParam(r, "id")
	inviteID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid invite ID format"})
		return
	}

	// Verify requester is owner or admin
	m, err := h.client.Membership.Query().
		Where(membership.UserID(userID), membership.TenantID(tenantID)).
		Only(ctx)
	if err != nil || (m.Role != membership.RoleOwner && m.Role != membership.RoleAdmin) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "only tenant owners and admins can cancel invites"})
		return
	}

	// Delete invite scoped to current tenant
	deletedCount, err := h.client.Invite.Delete().
		Where(invite.IDEQ(inviteID), invite.TenantID(tenantID)).
		Exec(ctx)
	if err != nil || deletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "invitation not found or already deleted"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "invitation cancelled successfully"})
}
