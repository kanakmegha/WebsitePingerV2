package checker

import (
	"context"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

type DomainResult struct {
	ExpiryDate    time.Time `json:"expiry_date"`
	Registrar     string    `json:"registrar"`
	DaysRemaining int       `json:"days_remaining"`
	Error         string    `json:"error,omitempty"`
}

func CheckWHOIS(ctx context.Context, domain string) (*DomainResult, error) {
	rawWhois, err := whois.Whois(domain)
	if err != nil {
		return &DomainResult{Error: err.Error()}, nil
	}

	parsed, err := whoisparser.Parse(rawWhois)
	if err != nil {
		return &DomainResult{Error: err.Error()}, nil
	}

	var expiryDate time.Time
	if parsed.Domain != nil && parsed.Domain.ExpirationDate != "" {
		expiryDate, _ = time.Parse(time.RFC3339, parsed.Domain.ExpirationDate)
	}

	registrar := "Unknown"
	if parsed.Registrar != nil && parsed.Registrar.Name != "" {
		registrar = parsed.Registrar.Name
	}

	daysRemaining := 0
	if !expiryDate.IsZero() {
		daysRemaining = int(time.Until(expiryDate).Hours() / 24)
	}

	return &DomainResult{
		ExpiryDate:    expiryDate,
		Registrar:     registrar,
		DaysRemaining: daysRemaining,
	}, nil
}
