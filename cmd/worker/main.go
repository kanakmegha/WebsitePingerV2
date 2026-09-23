package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kanakmegha/WebsitePingerV2/internal/alert"
	"github.com/kanakmegha/WebsitePingerV2/internal/checker"
	"github.com/kanakmegha/WebsitePingerV2/internal/coalescer"
	"github.com/kanakmegha/WebsitePingerV2/internal/email"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/monitorcheck"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/monitorcheckconfig"
	"github.com/kanakmegha/WebsitePingerV2/internal/queue"
	"github.com/kanakmegha/WebsitePingerV2/internal/ratelimit"

	"github.com/kanakmegha/WebsitePingerV2/internal/push"

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

	// Connect Ent / Postgres with retry logic
	var client *ent.Client
	var err error
	for i := 1; i <= 10; i++ {
		client, err = ent.Open("postgres", dbURL)
		if err == nil {
			if pingErr := client.Schema.Create(context.Background()); pingErr == nil {
				log.Println("[WORKER] Connected to postgres database successfully.")
				break
			}
		}
		log.Printf("[WORKER] Waiting for postgres connection (attempt %d/10)... error: %v", i, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("[WORKER] Could not connect to postgres after retries: %v", err)
	}
	defer client.Close()

	// Connect Redis with retry logic
	var rdb *redis.Client
	for i := 1; i <= 10; i++ {
		rdb = redis.NewClient(&redis.Options{Addr: redisAddr})
		if _, pingErr := rdb.Ping(context.Background()).Result(); pingErr == nil {
			log.Println("[WORKER] Connected to redis successfully.")
			break
		}
		log.Printf("[WORKER] Waiting for redis connection (attempt %d/10)...", i)
		time.Sleep(2 * time.Second)
	}

	q := queue.NewRedisQueue(rdb)

	limiter := ratelimit.NewRateLimiter(rdb)
	coalesceEngine := coalescer.NewCoalescer(rdb)

	emailCfg := email.LoadConfigFromEnv()
	emailSvc := email.NewService(emailCfg)
	pushSvc := push.NewPushService(client)
	alertEng := alert.NewEngine(client, rdb, emailSvc, pushSvc)

	const workerCount = 50
	var wg sync.WaitGroup

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	log.Printf("[WORKER] Starting Worker Pool with %d goroutines...", workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				job, err := q.Dequeue(ctx, 2*time.Second)
				if err != nil {
					cancel()
					continue
				}
				if job == nil {
					cancel()
					continue
				}

				processJob(ctx, client, limiter, coalesceEngine, alertEng, job)
				cancel()
			}
		}(i)
	}

	<-stopChan
	log.Println("[WORKER] Gracefully stopping worker pool...")
	wg.Wait()
}

func processJob(ctx context.Context, client *ent.Client, limiter *ratelimit.RateLimiter, c *coalescer.Coalescer, alertEng *alert.Engine, job *queue.CheckJob) {
	checkType := ratelimit.CheckType(job.CheckType)

	// Worker ALWAYS executes dequeued job without skipping
	result, err := c.Do(ctx, job.CheckType, job.Domain, func(ctx context.Context) (interface{}, error) {
		switch checkType {
		case ratelimit.CheckTypeHTTP:
			return checker.CheckHTTP(ctx, job.URL, job.TimeoutSec)
		case ratelimit.CheckTypeSSL:
			return checker.CheckSSL(ctx, job.Domain, job.TimeoutSec)
		case ratelimit.CheckTypeWHOIS:
			return checker.CheckWHOIS(ctx, job.Domain)
		case ratelimit.CheckTypeDNS:
			return checker.CheckDNS(ctx, job.Domain, job.TimeoutSec)
		default:
			return checker.CheckHTTP(ctx, job.URL, job.TimeoutSec)
		}
	})

	if err != nil {
		log.Printf("[WORKER] Execution error for %s (%s): %v", job.Domain, job.CheckType, err)
		return
	}

	log.Printf("[WORKER] Executed %s check for %s", job.CheckType, job.Domain)

	// Save results into PostgreSQL using Ent ORM
	saveCheckResults(ctx, client, alertEng, job, result)

	// Update next_check_at = NOW() + interval_seconds in monitor_check_configs
	now := time.Now()
	cfg, cfgErr := client.MonitorCheckConfig.Query().
		Where(
			monitorcheckconfig.MonitorID(job.MonitorID),
			monitorcheckconfig.CheckTypeEQ(monitorcheckconfig.CheckType(job.CheckType)),
		).
		Only(ctx)
	if cfgErr == nil && cfg != nil {
		nextCheck := now.Add(time.Duration(cfg.IntervalSeconds) * time.Second)
		client.MonitorCheckConfig.UpdateOneID(cfg.ID).
			SetLastCheckedAt(now).
			SetNextCheckAt(nextCheck).
			Exec(ctx)
	}
}

