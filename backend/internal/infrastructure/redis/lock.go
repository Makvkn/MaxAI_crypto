package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

// ErrLockHeld reports that another holder currently owns the lock. Callers
// translate it into a domain-level state instead of starting duplicate work
// (§61, §158).
var ErrLockHeld = errors.New("lock is already held")

// releaseScript deletes the key only when the caller still owns it, so a lock
// that expired and was re-acquired elsewhere is never released by the previous
// owner.
var releaseScript = goredis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

// extendScript refreshes the TTL only for the current owner.
var extendScript = goredis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
`)

// Locker acquires short-lived distributed locks.
type Locker struct {
	client *Client
}

// NewLocker builds a locker on the shared client.
func NewLocker(client *Client) *Locker {
	return &Locker{client: client}
}

// Lock is an acquired lock held by one owner.
type Lock struct {
	client *Client
	key    string
	token  string
}

// Acquire takes the lock at key for ttl, returning ErrLockHeld when another
// owner holds it.
func (l *Locker) Acquire(ctx context.Context, key string, ttl time.Duration) (*Lock, error) {
	token := uuid.NewString()
	ok, err := l.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("acquire lock %s: %w", key, err)
	}
	if !ok {
		return nil, ErrLockHeld
	}
	return &Lock{client: l.client, key: key, token: token}, nil
}

// Release frees the lock if this owner still holds it.
func (lock *Lock) Release(ctx context.Context) error {
	if err := releaseScript.Run(ctx, lock.client, []string{lock.key}, lock.token).Err(); err != nil && !errors.Is(err, goredis.Nil) {
		return fmt.Errorf("release lock %s: %w", lock.key, err)
	}
	return nil
}

// Extend refreshes the lock TTL for long-running work such as a full wallet
// synchronization.
func (lock *Lock) Extend(ctx context.Context, ttl time.Duration) error {
	if err := extendScript.Run(ctx, lock.client, []string{lock.key}, lock.token, ttl.Milliseconds()).Err(); err != nil && !errors.Is(err, goredis.Nil) {
		return fmt.Errorf("extend lock %s: %w", lock.key, err)
	}
	return nil
}
