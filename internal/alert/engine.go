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
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/monitor"
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

// EvaluateCheck evaluates a completed monitor check using a state-driven transition matrix:
// CASE 1: New Monitor (no previous state) -> If DOWN, send INCIDENT email & set last_status="down".
// CASE 2: UP -> DOWN -> Send INCIDENT email immediately & set last_status="down".
// CASE 3: DOWN -> DOWN -> DO NOT send email spam. Update last_checked_at timestamp.
// CASE 4: DOWN -> UP -> Send RECOVERY email & set last_status="up".
func (e *Engine) EvaluateCheck(ctx context.Context, monitorID uuid.UUID, tenantID uuid.UUID, checkType string, checkStatus string, errStr string, responseTimeMs int) error {
	// Only HTTP checks (primary site availability checks) drive overall monitor UP/DOWN state transitions
	if checkType != "http" {
		return nil
	}

	// Acquire Redis atomic lock per monitor to prevent concurrent worker race conditions
	lockKey := fmt.Sprintf("alert:transition:lock:%s", monitorID)
	acquired, _ := e.rdb.SetNX(ctx, lockKey, "processing", 10*time.Second).Result()
	if !acquired {
		// Another worker thread is currently evaluating state transition for this monitor
		log.Printf("[SKIP] Concurrent alert evaluation in progress for monitor %s", monitorID)
		return nil
	}
	defer e.rdb.Del(ctx, lockKey)

	m, err := e.client.Monitor.Get(ctx, monitorID)
	if err != nil || m == nil {
		return fmt.Errorf("failed to load monitor: %w", err)
	}

	now := time.Now()
	currentIsDown := checkStatus == "failure"
	currentStatusEnum := monitor.LastStatusUp
	if currentIsDown {
		currentStatusEnum = monitor.LastStatusDown
	}

	var prevStatusStr string
	if m.LastStatus != nil {
		prevStatusStr = string(*m.LastStatus)
	}

	// Fetch tenant details for email presentation
	tenantName := "Organization"
	t, err := e.client.Tenant.Get(ctx, tenantID)
	if err == nil && t != nil {
		tenantName = t.Name
	}

	// State-Driven Transition Logic
	if prevStatusStr == "" {
		// CASE 1: New Monitor (First check execution)
		if currentIsDown {
			log.Printf("[ALERT] DOWN detected for %s", m.Domain)
			e.client.Monitor.UpdateOneID(m.ID).
				SetLastStatus(currentStatusEnum).
				SetLastCheckedAt(now).
				SetLastAlertSentAt(now).
				Exec(ctx)
			e.recordAlertEvent(ctx, m, alertevent.StatusTriggered, fmt.Sprintf("Initial check failed: %s", errStr))
			go e.dispatchIncidentEmail(context.Background(), tenantID, tenantName, m.Domain, errStr, now, responseTimeMs)
		} else {
			log.Printf("[SKIP] No state change for %s", m.Domain)
			e.client.Monitor.UpdateOneID(m.ID).
				SetLastStatus(currentStatusEnum).
				SetLastCheckedAt(now).
				Exec(ctx)
		}
	} else if prevStatusStr == "up" && currentIsDown {
		// CASE 2: UP -> DOWN (New Incident Transition - Send Alert Email Immediately)
		log.Printf("[ALERT] DOWN detected for %s", m.Domain)
		e.client.Monitor.UpdateOneID(m.ID).
			SetLastStatus(monitor.LastStatusDown).
			SetLastCheckedAt(now).
			SetLastAlertSentAt(now).
			Exec(ctx)

		e.recordAlertEvent(ctx, m, alertevent.StatusTriggered, fmt.Sprintf("Incident detected: %s", errStr))
		go e.dispatchIncidentEmail(context.Background(), tenantID, tenantName, m.Domain, errStr, now, responseTimeMs)

	} else if prevStatusStr == "down" && currentIsDown {
		// CASE 3: DOWN -> DOWN (Ongoing Incident - Suppress duplicate email)
		log.Printf("[SKIP] No state change for %s", m.Domain)
		e.client.Monitor.UpdateOneID(m.ID).
			SetLastCheckedAt(now).
			Exec(ctx)

	} else if prevStatusStr == "down" && !currentIsDown {
		// CASE 4: DOWN -> UP (Recovery Transition)
		log.Printf("[ALERT] RECOVERY detected for %s", m.Domain)

		downtimeDuration := "N/A"
		if m.LastAlertSentAt != nil {
			downtimeDuration = now.Sub(*m.LastAlertSentAt).Round(time.Second).String()
		}

		e.client.Monitor.UpdateOneID(m.ID).
			SetLastStatus(monitor.LastStatusUp).
			SetLastCheckedAt(now).
			Exec(ctx)

		e.recordAlertEvent(ctx, m, alertevent.StatusResolved, fmt.Sprintf("Incident resolved for %s", m.Domain))
		go e.dispatchRecoveryEmail(context.Background(), tenantID, tenantName, m.Domain, downtimeDuration, now)
	} else {
		// UP -> UP: Normal operating state
		log.Printf("[SKIP] No state change for %s", m.Domain)
		e.client.Monitor.UpdateOneID(m.ID).
			SetLastCheckedAt(now).
			Exec(ctx)
	}

	return nil
}

