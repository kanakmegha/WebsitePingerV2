package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/membership"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/tenant"
)

type contextKey string

const (
	UserIDKey   contextKey = "userID"
	TenantIDKey contextKey = "tenantID"
)

var JWTSecret = []byte("super-secret-pinger-key-change-in-production")

// ResolveTenant encapsulates the core tenant resolution and validation logic.
// 1. Header X-Tenant-ID is required for tenant resolution.
// 2. If missing, return an error (HTTP 400 Bad Request) unless ALLOW_TENANT_FALLBACK=true env is explicitly enabled.
func ResolveTenant(ctx context.Context, client *ent.Client, userID uuid.UUID, headerTenantID string) (uuid.UUID, error) {
	if client == nil {
		return uuid.Nil, fmt.Errorf("database client unavailable")
	}

	// Case 1: Header X-Tenant-ID present
	if headerTenantID != "" {
		targetID, err := uuid.Parse(headerTenantID)
		if err == nil && targetID != uuid.Nil {
			hasMembership, mErr := client.Membership.Query().
				Where(
					membership.UserID(userID),
					membership.TenantID(targetID),
				).
				Exist(ctx)

			if mErr == nil && hasMembership {
				return targetID, nil
			}
			return uuid.Nil, fmt.Errorf("user does not have membership access to tenant %s", targetID)
		}
		return uuid.Nil, fmt.Errorf("invalid X-Tenant-ID header format: %s", headerTenantID)
	}

	// Case 2: Controlled Fallback ONLY when ALLOW_TENANT_FALLBACK env variable is explicitly set to "true"
	if os.Getenv("ALLOW_TENANT_FALLBACK") == "true" {
		log.Printf("[AUTH MIDDLEWARE] WARNING: Using temporary tenant fallback for user %s (ALLOW_TENANT_FALLBACK=true)", userID)
		firstMem, err := client.Membership.Query().
			Where(membership.UserID(userID)).
			Order(ent.Asc(membership.FieldCreatedAt)).
			First(ctx)

		if err == nil && firstMem != nil {
			return firstMem.TenantID, nil
		}
		return uuid.Nil, fmt.Errorf("no valid tenant membership found for user %s", userID)
	}

	log.Printf("[AUTH MIDDLEWARE] ERROR: Missing X-Tenant-ID header for user %s", userID)
	return uuid.Nil, fmt.Errorf("missing X-Tenant-ID header")
}

// RequireAuth authenticates request JWT (extracting user_id) and resolves tenant_id via ResolveTenant helper.
func RequireAuth(client *ent.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized: missing authorization header"}`))
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return JWTSecret, nil
			})

			if err != nil || !token.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized: invalid or expired token"}`))
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized: invalid token claims"}`))
				return
			}

			userIDStr, _ := claims["user_id"].(string)
			userID, uErr := uuid.Parse(userIDStr)
			if uErr != nil || userID == uuid.Nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized: invalid user ID in token"}`))
				return
			}

			// Priority fallback check for legacy tokens
			headerTenantID := r.Header.Get("X-Tenant-ID")
			if headerTenantID == "" {
				jwtTenantIDStr, _ := claims["tenant_id"].(string)
				if jwtTenantIDStr != "" {
					headerTenantID = jwtTenantIDStr
					log.Printf("[AUTH MIDDLEWARE] Legacy JWT tenant_id fallback used for user %s", userID)
				}
			}

			// Resolve tenant_id via ResolveTenant helper
			tenantID, resErr := ResolveTenant(r.Context(), client, userID, headerTenantID)
			if resErr != nil {
				log.Printf("[AUTH MIDDLEWARE] User %s tenant resolution failed: %v", userID, resErr)
				if strings.Contains(resErr.Error(), "missing X-Tenant-ID header") {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error":"bad_request: missing X-Tenant-ID header"}`))
					return
				}
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(fmt.Sprintf(`{"error":"forbidden: %s"}`, resErr.Error())))
				return
			}

			// Ensure Tenant Record exists in database
			tExists, _ := client.Tenant.Query().Where(tenant.IDEQ(tenantID)).Exist(r.Context())
			if !tExists {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"forbidden: tenant record does not exist in database"}`))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, TenantIDKey, tenantID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthMiddleware remains for backward compatibility
func AuthMiddleware(next http.Handler) http.Handler {
	return RequireAuth(nil)(next)
}

// OptionalAuth parses Authorization header if present to populate UserID in context, but does not block unauthenticated requests.
func OptionalAuth(client *ent.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
					return JWTSecret, nil
				})
				if err == nil && token.Valid {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						if userIDStr, _ := claims["user_id"].(string); userIDStr != "" {
							if userID, uErr := uuid.Parse(userIDStr); uErr == nil && userID != uuid.Nil {
								ctx := context.WithValue(r.Context(), UserIDKey, userID)
								r = r.WithContext(ctx)
							}
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetTenantID(ctx context.Context) uuid.UUID {
	val, _ := ctx.Value(TenantIDKey).(uuid.UUID)
	return val
}

func GetUserID(ctx context.Context) uuid.UUID {
	val, _ := ctx.Value(UserIDKey).(uuid.UUID)
	return val
}
