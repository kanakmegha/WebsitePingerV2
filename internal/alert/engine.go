package alert

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/kanakmegha/WebsitePingerV2/internal/email"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/alert"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/alertevent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/membership"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/notificationchannel"
	"github.com/redis/go-redis/v9"
)

type Engine struct {
	client   *ent.Client
	rdb      *redis.Client
	emailSvc *email.Service
}

func NewEngine(client *ent.Client, rdb *redis.Client, emailSvc *email.Service) *Engine {
	return &Engine{client: client, rdb: rdb, emailSvc: emailSvc}
}

// EvaluateCheck evaluates a completed monitor check and fires or resolves alerts idempotently.
func (e *Engine) EvaluateCheck(ctx context.Context, monitorID uuid.UUID, tenantID uuid.UUID, checkType string, checkStatus string, errStr string) error {
	// Query configured alerts for this monitor
	alerts, err := e.client.Alert.Query().
		Where(alert.MonitorID(monitorID)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query alerts: %w", err)
	}

	m, err := e.client.Monitor.Get(ctx, monitorID)
	if err != nil || m == nil {
		return fmt.Errorf("failed to load monitor: %w", err)
	}

	for _, a := range alerts {
		// State Machine Transition Rules:
		// UP -> DOWN: send firing alert email
		// DOWN -> DOWN: suppressed by Redis state lock (no duplicate email spam)
		// DOWN -> UP: send recovery alert email & release state lock
		if string(a.Type) == checkType || (a.Type == alert.TypeDown && checkType == "http") {
			if checkStatus == "failure" {
				e.handleDownAlert(ctx, a, m, tenantID, checkType, errStr)
			} else if checkStatus == "success" {
				e.handleRecoveryAlert(ctx, a, m, tenantID, checkType)
			}
		}
	}

	return nil
}

func (e *Engine) handleDownAlert(ctx context.Context, a *ent.Alert, m *ent.Monitor, tenantID uuid.UUID, checkType string, errStr string) {
	lockKey := fmt.Sprintf("alert:triggered:%s:%s", m.ID, a.Type)

	// Attempt atomic set to prevent duplicate alert triggering across workers (DOWN -> DOWN suppression)
	set, _ := e.rdb.SetNX(ctx, lockKey, "triggered", 24*time.Hour).Result()
	if !set {
		// Alert already active in DOWN state; suppress duplicate email spam
		return
	}

	// Create AlertEvent in DB
	msg := fmt.Sprintf("Monitor [%s] check failed: %s", checkType, errStr)
	event, err := e.client.AlertEvent.Create().
		SetAlertID(a.ID).
		SetMonitorID(m.ID).
		SetStatus(alertevent.StatusTriggered).
		SetMessage(msg).
		Save(ctx)

	if err != nil {
		log.Printf("[ALERT ENGINE] Failed to record alert event: %v", err)
		return
	}

	// Dispatch notifications asynchronously via worker goroutine
	go e.dispatchNotifications(context.Background(), tenantID, m.Name, checkType, "firing", event.Message)
}

func (e *Engine) handleRecoveryAlert(ctx context.Context, a *ent.Alert, m *ent.Monitor, tenantID uuid.UUID, checkType string) {
	lockKey := fmt.Sprintf("alert:triggered:%s:%s", m.ID, a.Type)

	// Check if active alert lock exists in Redis
	val, _ := e.rdb.Get(ctx, lockKey).Result()
	if val != "triggered" {
		return // No active alert to resolve (UP -> UP state)
	}

	// Clear lock in Redis (DOWN -> UP transition)
	e.rdb.Del(ctx, lockKey)

	// Create Resolution AlertEvent in DB
	msg := fmt.Sprintf("Monitor [%s] check recovered successfully.", checkType)
	event, err := e.client.AlertEvent.Create().
		SetAlertID(a.ID).
		SetMonitorID(m.ID).
		SetStatus(alertevent.StatusResolved).
		SetMessage(msg).
		Save(ctx)

	if err != nil {
		log.Printf("[ALERT ENGINE] Failed to record resolution event: %v", err)
		return
	}

	// Dispatch recovery notification asynchronously
	go e.dispatchNotifications(context.Background(), tenantID, m.Name, checkType, "resolved", event.Message)
}

func (e *Engine) dispatchNotifications(ctx context.Context, tenantID uuid.UUID, monitorName string, checkType string, status string, message string) {
	// Query configured notification channels for tenant
	channels, err := e.client.NotificationChannel.Query().
		Where(notificationchannel.TenantID(tenantID)).
		All(ctx)
	if err != nil {
		log.Printf("[ALERT DISPATCH] Failed to fetch notification channels for tenant %s: %v", tenantID, err)
		return
	}

	// Fetch tenant owner/admin email recipients as fallback/primary targets
	members, err := e.client.Membership.Query().
		Where(membership.TenantID(tenantID)).
		WithUser().
		All(ctx)
	if err != nil {
		log.Printf("[ALERT DISPATCH] Failed to fetch tenant members: %v", err)
	}

	recipientEmails := make(map[string]bool)
	for _, mem := range members {
		if mem.Edges.User != nil && mem.Edges.User.Email != "" {
			recipientEmails[mem.Edges.User.Email] = true
		}
	}

	for _, ch := range channels {
		log.Printf("[ALERT DISPATCH] Tenant %s | Channel %s (%s) => %s", tenantID, ch.ID, ch.Type, message)

		if ch.Type == notificationchannel.TypeEmail {
			if cfgEmail, ok := ch.Config["email"].(string); ok && cfgEmail != "" {
				recipientEmails[cfgEmail] = true
			}
		}
	}

	if e.emailSvc != nil {
		for recipient := range recipientEmails {
			if err := e.emailSvc.SendAlertEmail(recipient, monitorName, checkType, status, message); err != nil {
				log.Printf("[ALERT DISPATCH ERROR] Failed sending alert email to %s: %v", recipient, err)
			}
		}
	}
}
