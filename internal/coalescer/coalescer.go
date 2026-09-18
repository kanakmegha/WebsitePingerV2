package coalescer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type Coalescer struct {
	sf  singleflight.Group
	rdb *redis.Client
}

func NewCoalescer(rdb *redis.Client) *Coalescer {
	return &Coalescer{rdb: rdb}
}

// Do executes fetchFunc for (checkType, domain) ensuring that only ONE request runs across nodes.
// If another process or node is already fetching, this blocks until completion and returns the shared result.
func (c *Coalescer) Do(ctx context.Context, checkType string, domain string, fetchFunc func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	key := fmt.Sprintf("fetch:%s:%s", checkType, domain)
	channelName := fmt.Sprintf("channel:fetch:%s:%s", checkType, domain)
	lockKey := fmt.Sprintf("lock:fetch:%s:%s", checkType, domain)

	// Step 1: Intra-node coalescing using Go singleflight
	val, err, _ := c.sf.Do(key, func() (interface{}, error) {
		// Step 2: Inter-node distributed lock attempt using Redis
		acquired, err := c.rdb.SetNX(ctx, lockKey, "locked", 30*time.Second).Result()
		if err != nil {
			// Fallback to local fetch if Redis fails
			return fetchFunc(ctx)
		}

		if acquired {
			// Current node is the primary fetcher
			defer c.rdb.Del(ctx, lockKey)

			res, fetchErr := fetchFunc(ctx)
			if fetchErr != nil {
				return nil, fetchErr
			}

			// Broadcast result to other waiting worker nodes via Redis Pub/Sub
			payload, jsonErr := json.Marshal(res)
			if jsonErr == nil {
				c.rdb.Publish(ctx, channelName, payload)
			}

			return res, nil
		}

		// Another node is already fetching; subscribe to Redis Pub/Sub for the result
		pubsub := c.rdb.Subscribe(ctx, channelName)
		defer pubsub.Close()

		ch := pubsub.Channel()

		select {
		case msg := <-ch:
			var res interface{}
			if unmarshalErr := json.Unmarshal([]byte(msg.Payload), &res); unmarshalErr == nil {
				return res, nil
			}
			return nil, fmt.Errorf("failed to unmarshal coalesced result from pubsub")
		case <-time.After(15 * time.Second):
			// Timeout waiting for peer worker; fallback to local execution
			return fetchFunc(ctx)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})

	return val, err
}
