package worker_registry

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRegistry_PublishAndListActive(t *testing.T) {
	mr := miniredis.RunT(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	t.Cleanup(func() { _ = rdb.Close() })

	reg := New(rdb, "test")
	ctx := context.Background()

	now := time.Now().UTC()
	hb := WorkerHeartbeat{
		ID:            "worker-1",
		Hostname:      "host-a",
		Status:        "active",
		CurrentJob:    "",
		ProcessedJobs: 12,
		Concurrency:   5,
		StartedAt:     now.Add(-10 * time.Minute),
		LastHeartbeat: now,
		Version:       "1.0.0",
	}
	require.NoError(t, reg.Publish(ctx, hb, 30*time.Second))

	active, err := reg.ListActive(ctx, 35*time.Second, 100)
	require.NoError(t, err)
	require.Len(t, active, 1)
	require.Equal(t, hb.ID, active[0].ID)
	require.Equal(t, hb.Hostname, active[0].Hostname)
	require.Equal(t, hb.Status, active[0].Status)
	require.Equal(t, hb.ProcessedJobs, active[0].ProcessedJobs)
	require.Equal(t, hb.Concurrency, active[0].Concurrency)
	require.Equal(t, hb.Version, active[0].Version)
}
