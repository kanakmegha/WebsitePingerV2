package push

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/alertevent"
	"github.com/kanakmegha/WebsitePingerV2/internal/ent/pushsubscription"
)

type PushService struct {
	client           *ent.Client
	vapidPublicKey   string
	vapidPrivateKey  string
	subscriberSubject string
}

func NewPushService(client *ent.Client) *PushService {
	pubKey := os.Getenv("VAPID_PUBLIC_KEY")
	privKey := os.Getenv("VAPID_PRIVATE_KEY")
	subject := os.Getenv("VAPID_SUBJECT")

	if subject == "" {
		subject = "mailto:admin@pinger.com"
	}

	if pubKey == "" || privKey == "" {
		log.Println("[PUSH] VAPID keys not configured in environment. Generating dynamic VAPID key pair for runtime...")
		privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
		if err != nil {
			log.Printf("[PUSH ERROR] Failed generating VAPID key pair: %v", err)
		} else {
			pubKey = publicKey
			privKey = privateKey
			log.Printf("[PUSH] Dynamic VAPID Key Pair initialized successfully.")
		}
	} else {
		log.Println("[PUSH] Loaded VAPID Key Pair from environment.")
	}

	return &PushService{
		client:            client,
		vapidPublicKey:    pubKey,
		vapidPrivateKey:   privKey,
		subscriberSubject: subject,
	}
}

func (s *PushService) GetPublicKey() string {
	return s.vapidPublicKey
}

type PushNotificationPayload struct {
	Title       string `json:"title"`
	Body        string `json:"body"`
	URL         string `json:"url"`
	UnreadCount int    `json:"unread_count"`
}

func (s *PushService) SendTenantPushNotification(ctx context.Context, tenantID uuid.UUID, title string, body string, targetURL string) {
	if s.vapidPublicKey == "" || s.vapidPrivateKey == "" {
		log.Println("[PUSH SKIP] Push notifications disabled: VAPID keys missing.")
		return
	}

	subs, err := s.client.PushSubscription.Query().
		Where(pushsubscription.TenantID(tenantID)).
		All(ctx)
	if err != nil || len(subs) == 0 {
		return
	}

	unreadCount, _ := s.client.AlertEvent.Query().
		Where(
			alertevent.TenantID(tenantID),
			alertevent.StatusEQ(alertevent.StatusUnread),
		).
		Count(ctx)

	payloadData, err := json.Marshal(PushNotificationPayload{
		Title:       title,
		Body:        body,
		URL:         targetURL,
		UnreadCount: unreadCount,
	})
	if err != nil {
		log.Printf("[PUSH ERROR] Failed marshalling push payload: %v", err)
		return
	}

	log.Printf("[PUSH DISPATCH] Dispatching push notification to %d subscriber(s) for tenant %s", len(subs), tenantID)

	for _, sub := range subs {
		sSubscription := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				P256dh: sub.P256dh,
				Auth:   sub.Auth,
			},
		}

		resp, err := webpush.SendNotification(payloadData, sSubscription, &webpush.Options{
			Subscriber:      s.subscriberSubject,
			VAPIDPublicKey:  s.vapidPublicKey,
			VAPIDPrivateKey: s.vapidPrivateKey,
			TTL:             60,
		})

		if err != nil {
			log.Printf("[PUSH DISPATCH ERROR] Failed to send push to endpoint %s: %v", sub.Endpoint, err)
		} else {
			if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
				log.Printf("[PUSH CLEANUP] Endpoint expired or unsubscribed (%d). Removing subscription ID: %s", resp.StatusCode, sub.ID)
				_ = s.client.PushSubscription.DeleteOneID(sub.ID).Exec(ctx)
			}
			resp.Body.Close()
		}
	}
}
