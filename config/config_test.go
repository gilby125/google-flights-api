package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoad tests the Load function which reads from environment variables.
func TestLoad(t *testing.T) {
	// Clear existing env vars that might interfere
	os.Clearenv()

	t.Run("defaults", func(t *testing.T) {
		cfg, err := Load()
		require.NoError(t, err)

		assert.Equal(t, "8080", cfg.Port)
		assert.Equal(t, "development", cfg.Environment)
		assert.Equal(t, "postgres", cfg.PostgresConfig.Host)
		assert.Equal(t, "5432", cfg.PostgresConfig.Port)
		assert.Equal(t, "flights", cfg.PostgresConfig.User)
		assert.Equal(t, "", cfg.PostgresConfig.Password)
		assert.Equal(t, "flights", cfg.PostgresConfig.DBName)
		assert.Equal(t, "verify-full", cfg.PostgresConfig.SSLMode)
		assert.True(t, cfg.PostgresConfig.RequireSSL)
		assert.Equal(t, "bolt://neo4j:7687", cfg.Neo4jConfig.URI)
		assert.Equal(t, "neo4j", cfg.Neo4jConfig.User)
		assert.Equal(t, "", cfg.Neo4jConfig.Password)
		assert.Equal(t, "redis", cfg.RedisConfig.Host)
		assert.Equal(t, "6379", cfg.RedisConfig.Port)
		assert.Equal(t, "", cfg.RedisConfig.Password)
		assert.Equal(t, 0, cfg.RedisConfig.DB)
		assert.Equal(t, 5, cfg.WorkerConfig.Concurrency)
		assert.Equal(t, 3, cfg.WorkerConfig.MaxRetries)
		assert.Equal(t, 30*time.Second, cfg.WorkerConfig.RetryDelay)
		assert.Equal(t, 5*time.Minute, cfg.WorkerConfig.JobTimeout)
		assert.Equal(t, 30*time.Second, cfg.WorkerConfig.ShutdownTimeout)
		assert.True(t, cfg.WorkerEnabled)
	})

	t.Run("environment variable override", func(t *testing.T) {
		t.Setenv("PORT", "9090")
		t.Setenv("ENVIRONMENT", "production")
		t.Setenv("DB_HOST", "db.example.com")
		t.Setenv("DB_PASSWORD", "secret")
		t.Setenv("DB_SSLMODE", "disable")
		t.Setenv("DB_REQUIRE_SSL", "false")
		t.Setenv("NEO4J_URI", "neo4j://neo.example.com:7687")
		t.Setenv("REDIS_HOST", "cache.example.com")
		t.Setenv("WORKER_CONCURRENCY", "10")
		t.Setenv("WORKER_ENABLED", "false")

		cfg, err := Load()
		require.NoError(t, err)

		assert.Equal(t, "9090", cfg.Port)
		assert.Equal(t, "production", cfg.Environment)
		assert.Equal(t, "db.example.com", cfg.PostgresConfig.Host)
		assert.Equal(t, "secret", cfg.PostgresConfig.Password)
		assert.Equal(t, "disable", cfg.PostgresConfig.SSLMode)
		assert.False(t, cfg.PostgresConfig.RequireSSL)
		assert.Equal(t, "neo4j://neo.example.com:7687", cfg.Neo4jConfig.URI)
		assert.Equal(t, "cache.example.com", cfg.RedisConfig.Host)
		assert.Equal(t, 10, cfg.WorkerConfig.Concurrency)
		assert.False(t, cfg.WorkerEnabled)
	})
}

// TestLoadTestConfig tests the LoadTestConfig helper function
func TestLoadTestConfig(t *testing.T) {
	cfg := LoadTestConfig()

	assert.Equal(t, "test", cfg.Environment)
	assert.Equal(t, "localhost", cfg.PostgresConfig.Host)
	assert.Equal(t, "5432", cfg.PostgresConfig.Port)
	assert.Equal(t, "flights", cfg.PostgresConfig.User) // Expect new default
	assert.Equal(t, "flights_test", cfg.PostgresConfig.DBName)
	assert.Equal(t, "disable", cfg.PostgresConfig.SSLMode)
	assert.Equal(t, "localhost", cfg.RedisConfig.Host)
	assert.Equal(t, "6379", cfg.RedisConfig.Port) // Expect new default
	// WorkerEnabled is not explicitly set by LoadTestConfig, so it retains default bool value (false)
	assert.False(t, cfg.WorkerEnabled)
}

// TestTestConfig tests the TestConfig helper function
func TestTestConfig(t *testing.T) {
	cfg := TestConfig()

	// Inherits from LoadTestConfig
	assert.Equal(t, "test", cfg.Environment)
	assert.Equal(t, "localhost", cfg.PostgresConfig.Host)
	assert.Equal(t, "flights_test", cfg.PostgresConfig.DBName)
	assert.Equal(t, "disable", cfg.PostgresConfig.SSLMode)
	assert.Equal(t, "localhost", cfg.RedisConfig.Host)
	assert.Equal(t, "6379", cfg.RedisConfig.Port) // Expect new default

	// Specifically set by TestConfig
	assert.False(t, cfg.WorkerEnabled)
}
