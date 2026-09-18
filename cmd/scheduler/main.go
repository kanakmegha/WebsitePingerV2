package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/monitorcheckconfig"
	"github.com/kanakmegha/WebsitePingerV2/internal/queue"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@postgres:5432/pinger?sslmode=disable"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	// Connect PostgreSQL via Ent with retry logic
	var client *ent.Client
	var err error
	for i := 1; i <= 10; i++ {
		client, err = ent.Open("postgres", dbURL)
		if err == nil {
			if pingErr := client.Schema.Create(context.Background()); pingErr == nil {
				log.Println("[SCHEDULER] Connected to postgres database successfully.")
				break
			}
		}
		log.Printf("[SCHEDULER] Waiting for postgres connection (attempt %d/10)... error: %v", i, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("[SCHEDULER] Could not connect to postgres after retries: %v", err)
	}
	defer client.Close()

	// Connect Redis with retry logic
	var rdb *redis.Client
	for i := 1; i <= 10; i++ {
		rdb = redis.NewClient(&redis.Options{Addr: redisAddr})
		if _, pingErr := rdb.Ping(context.Background()).Result(); pingErr == nil {
			log.Println("[SCHEDULER] Connected to redis successfully.")
			break
		}
		log.Printf("[SCHEDULER] Waiting for redis connection (attempt %d/10)...", i)
		time.Sleep(2 * time.Second)
	}

	q := queue.NewRedisQueue(rdb)

	log.Println("[SCHEDULER] Started Per-Check Independent Scheduler Engine...")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			now := time.Now()

			// Query active per-check configurations that are due (next_check_at <= now)
			configs, err := client.MonitorCheckConfig.Query().
				Where(
					monitorcheckconfig.IsEnabled(true),
					monitorcheckconfig.NextCheckAtLTE(now),
				).
				WithMonitor().
				All(ctx)

			if err != nil {
				log.Printf("[SCHEDULER] Error querying check configs: %v", err)
				continue
			}

			enqueuedCount := 0

			for _, cfg := range configs {
				m := cfg.Edges.Monitor
				if m == nil || !m.IsActive {
					continue
				}

				job := queue.CheckJob{
					MonitorID:  m.ID,
					TenantID:   m.TenantID,
					CheckType:  string(cfg.CheckType),
					Domain:     m.Domain,
					URL:        m.URL,
					TimeoutSec: cfg.TimeoutSeconds,
					Timestamp:  now,
				}

				if err := q.Enqueue(ctx, job); err == nil {
					enqueuedCount++
					nextCheck := now.Add(time.Duration(cfg.IntervalSeconds) * time.Second)
					// Advance next_check_at timestamp for this check config and set last_checked_at
					client.MonitorCheckConfig.UpdateOneID(cfg.ID).
						SetLastCheckedAt(now).
						SetNextCheckAt(nextCheck).
						Exec(ctx)
				}
			}

			if enqueuedCount > 0 {
				log.Printf("[SCHEDULER] Successfully enqueued %d due check jobs to Redis", enqueuedCount)
			}

		case <-stopChan:
			log.Println("[SCHEDULER] Shutting down scheduler gracefully...")
			return
		}
	}
}
