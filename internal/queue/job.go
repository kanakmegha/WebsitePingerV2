package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const QueueName = "pinger:monitor_jobs"

type CheckJob struct {
	MonitorID  uuid.UUID `json:"monitor_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	CheckType  string    `json:"check_type"`
	Domain     string    `json:"domain"`
	URL        string    `json:"url"`
	TimeoutSec int       `json:"timeout_sec"`
	Timestamp  time.Time `json:"timestamp"`
}

type RedisQueue struct {
	rdb *redis.Client
}

func NewRedisQueue(rdb *redis.Client) *RedisQueue {
	return &RedisQueue{rdb: rdb}
}

func (q *RedisQueue) Enqueue(ctx context.Context, job CheckJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	return q.rdb.LPush(ctx, QueueName, data).Err()
}

func (q *RedisQueue) Dequeue(ctx context.Context, timeout time.Duration) (*CheckJob, error) {
	res, err := q.rdb.BRPop(ctx, timeout, QueueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Timeout
		}
		return nil, err
	}

	if len(res) < 2 {
		return nil, fmt.Errorf("invalid redis brpop response")
	}

	var job CheckJob
	if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}
