package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLockManager struct {
	client *redis.Client
}

// NewRedisLockManager constructs a thread-safe Redis client instance
func NewRedisLockManager(redisAddress string) *RedisLockManager {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", // Default empty config matching compose topology
		DB:       0,  // Use default database partition
	})
	return &RedisLockManager{client: rdb}
}

// AcquireJobLock attempts an atomic SET NX action to lock job processing windows
func (rl *RedisLockManager) AcquireJobLock(ctx context.Context, jobID string, ttl time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("lock:job:%s", jobID)

	// SETNX returns true if the key did not exist and was successfully set
	success, err := rl.client.SetNX(ctx, lockKey, "processing", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis lock operation execution failure: %w", err)
	}
	return success, nil
}

// ReleaseJobLock explicitly cleans up the execution key token
func (rl *RedisLockManager) ReleaseJobLock(ctx context.Context, jobID string) error {
	lockKey := fmt.Sprintf("lock:job:%s", jobID)
	return rl.client.Del(ctx, lockKey).Err()
}

func (rl *RedisLockManager) Close() error {
	return rl.client.Close()
}
