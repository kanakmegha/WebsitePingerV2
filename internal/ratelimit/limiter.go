package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CheckType represents the type of website inspection.
type CheckType string

const (
	CheckTypeHTTP  CheckType = "http"
	CheckTypeDNS   CheckType = "dns"
	CheckTypeSSL   CheckType = "ssl"
	CheckTypeWHOIS CheckType = "whois"
)

// MinIntervals enforces strict minimum fetch intervals regardless of user settings.
var MinIntervals = map[CheckType]time.Duration{
	CheckTypeHTTP:  30 * time.Second,  // Configurable, >= 30 seconds
	CheckTypeDNS:   5 * time.Minute,   // Minimum 5 minutes
	CheckTypeSSL:   1 * time.Hour,     // Minimum 1 hour
	CheckTypeWHOIS: 24 * time.Hour,    // Minimum 24 hours
}

type RateLimiter struct {
	rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

// CanExecute checks if an outbound check for a given domain and check_type can proceed.
// If allowed, it sets the rate limit lock key for the mandatory minimum duration and returns true.
func (rl *RateLimiter) CanExecute(ctx context.Context, checkType CheckType, domain string) (bool, time.Duration, error) {
	minInterval, exists := MinIntervals[checkType]
	if !exists {
		minInterval = 30 * time.Second
	}

	key := fmt.Sprintf("ratelimit:%s:%s", checkType, domain)

	// Attempt atomic SETNX with expiration equal to minimum interval
	set, err := rl.rdb.SetNX(ctx, key, "1", minInterval).Result()
	if err != nil {
		return false, 0, fmt.Errorf("redis setnx error: %w", err)
	}

	if set {
		// Rate limit lock successfully acquired
		return true, 0, nil
	}

	// Rate limit hit - retrieve remaining TTL
	ttl, err := rl.rdb.TTL(ctx, key).Result()
	if err != nil {
		return false, 0, nil
	}

	return false, ttl, nil
}
