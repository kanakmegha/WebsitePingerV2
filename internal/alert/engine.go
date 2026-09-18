package alert

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/alert"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/alertevent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/notificationchannel"
	"github.com/redis/go-redis/v9"
)

type Engine struct {
	client *ent.Client
	rdb    *redis.Client
}

func NewEngine(client *ent.Client, rdb *redis.Client) *Engine {
	return &Engine{client: client, rdb: rdb}
}

// EvaluateCheck evaluates a completed monitor check and fires or resolves alerts idempotently.
func (e *Engine) EvaluateCheck(ctx context.Context, monitorID uuid.UUID, tenantID uuid.UUID, checkStatus string, errStr string) error {
	// Query configured alerts for this monitor
	alerts, err := e.client.Alert.Query().
		Where(alert.MonitorID(monitorID)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query alerts: %w", err)
	}

	for _, a := range alerts {
		if a.Type == alert.TypeDown {
			if checkStatus == "failure" {
				e.handleDownAlert(ctx, a, monitorID, tenantID, errStr)
			} else if checkStatus == "success" {
				e.handleRecoveryAlert(ctx, a, monitorID, tenantID)
			}
		}
	}

	return nil
}

func (e *Engine) handleDownAlert(ctx context.Context, a *ent.Alert, monitorID uuid.UUID, tenantID uuid.UUID, errStr string) {
	lockKey := fmt.Sprintf("alert:triggered:%s:%s", monitorID, a.Type)

	// Attempt atomic set to prevent duplicate alert triggering across workers
	set, _ := e.rdb.SetNX(ctx, lockKey, "triggered", 24*time.Hour).Result()
	if !set {
		// Alert already triggered and active; suppress duplicate
		return
	}

	// Create AlertEvent in DB
	event, err := e.client.AlertEvent.Create().
		SetAlertID(a.ID).
		SetMonitorID(monitorID).
		SetStatus(alertevent.StatusTriggered).
		SetMessage(fmt.Sprintf("Monitor failed inspection: %s", errStr)).
		Save(ctx)

	if err != nil {
		log.Printf("Failed to record alert event: %v", err)
		return
	}

	// Dispatch notification channels
	e.dispatchNotifications(ctx, tenantID, event.Message)
}

func (e *Engine) handleRecoveryAlert(ctx context.Context, a *ent.Alert, monitorID uuid.UUID, tenantID uuid.UUID) {
	lockKey := fmt.Sprintf("alert:triggered:%s:%s", monitorID, a.Type)

	// Check if active alert lock exists in Redis
	val, _ := e.rdb.Get(ctx, lockKey).Result()
	if val != "triggered" {
		return // No active alert to resolve
	}

	// Clear lock in Redis
	e.rdb.Del(ctx, lockKey)

	// Create Resolution AlertEvent in DB
	event, err := e.client.AlertEvent.Create().
		SetAlertID(a.ID).
		SetMonitorID(monitorID).
		SetStatus(alertevent.StatusResolved).
		SetMessage("Monitor check recovered successfully.").
		Save(ctx)

	if err != nil {
		log.Printf("Failed to record resolution event: %v", err)
		return
	}

	e.dispatchNotifications(ctx, tenantID, event.Message)
}

func (e *Engine) dispatchNotifications(ctx context.Context, tenantID uuid.UUID, message string) {
	channels, err := e.client.NotificationChannel.Query().
		Where(notificationchannel.TenantID(tenantID)).
		All(ctx)
	if err != nil {
		log.Printf("Failed to fetch notification channels: %v", err)
		return
	}

	for _, ch := range channels {
		log.Printf("[ALERT DISPATCH] Tenant %s | Channel %s (%s) => %s", tenantID, ch.ID, ch.Type, message)
		// Integrate Webhook, Email, Slack dispatchers here
	}
}
