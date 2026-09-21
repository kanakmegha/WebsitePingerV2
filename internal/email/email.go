package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Pass     string
	FromAddr string
}

func LoadConfigFromEnv() Config {
	portStr := os.Getenv("SMTP_PORT")
	port := 587
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	fromAddr := os.Getenv("SMTP_FROM")
	if fromAddr == "" {
		fromAddr = "noreply@pinger.com"
	}

	pass := os.Getenv("SMTP_PASS")
	if pass == "" {
		pass = os.Getenv("SMTP_PASSWORD")
	}

	return Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     port,
		User:     os.Getenv("SMTP_USER"),
		Pass:     pass,
		FromAddr: fromAddr,
	}
}

type Service struct {
	cfg Config
}

func NewService(cfg Config) *Service {
	return &Service{cfg: cfg}
}

// SendAlertEmail sends an HTML incident/recovery alert email to the recipient.
func (s *Service) SendAlertEmail(to string, monitorName string, alertType string, status string, message string) error {
	subject := fmt.Sprintf("[%s] %s Alert: %s", status, alertType, monitorName)

	const tmpl = `
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #090d16; color: #f8fafc; padding: 24px; }
    .card { background-color: #0f172a; border: 1px solid #1e293b; border-radius: 12px; padding: 24px; max-width: 600px; margin: 0 auto; }
    .badge { display: inline-block; padding: 4px 12px; border-radius: 9999px; font-weight: bold; font-size: 12px; text-transform: uppercase; }
    .badge-firing { background-color: rgba(239, 68, 68, 0.2); color: #f87171; border: 1px solid rgba(239, 68, 68, 0.4); }
    .badge-resolved { background-color: rgba(16, 185, 129, 0.2); color: #34d399; border: 1px solid rgba(16, 185, 129, 0.4); }
    .title { font-size: 20px; font-weight: 800; margin-top: 16px; margin-bottom: 8px; color: #ffffff; }
    .desc { font-size: 14px; color: #94a3b8; line-height: 1.5; }
    .footer { margin-top: 24px; border-top: 1px solid #1e293b; pt: 16px; font-size: 12px; color: #64748b; }
  </style>
</head>
<body>
  <div class="card">
    {{if eq .Status "firing"}}
      <span class="badge badge-firing">Incident Firing</span>
    {{else}}
      <span class="badge badge-resolved">Incident Resolved</span>
    {{end}}
    <h1 class="title">{{.MonitorName}}</h1>
    <p class="desc"><strong>Check Type:</strong> {{.AlertType}}</p>
    <p class="desc"><strong>Message:</strong> {{.Message}}</p>
    <div class="footer">
      Pinger Monitoring SaaS • Automated Alert Dispatcher
    </div>
  </div>
</body>
</html>
`

	t, err := template.New("alert_email").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse alert email template: %w", err)
	}

	var bodyBytes bytes.Buffer
	data := struct {
		MonitorName string
		AlertType   string
		Status      string
		Message     string
	}{
		MonitorName: monitorName,
		AlertType:   alertType,
		Status:      status,
		Message:     message,
	}

	if err := t.Execute(&bodyBytes, data); err != nil {
		return fmt.Errorf("failed to execute alert email template: %w", err)
	}

	return s.sendHTML(to, subject, bodyBytes.String())
}

// SendInviteEmail sends an HTML organization invitation email to the recipient with unique subject reference to prevent Gmail threading.
func (s *Service) SendInviteEmail(to string, tenantName string, role string, inviteURL string, shortToken string) error {
	subject := fmt.Sprintf("You're invited to join %s on Pinger (Ref: %s)", tenantName, shortToken)

	const tmpl = `
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #090d16; color: #f8fafc; padding: 24px; }
    .card { background-color: #0f172a; border: 1px solid #1e293b; border-radius: 12px; padding: 24px; max-width: 600px; margin: 0 auto; }
    .title { font-size: 20px; font-weight: 800; margin-bottom: 8px; color: #ffffff; }
    .desc { font-size: 14px; color: #94a3b8; line-height: 1.5; margin-bottom: 24px; }
    .btn { display: inline-block; background-color: #10b981; color: #022c22; font-weight: bold; font-size: 14px; padding: 12px 24px; border-radius: 8px; text-decoration: none; }
    .footer { margin-top: 24px; border-top: 1px solid #1e293b; padding-top: 16px; font-size: 12px; color: #64748b; }
  </style>
</head>
<body>
  <div class="card">
    <h1 class="title">Join {{.TenantName}}</h1>
    <p class="desc">You have been invited to join <strong>{{.TenantName}}</strong> as an <strong>{{.Role}}</strong> on Pinger Website Monitor.</p>
    <a href="{{.InviteURL}}" class="btn">Accept Invitation</a>
    <div class="footer">
      If you did not expect this invitation, you can safely ignore this email.<br>
      Pinger SaaS Workspace • Ref: {{.ShortToken}}
    </div>
  </div>
</body>
</html>
`

	t, err := template.New("invite_email").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse invite email template: %w", err)
	}

	var bodyBytes bytes.Buffer
	data := struct {
		TenantName string
		Role       string
		InviteURL  string
		ShortToken string
	}{
		TenantName: tenantName,
		Role:       role,
		InviteURL:  inviteURL,
		ShortToken: shortToken,
	}

	if err := t.Execute(&bodyBytes, data); err != nil {
		return fmt.Errorf("failed to execute invite email template: %w", err)
	}

	return s.sendHTML(to, subject, bodyBytes.String())
}

func (s *Service) sendHTML(to string, subject string, htmlBody string) error {
	if s.cfg.Host == "" {
		log.Printf("[EMAIL MOCK/DRY-RUN] To: %s | Subject: %s | Body length: %d bytes (SMTP_HOST empty)", to, subject, len(htmlBody))
		return nil
	}

	msgID := fmt.Sprintf("<%d.%d@pinger.local>", time.Now().UnixNano(), os.Getpid())

	headers := make(map[string]string)
	headers["From"] = s.cfg.FromAddr
	headers["To"] = to
	headers["Subject"] = subject
	headers["Message-ID"] = msgID
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""
	headers["X-Entity-Ref-ID"] = msgID

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Pass, s.cfg.Host)
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	if err := smtp.SendMail(addr, auth, s.cfg.User, []string{to}, []byte(message)); err != nil {
		return fmt.Errorf("smtp sendmail failed: %w", err)
	}

	log.Printf("[EMAIL SENT] To: %s | Subject: %s | Message-ID: %s", to, subject, msgID)
	return nil
}