func saveCheckResults(ctx context.Context, client *ent.Client, alertEng *alert.Engine, job *queue.CheckJob, res interface{}) {
	status := monitorcheck.StatusSuccess
	var errMessage string
	var responseTimeMs int

	mcBuilder := client.MonitorCheck.Create().
		SetMonitorID(job.MonitorID).
		SetCheckType(monitorcheck.CheckType(job.CheckType))

	switch v := res.(type) {
	case *checker.HTTPResult:
		responseTimeMs = int(v.ResponseTimeMS)
		if !v.Success {
			status = monitorcheck.StatusFailure
			errMessage = v.Error
		}
		mcStr, err := mcBuilder.SetStatus(status).SetError(errMessage).Save(ctx)
		if err == nil {
			client.HTTPCheckResult.Create().
				SetCheckID(mcStr.ID).
				SetStatusCode(v.StatusCode).
				SetResponseTimeMs(v.ResponseTimeMS).
				SetResponseSizeBytes(v.ResponseSizeBytes).
				SetFinalURL(v.FinalURL).
				Save(ctx)
		}

	case *checker.SSLResult:
		if !v.Valid {
			status = monitorcheck.StatusFailure
			errMessage = v.Error
		}
		mcStr, err := mcBuilder.SetStatus(status).SetError(errMessage).Save(ctx)
		if err == nil {
			client.SSLCheckResult.Create().
				SetCheckID(mcStr.ID).
				SetExpiryDate(v.ExpiryDate).
				SetIssuer(v.Issuer).
				SetValid(v.Valid).
				SetDaysRemaining(v.DaysRemaining).
				Save(ctx)
		}

	case *checker.DomainResult:
		if v.Error != "" {
			status = monitorcheck.StatusFailure
			errMessage = v.Error
		}
		mcStr, err := mcBuilder.SetStatus(status).SetError(v.Error).Save(ctx)
		if err == nil {
			client.DomainCheckResult.Create().
				SetCheckID(mcStr.ID).
				SetExpiryDate(v.ExpiryDate).
				SetRegistrar(v.Registrar).
				SetDaysRemaining(v.DaysRemaining).
				Save(ctx)
		}

	case *checker.DNSResult:
		if v.Error != "" || !v.HasARecord {
			status = monitorcheck.StatusFailure
			if v.Error != "" {
				errMessage = v.Error
			} else {
				errMessage = "DNS check failed: No A record found"
			}
		}
		mcStr, err := mcBuilder.SetStatus(status).SetError(errMessage).Save(ctx)
		if err == nil {
			client.DNSCheckResult.Create().
				SetCheckID(mcStr.ID).
				SetHasARecord(v.HasARecord).
				SetHasAaaaRecord(v.HasAAAARecord).
				SetARecords(v.ARecords).
				SetAaaaRecords(v.AAAARecords).
				SetMxRecords(v.MXRecords).
				SetTxtRecords(v.TXTRecords).
				SetSpfValid(v.SPFValid).
				SetDmarcValid(v.DMARCValid).
				Save(ctx)
		}
	}

	// Step 4: Evaluate State-Driven Alerting Rules
	statusStr := "success"
	if status == monitorcheck.StatusFailure {
		statusStr = "failure"
	}
	alertEng.EvaluateCheck(ctx, job.MonitorID, job.TenantID, job.CheckType, statusStr, errMessage, responseTimeMs)
}
