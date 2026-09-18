package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

type SSLResult struct {
	ExpiryDate    time.Time `json:"expiry_date"`
	Issuer        string    `json:"issuer"`
	Valid         bool      `json:"valid"`
	DaysRemaining int       `json:"days_remaining"`
	Error         string    `json:"error,omitempty"`
}

func CheckSSL(ctx context.Context, domain string, timeoutSec int) (*SSLResult, error) {
	if timeoutSec <= 0 {
		timeoutSec = 10
	}

	dialer := &net.Dialer{Timeout: time.Duration(timeoutSec) * time.Second}
	targetAddr := fmt.Sprintf("%s:443", domain)

	conn, err := tls.DialWithDialer(dialer, "tcp", targetAddr, &tls.Config{
		ServerName:         domain,
		InsecureSkipVerify: false, // Verify certificate authenticity
	})

	if err != nil {
		// Retry with skip verify to inspect expired/untrusted cert details
		conn, err = tls.DialWithDialer(dialer, "tcp", targetAddr, &tls.Config{
			ServerName:         domain,
			InsecureSkipVerify: true,
		})
		if err != nil {
			return &SSLResult{Valid: false, Error: err.Error()}, nil
		}
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return &SSLResult{Valid: false, Error: "no peer certificates found"}, nil
	}

	cert := certs[0]
	now := time.Now()
	daysRemaining := int(time.Until(cert.NotAfter).Hours() / 24)
	isValid := now.After(cert.NotBefore) && now.Before(cert.NotAfter)

	issuer := cert.Issuer.CommonName
	if issuer == "" && len(cert.Issuer.Organization) > 0 {
		issuer = cert.Issuer.Organization[0]
	}

	return &SSLResult{
		ExpiryDate:    cert.NotAfter,
		Issuer:        issuer,
		Valid:         isValid,
		DaysRemaining: daysRemaining,
	}, nil
}
