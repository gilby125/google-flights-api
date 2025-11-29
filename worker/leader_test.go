package worker

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(func() { mr.Close() })

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	t.Cleanup(func() { client.Close() })

	return mr, client
}

func TestLeaderElector_AcquireLock_Success(t *testing.T) {
	_, client := setupTestRedis(t)

	le := NewLeaderElector(
		client,
		"test:leader",
		30*time.Second,
		10*time.Second,
		nil,
		nil,
	)

	// Manually try to acquire lock
	ctx := context.Background()
	acquired := le.tryAcquireLock(ctx)

	assert.True(t, acquired, "Should acquire lock when none exists")

	// Verify lock is set in Redis
	val, err := client.Get(ctx, "test:leader").Result()
	require.NoError(t, err)
	assert.Equal(t, le.instanceID, val)
}

func TestLeaderElector_AcquireLock_AlreadyHeld(t *testing.T) {
	mr, client := setupTestRedis(t)

	// Pre-set the lock with a different instance
	mr.Set("test:leader", "other-instance-123")

	le := NewLeaderElector(
		client,
		"test:leader",
		30*time.Second,
		10*time.Second,
		nil,
		nil,
	)

	ctx := context.Background()
	acquired := le.tryAcquireLock(ctx)

	assert.False(t, acquired, "Should not acquire lock when already held")
}

func TestLeaderElector_RenewLock_Success(t *testing.T) {
	mr, client := setupTestRedis(t)

	le := NewLeaderElector(
		client,
		"test:leader",
		30*time.Second,
		10*time.Second,
		nil,
		nil,
	)

	// Set lock with our instance ID
	mr.Set("test:leader", le.instanceID)

	ctx := context.Background()
	renewed := le.renewLock(ctx)

	assert.True(t, renewed, "Should renew lock when we own it")

	// Verify TTL was set
	ttl := mr.TTL("test:leader")
	assert.Greater(t, ttl, time.Duration(0), "Lock should have TTL after renewal")
}

func TestLeaderElector_RenewLock_LostOwnership(t *testing.T) {
	mr, client := setupTestRedis(t)

	le := NewLeaderElector(
		client,
		"test:leader",
		30*time.Second,
		10*time.Second,
		nil,
		nil,
	)

	// Set lock with a different instance ID (simulating lock takeover)
	mr.Set("test:leader", "other-instance-456")

	ctx := context.Background()
	renewed := le.renewLock(ctx)

	assert.False(t, renewed, "Should not renew lock when owned by another instance")
}

func TestLeaderElector_ReleaseLock(t *testing.T) {
	mr, client := setupTestRedis(t)

	le := NewLeaderElector(
		client,
		"test:leader",
		30*time.Second,
		10*time.Second,
		nil,
		nil,
	)

	// Set lock with our instance ID
	mr.Set("test:leader", le.instanceID)

	ctx := context.Background()
	le.releaseLock(ctx)

	// Verify lock is deleted
	exists := mr.Exists("test:leader")
	assert.False(t, exists, "Lock should be deleted after release")
}

func TestLeaderElector_ReleaseLock_NotOwned(t *testing.T) {
	mr, client := setupTestRedis(t)

	le := NewLeaderElector(
		client,
		"test:leader",
		30*time.Second,
		10*time.Second,
		nil,
		nil,
	)

	// Set lock with a different instance ID
	mr.Set("test:leader", "other-instance-789")

	ctx := context.Background()
	le.releaseLock(ctx)

	// Verify lock is NOT deleted (still owned by other instance)
	val, err := mr.Get("test:leader")
	require.NoError(t, err)
	assert.Equal(t, "other-instance-789", val, "Lock should not be deleted when owned by another instance")
}

func TestLeaderElector_Callbacks(t *testing.T) {
	mr, client := setupTestRedis(t)

	becameLeaderCalled := false
	lostLeaderCalled := false

	le := NewLeaderElector(
		client,
		"test:leader",
		100*time.Millisecond, // Short TTL for testing
		50*time.Millisecond,  // Short renewal interval
		func() { becameLeaderCalled = true },
		func() { lostLeaderCalled = true },
	)

	// Start leader election
	le.Start()

	// Wait for leader acquisition
	time.Sleep(80 * time.Millisecond)

	assert.True(t, le.IsLeader(), "Should be leader after startup")
	assert.True(t, becameLeaderCalled, "onBecomeLeader callback should be called")

	// Simulate losing the lock by setting it to a different value
	// This prevents the leader from renewing it
	mr.Set("test:leader", "another-instance-took-over")

	// Wait for next renewal attempt to detect loss
	time.Sleep(80 * time.Millisecond)

	assert.False(t, le.IsLeader(), "Should lose leadership when lock is owned by another")
	assert.True(t, lostLeaderCalled, "onLoseLeader callback should be called")

	le.Stop()
}

func TestLeaderElector_StartStop(t *testing.T) {
	_, client := setupTestRedis(t)

	le := NewLeaderElector(
		client,
		"test:leader",
		30*time.Second,
		10*time.Second,
		nil,
		nil,
	)

	// Should not panic and should complete quickly
	le.Start()

	// Give it a moment to run
	time.Sleep(50 * time.Millisecond)

	assert.True(t, le.IsLeader(), "Should acquire leadership on start")

	le.Stop()

	// After stop, lock should be released
	ctx := context.Background()
	exists, err := client.Exists(ctx, "test:leader").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists, "Lock should be released after stop")
}

func TestLeaderElector_InstanceID(t *testing.T) {
	_, client := setupTestRedis(t)

	le := NewLeaderElector(
		client,
		"test:leader",
		30*time.Second,
		10*time.Second,
		nil,
		nil,
	)

	instanceID := le.InstanceID()
	assert.NotEmpty(t, instanceID, "Instance ID should not be empty")
	assert.Contains(t, instanceID, "-", "Instance ID should contain hostname-timestamp format")
}

func TestLeaderElector_MultipleInstances(t *testing.T) {
	_, client := setupTestRedis(t)

	leader1Became := false
	leader2Became := false

	le1 := NewLeaderElector(
		client,
		"test:leader",
		100*time.Millisecond,
		30*time.Millisecond,
		func() { leader1Became = true },
		nil,
	)

	le2 := NewLeaderElector(
		client,
		"test:leader",
		100*time.Millisecond,
		30*time.Millisecond,
		func() { leader2Became = true },
		nil,
	)

	// Start both
	le1.Start()
	time.Sleep(50 * time.Millisecond) // Give le1 time to acquire
	le2.Start()

	// Wait for election
	time.Sleep(150 * time.Millisecond)

	// Only one should be leader
	assert.True(t, le1.IsLeader() != le2.IsLeader() || (le1.IsLeader() && !le2.IsLeader()),
		"Exactly one instance should be leader")
	assert.True(t, leader1Became, "First instance should have become leader")
	assert.False(t, leader2Became, "Second instance should not have become leader")

	le1.Stop()
	le2.Stop()
}
