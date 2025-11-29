package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// LeaderElector manages distributed leader election using Redis.
// Only the leader instance runs the scheduler to prevent duplicate job execution.
type LeaderElector struct {
	redisClient    *redis.Client
	lockKey        string
	lockTTL        time.Duration
	renewInterval  time.Duration
	instanceID     string
	isLeader       atomic.Bool
	stopChan       chan struct{}
	wg             sync.WaitGroup
	onBecomeLeader func()
	onLoseLeader   func()
}

// NewLeaderElector creates a new leader elector.
// onBecomeLeader is called when this instance acquires leadership.
// onLoseLeader is called when this instance loses leadership.
func NewLeaderElector(
	redisClient *redis.Client,
	lockKey string,
	lockTTL time.Duration,
	renewInterval time.Duration,
	onBecomeLeader func(),
	onLoseLeader func(),
) *LeaderElector {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "worker"
	}
	instanceID := fmt.Sprintf("%s-%d", hostname, time.Now().UnixNano())

	return &LeaderElector{
		redisClient:    redisClient,
		lockKey:        lockKey,
		lockTTL:        lockTTL,
		renewInterval:  renewInterval,
		instanceID:     instanceID,
		stopChan:       make(chan struct{}),
		onBecomeLeader: onBecomeLeader,
		onLoseLeader:   onLoseLeader,
	}
}

// Start begins the leader election loop.
// It runs in a goroutine and periodically attempts to acquire or renew leadership.
func (le *LeaderElector) Start() {
	le.wg.Add(1)
	go le.electionLoop()
	log.Printf("Leader election started for instance %s (key: %s, TTL: %v, renew: %v)",
		le.instanceID, le.lockKey, le.lockTTL, le.renewInterval)
}

// Stop releases leadership (if held) and stops the election loop.
func (le *LeaderElector) Stop() {
	close(le.stopChan)
	le.wg.Wait()

	// Release lock if we're the leader
	if le.isLeader.Load() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		le.releaseLock(ctx)
		le.isLeader.Store(false)
		log.Printf("Leader election stopped - released lock for instance %s", le.instanceID)
	} else {
		log.Printf("Leader election stopped for instance %s (was not leader)", le.instanceID)
	}
}

// IsLeader returns whether this instance currently holds leadership.
func (le *LeaderElector) IsLeader() bool {
	return le.isLeader.Load()
}

// InstanceID returns the unique identifier for this instance.
func (le *LeaderElector) InstanceID() string {
	return le.instanceID
}

func (le *LeaderElector) electionLoop() {
	defer le.wg.Done()

	// Try to acquire immediately on startup
	le.tryMaintainLeadership()

	ticker := time.NewTicker(le.renewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-le.stopChan:
			return
		case <-ticker.C:
			le.tryMaintainLeadership()
		}
	}
}

func (le *LeaderElector) tryMaintainLeadership() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if le.isLeader.Load() {
		// We're the leader - try to renew
		if !le.renewLock(ctx) {
			log.Printf("Lost leadership: failed to renew lock (instance: %s)", le.instanceID)
			le.isLeader.Store(false)
			if le.onLoseLeader != nil {
				le.onLoseLeader()
			}
		}
	} else {
		// We're not the leader - try to acquire
		if le.tryAcquireLock(ctx) {
			log.Printf("Acquired leadership (instance: %s)", le.instanceID)
			le.isLeader.Store(true)
			if le.onBecomeLeader != nil {
				le.onBecomeLeader()
			}
		}
	}
}

func (le *LeaderElector) tryAcquireLock(ctx context.Context) bool {
	// SET key value NX PX ttl - only set if not exists
	result, err := le.redisClient.SetNX(ctx, le.lockKey, le.instanceID, le.lockTTL).Result()
	if err != nil {
		log.Printf("Error acquiring leader lock: %v", err)
		return false
	}
	return result
}

// Lua script to renew lock only if we own it.
// This ensures atomic check-and-update to prevent race conditions.
var renewScript = redis.NewScript(`
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("PEXPIRE", KEYS[1], ARGV[2])
	else
		return 0
	end
`)

func (le *LeaderElector) renewLock(ctx context.Context) bool {
	result, err := renewScript.Run(ctx, le.redisClient,
		[]string{le.lockKey},
		le.instanceID,
		le.lockTTL.Milliseconds(),
	).Int()

	if err != nil {
		log.Printf("Error renewing leader lock: %v", err)
		return false
	}
	return result == 1
}

// Lua script to release lock only if we own it.
// This prevents accidentally releasing a lock acquired by another instance.
var releaseScript = redis.NewScript(`
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
`)

func (le *LeaderElector) releaseLock(ctx context.Context) {
	result, err := releaseScript.Run(ctx, le.redisClient,
		[]string{le.lockKey},
		le.instanceID,
	).Int()

	if err != nil {
		log.Printf("Error releasing leader lock: %v", err)
	} else if result == 1 {
		log.Printf("Successfully released leader lock (instance: %s)", le.instanceID)
	} else {
		log.Printf("Lock was not held by this instance or already released (instance: %s)", le.instanceID)
	}
}