func (e *Engine) dispatchIncidentEmail(ctx context.Context, tenantID uuid.UUID, tenantName string, domain string, reason string, timestamp time.Time, responseTimeMs int) {
	recipients := e.getTenantRecipientEmails(ctx, tenantID)
	if e.emailSvc != nil {
		for recipient := range recipients {
			if err := e.emailSvc.SendIncidentAlertEmail(recipient, domain, reason, timestamp, responseTimeMs, tenantName); err != nil {
				log.Printf("[ALERT DISPATCH ERROR] Failed sending incident alert email to %s: %v", recipient, err)
			}
		}
	}
}

func (e *Engine) dispatchRecoveryEmail(ctx context.Context, tenantID uuid.UUID, tenantName string, domain string, downtimeDuration string, recoveredAt time.Time) {
	recipients := e.getTenantRecipientEmails(ctx, tenantID)
	if e.emailSvc != nil {
		for recipient := range recipients {
			if err := e.emailSvc.SendRecoveryAlertEmail(recipient, domain, downtimeDuration, recoveredAt, tenantName); err != nil {
				log.Printf("[ALERT DISPATCH ERROR] Failed sending recovery alert email to %s: %v", recipient, err)
			}
		}
	}
}

func (e *Engine) getTenantRecipientEmails(ctx context.Context, tenantID uuid.UUID) map[string]bool {
	recipients := make(map[string]bool)

	// Fetch all tenant members (owner, admin, member)
	members, err := e.client.Membership.Query().
		Where(membership.TenantID(tenantID)).
		WithUser().
		All(ctx)
	if err == nil {
		for _, mem := range members {
			if mem.Edges.User != nil && mem.Edges.User.Email != "" {
				recipients[mem.Edges.User.Email] = true
			}
		}
	}

	// Fetch notification channel emails
	channels, err := e.client.NotificationChannel.Query().
		Where(notificationchannel.TenantID(tenantID)).
		All(ctx)
	if err == nil {
		for _, ch := range channels {
			if ch.Type == notificationchannel.TypeEmail {
				if cfgEmail, ok := ch.Config["email"].(string); ok && cfgEmail != "" {
					recipients[cfgEmail] = true
				}
			}
		}
	}

	return recipients
}

func (e *Engine) recordAlertEvent(ctx context.Context, m *ent.Monitor, status alertevent.Status, message string) {
	alerts, _ := e.client.Alert.Query().
		Where(alert.MonitorID(m.ID)).
		All(ctx)

	for _, a := range alerts {
		_, _ = e.client.AlertEvent.Create().
			SetAlertID(a.ID).
			SetMonitorID(m.ID).
			SetStatus(status).
			SetMessage(message).
			Save(ctx)
	}
}
