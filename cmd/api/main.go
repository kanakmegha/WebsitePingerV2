package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/membership"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/user"
	"github.com/kanakmegha/WebsitePingerV2/internal/handler"
	customMiddleware "github.com/kanakmegha/WebsitePingerV2/internal/middleware"

	_ "github.com/lib/pq"
)

func seedDefaultUser(ctx context.Context, client *ent.Client) {
	email := "admin@pinger.com"
	password := "admin123"

	exists, err := client.User.Query().Where(user.Email(email)).Exist(ctx)
	if err != nil {
		log.Printf("[SEED] Error checking default user existence: %v", err)
		return
	}
	if exists {
		log.Println("[SEED] Default user already exists, skipping seed.")
		return
	}

	// Transactional seeding
	tx, err := client.Tx(ctx)
	if err != nil {
		log.Printf("[SEED] Failed starting seed transaction: %v", err)
		return
	}
	defer tx.Rollback()

	t, err := tx.Tenant.Create().
		SetName("Default Organization").
		SetPlan("free").
		Save(ctx)
	if err != nil {
		log.Printf("[SEED] Failed to seed tenant: %v", err)
		return
	}

	u, err := tx.User.Create().
		SetEmail(email).
		SetPasswordHash(password).
		Save(ctx)
	if err != nil {
		log.Printf("[SEED] Failed to seed default user: %v", err)
		return
	}

	tx.TenantSetting.Create().
		SetTenantID(t.ID).
		SetHTTPIntervalSeconds(60).
		SetDNSIntervalSeconds(300).
		SetSslIntervalSeconds(3600).
		SetDomainIntervalSeconds(86400).
		SetEmailAuthIntervalSeconds(300).
		Save(ctx)

	tx.Membership.Create().
		SetUserID(u.ID).
		SetTenantID(t.ID).
		SetRole(membership.RoleOwner).
		Save(ctx)

	if err := tx.Commit(); err != nil {
		log.Printf("[SEED] Failed to commit seed transaction: %v", err)
		return
	}

	log.Printf("[SEED] Default user successfully created (Email: %s, Password: %s, TenantID: %s)", email, password, t.ID)
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@postgres:5432/pinger?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Connect PostgreSQL via Ent ORM with retry logic
	var client *ent.Client
	var err error
	for i := 1; i <= 10; i++ {
		client, err = ent.Open("postgres", dbURL)
		if err == nil {
			// Test ping connection
			if err = client.Schema.Create(context.Background()); err == nil {
				log.Println("[API] Connected to database and applied schema migrations successfully.")
				break
			}
		}
		log.Printf("[API] Waiting for database connection (attempt %d/10)... error: %v", i, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("[API] Could not connect to database after retries: %v", err)
	}
	defer client.Close()

	// Seed default user and tenant
	seedDefaultUser(context.Background(), client)

	r := chi.NewRouter()

	// Standard middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS Headers
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Healthcheck endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	authH := handler.NewAuthHandler(client)
	monitorH := handler.NewMonitorHandler(client)
	settingsH := handler.NewSettingsHandler(client)

	// Public Auth Routes
	r.Post("/api/auth/register", authH.Register)
	r.Post("/api/auth/login", authH.Login)

	// Protected Multi-Tenant API Routes
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.RequireAuth(client))

		r.Get("/api/me/tenants", authH.GetUserTenants)

		r.Get("/api/monitors", monitorH.List)
		r.Post("/api/monitors", monitorH.Create)
		r.Get("/api/monitors/{id}", monitorH.Get)
		r.Get("/api/monitors/{id}/checks", monitorH.GetChecks)
		r.Delete("/api/monitors/{id}", monitorH.Delete)

		r.Get("/api/settings", settingsH.GetSettings)
		r.Put("/api/settings", settingsH.UpdateSettings)
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("[API] Server listening on port :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	log.Println("[API] Shutting down API server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
