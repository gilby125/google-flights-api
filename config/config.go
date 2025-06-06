package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Port              string
	Environment       string
	LoggingConfig     LoggingConfig
	PostgresConfig    PostgresConfig
	Neo4jConfig       Neo4jConfig
	RedisConfig       RedisConfig
	WorkerConfig      WorkerConfig
	LetsEncryptConfig LetsEncryptConfig
	WorkerEnabled     bool
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// LetsEncryptConfig holds ACME/LetsEncrypt configuration
type LetsEncryptConfig struct {
	Email              string `split_words:"true"`
	AcceptTOS          bool   `split_words:"true" default:"false"`
	CacheDir           string `split_words:"true" default:"./certcache"`
	CloudflareAPIToken string `split_words:"true" json:"-"`
	CloudflareZoneID   string `split_words:"true"`
	DNSPropagationWait int    `default:"30"`
}

// PostgresConfig holds PostgreSQL connection configuration
type PostgresConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	SSLMode     string
	SSLCert     string `env:"DB_SSL_CERT" env-default:""`
	SSLKey      string `env:"DB_SSL_KEY" env-default:""`
	SSLRootCert string `env:"DB_SSL_ROOT_CERT" env-default:""`
	RequireSSL  bool   `env:"DB_REQUIRE_SSL" env-default:"true"`
}

// Neo4jConfig holds Neo4j connection configuration
type Neo4jConfig struct {
	URI      string
	User     string
	Password string
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// WorkerConfig holds worker configuration
type WorkerConfig struct {
	Concurrency     int
	MaxRetries      int
	RetryDelay      time.Duration
	JobTimeout      time.Duration
	ShutdownTimeout time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	// Explicitly load .env from the parent directory relative to the 'google-flights-api' execution context
	_ = godotenv.Load("../.env")

	port := getEnv("PORT", "8080")
	environment := getEnv("ENVIRONMENT", "development")
	workerEnabled, _ := strconv.ParseBool(getEnv("WORKER_ENABLED", "true"))

	loggingConfig := LoggingConfig{
		Level:  getEnv("LOG_LEVEL", "info"),
		Format: getEnv("LOG_FORMAT", "json"),
	}

	postgresConfig := PostgresConfig{
		Host:        getEnv("DB_HOST", "postgres"),
		Port:        getEnv("DB_PORT", "5432"),
		User:        getEnv("DB_USER", "flights"),
		Password:    getEnv("DB_PASSWORD", ""),
		DBName:      getEnv("DB_NAME", "flights"),
		SSLMode:     getEnv("DB_SSLMODE", "verify-full"),
		SSLCert:     getEnv("DB_SSL_CERT", ""),
		SSLKey:      getEnv("DB_SSL_KEY", ""),
		SSLRootCert: getEnv("DB_SSL_ROOT_CERT", ""),
		RequireSSL:  getEnv("DB_REQUIRE_SSL", "true") == "true",
	}

	neo4jConfig := Neo4jConfig{
		URI:      getEnv("NEO4J_URI", "bolt://neo4j:7687"),
		User:     getEnv("NEO4J_USER", "neo4j"),
		Password: getEnv("NEO4J_PASSWORD", ""),
	}

	redisConfig := RedisConfig{
		Host:     getEnv("REDIS_HOST", "redis"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	}

	concurrency, _ := strconv.Atoi(getEnv("WORKER_CONCURRENCY", "5"))
	maxRetries, _ := strconv.Atoi(getEnv("WORKER_MAX_RETRIES", "3"))
	retryDelay, _ := time.ParseDuration(getEnv("WORKER_RETRY_DELAY", "30s"))
	jobTimeout, _ := time.ParseDuration(getEnv("WORKER_JOB_TIMEOUT", "5m"))
	shutdownTimeout, _ := time.ParseDuration(getEnv("WORKER_SHUTDOWN_TIMEOUT", "30s"))

	workerConfig := WorkerConfig{
		Concurrency:     concurrency,
		MaxRetries:      maxRetries,
		RetryDelay:      retryDelay,
		JobTimeout:      jobTimeout,
		ShutdownTimeout: shutdownTimeout,
	}

	return &Config{
		Port:           port,
		Environment:    environment,
		LoggingConfig:  loggingConfig,
		PostgresConfig: postgresConfig,
		Neo4jConfig:    neo4jConfig,
		RedisConfig:    redisConfig,
		WorkerConfig:   workerConfig,
		WorkerEnabled:  workerEnabled,
	}, nil
}

// LoadTestConfig loads test configuration
func LoadTestConfig() *Config {
	return &Config{
		PostgresConfig: PostgresConfig{
			Host:     getEnv("DB_HOST", "localhost"),         // Use env var if set, default to localhost
			Port:     getEnv("DB_PORT", "5432"),              // Use env var if set, default to 5432
			User:     getEnv("DB_USER", "flights"),           // Match docker-compose/Load defaults
			Password: getEnv("DB_PASSWORD", ""),              // Load password from env
			DBName:   getEnv("DB_NAME_TEST", "flights_test"), // Use separate test DB name env var
			SSLMode:  getEnv("DB_SSLMODE", "disable"),        // Allow override, default disable for tests
		},
		RedisConfig: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"), // Use env var if set, default to localhost
			Port: getEnv("REDIS_PORT", "6379"),      // Use env var if set, default to 6379 (standard Redis)
		},
		Neo4jConfig: Neo4jConfig{
			URI:      getEnv("NEO4J_URI", "bolt://localhost:7687"), // Use env var or default
			User:     getEnv("NEO4J_USER", "neo4j"),                // Use env var or default
			Password: getEnv("NEO4J_PASSWORD", ""),                 // Use env var or default
		},
		Environment: "test",
	}
}

// TestConfig returns a default test configuration
func TestConfig() *Config {
	cfg := LoadTestConfig()
	cfg.WorkerEnabled = false
	return cfg
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(strings.TrimSpace(value)) == 0 {
		return defaultValue
	}
	return strings.TrimSpace(value) // Trim whitespace before returning
}
