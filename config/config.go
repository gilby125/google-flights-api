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
	HTTPBindAddr      string
	APIEnabled        bool
	Environment       string
	LoggingConfig     LoggingConfig
	PostgresConfig    PostgresConfig
	Neo4jConfig       Neo4jConfig
	RedisConfig       RedisConfig
	WorkerConfig      WorkerConfig
	FlightConfig      FlightConfig
	LetsEncryptConfig LetsEncryptConfig
	NTFYConfig        NTFYConfig
	AdminAuthConfig   AdminAuthConfig
	WorkerEnabled     bool
	InitSchema        bool
	SeedNeo4j         bool
}

// FlightConfig holds flight search configuration
type FlightConfig struct {
	ExcludedAirlines []string // Airline codes to exclude from results (e.g., NK, G4, F9)
	TopNDeals        int      // Number of top deals to fetch full itineraries for
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
	Host                   string
	Port                   string
	Password               string
	DB                     int
	QueueGroup             string
	QueueStreamPrefix      string
	QueueBlockTimeout      time.Duration
	QueueVisibilityTimeout time.Duration
}

// WorkerConfig holds worker configuration
type WorkerConfig struct {
	Concurrency        int
	MaxRetries         int
	RetryDelay         time.Duration
	JobTimeout         time.Duration
	ShutdownTimeout    time.Duration
	SchedulerLockTTL   time.Duration
	SchedulerLockRenew time.Duration
	SchedulerLockKey   string
}

// NTFYConfig holds NTFY push notification configuration
type NTFYConfig struct {
	ServerURL      string
	Topic          string
	Username       string
	Password       string
	Enabled        bool
	StallThreshold time.Duration
	ErrorThreshold int
	ErrorWindow    time.Duration
}

// AdminAuthConfig holds admin authentication configuration
type AdminAuthConfig struct {
	Enabled  bool
	Username string
	Password string
	Token    string // Alternative: Bearer token auth
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load(".env")

	port := getEnv("PORT", "8080")
	httpBindAddr := getEnv("HTTP_BIND_ADDR", "")
	environment := getEnv("ENVIRONMENT", "development")
	apiEnabled, _ := strconv.ParseBool(getEnv("API_ENABLED", "true"))
	workerEnabled, _ := strconv.ParseBool(getEnv("WORKER_ENABLED", "true"))
	initSchema, _ := strconv.ParseBool(getEnv("INIT_SCHEMA", "true"))
	seedNeo4j, _ := strconv.ParseBool(getEnv("SEED_NEO4J", "true"))

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

	queueBlockTimeout, err := time.ParseDuration(getEnv("REDIS_QUEUE_BLOCK_TIMEOUT", "5s"))
	if err != nil {
		queueBlockTimeout = 5 * time.Second
	}
	queueVisibilityTimeout, err := time.ParseDuration(getEnv("REDIS_QUEUE_VISIBILITY_TIMEOUT", "2m"))
	if err != nil {
		queueVisibilityTimeout = 2 * time.Minute
	}

	redisConfig := RedisConfig{
		Host:                   getEnv("REDIS_HOST", "redis"),
		Port:                   getEnv("REDIS_PORT", "6379"),
		Password:               getEnv("REDIS_PASSWORD", ""),
		DB:                     0,
		QueueGroup:             getEnv("REDIS_QUEUE_GROUP", "flights_workers"),
		QueueStreamPrefix:      getEnv("REDIS_QUEUE_STREAM_PREFIX", "flights"),
		QueueBlockTimeout:      queueBlockTimeout,
		QueueVisibilityTimeout: queueVisibilityTimeout,
	}

	concurrency, _ := strconv.Atoi(getEnv("WORKER_CONCURRENCY", "5"))
	maxRetries, _ := strconv.Atoi(getEnv("WORKER_MAX_RETRIES", "3"))
	retryDelay, _ := time.ParseDuration(getEnv("WORKER_RETRY_DELAY", "30s"))
	jobTimeout, _ := time.ParseDuration(getEnv("WORKER_JOB_TIMEOUT", "10m"))
	shutdownTimeout, _ := time.ParseDuration(getEnv("WORKER_SHUTDOWN_TIMEOUT", "30s"))
	schedulerLockTTL, _ := time.ParseDuration(getEnv("SCHEDULER_LOCK_TTL", "30s"))
	schedulerLockRenew, _ := time.ParseDuration(getEnv("SCHEDULER_LOCK_RENEW", "10s"))
	schedulerLockKey := getEnv("SCHEDULER_LOCK_KEY", "scheduler:leader")

	workerConfig := WorkerConfig{
		Concurrency:        concurrency,
		MaxRetries:         maxRetries,
		RetryDelay:         retryDelay,
		JobTimeout:         jobTimeout,
		ShutdownTimeout:    shutdownTimeout,
		SchedulerLockTTL:   schedulerLockTTL,
		SchedulerLockRenew: schedulerLockRenew,
		SchedulerLockKey:   schedulerLockKey,
	}

	// NTFY notification config
	ntfyEnabled, _ := strconv.ParseBool(getEnv("NTFY_ENABLED", "false"))
	ntfyErrorThreshold, _ := strconv.Atoi(getEnv("NTFY_ERROR_THRESHOLD", "10"))
	ntfyStallThreshold, _ := time.ParseDuration(getEnv("NTFY_STALL_THRESHOLD", "15m"))
	ntfyErrorWindow, _ := time.ParseDuration(getEnv("NTFY_ERROR_WINDOW", "5m"))

	ntfyConfig := NTFYConfig{
		ServerURL:      getEnv("NTFY_SERVER_URL", "https://ntfy.sh"),
		Topic:          getEnv("NTFY_TOPIC", ""),
		Username:       getEnv("NTFY_USERNAME", ""),
		Password:       getEnv("NTFY_PASSWORD", ""),
		Enabled:        ntfyEnabled,
		StallThreshold: ntfyStallThreshold,
		ErrorThreshold: ntfyErrorThreshold,
		ErrorWindow:    ntfyErrorWindow,
	}

	// Admin authentication config
	adminAuthEnabled, _ := strconv.ParseBool(getEnv("ADMIN_AUTH_ENABLED", "false"))
	adminAuthConfig := AdminAuthConfig{
		Enabled:  adminAuthEnabled,
		Username: getEnv("ADMIN_AUTH_USERNAME", ""),
		Password: getEnv("ADMIN_AUTH_PASSWORD", ""),
		Token:    getEnv("ADMIN_AUTH_TOKEN", ""),
	}

	// Flight search config
	// Default excluded airlines: Spirit, Allegiant, Frontier, Sun Country, Avelo, Breeze
	excludedAirlinesStr := getEnv("EXCLUDED_AIRLINES", "NK,G4,F9,SY,XP,MX")
	excludedAirlines := []string{}
	if excludedAirlinesStr != "" {
		for _, code := range strings.Split(excludedAirlinesStr, ",") {
			code = strings.TrimSpace(strings.ToUpper(code))
			if code != "" {
				excludedAirlines = append(excludedAirlines, code)
			}
		}
	}
	topNDeals, _ := strconv.Atoi(getEnv("TOP_N_DEALS", "3"))
	if topNDeals < 1 {
		topNDeals = 3
	}
	flightConfig := FlightConfig{
		ExcludedAirlines: excludedAirlines,
		TopNDeals:        topNDeals,
	}

	return &Config{
		Port:            port,
		HTTPBindAddr:    httpBindAddr,
		APIEnabled:      apiEnabled,
		Environment:     environment,
		LoggingConfig:   loggingConfig,
		PostgresConfig:  postgresConfig,
		Neo4jConfig:     neo4jConfig,
		RedisConfig:     redisConfig,
		WorkerConfig:    workerConfig,
		FlightConfig:    flightConfig,
		NTFYConfig:      ntfyConfig,
		AdminAuthConfig: adminAuthConfig,
		WorkerEnabled:   workerEnabled,
		InitSchema:      initSchema,
		SeedNeo4j:       seedNeo4j,
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
			Host:                   getEnv("REDIS_HOST", "localhost"), // Use env var if set, default to localhost
			Port:                   getEnv("REDIS_PORT", "6379"),      // Use env var if set, default to 6379 (standard Redis)
			QueueGroup:             getEnv("REDIS_QUEUE_GROUP", "flights_workers"),
			QueueStreamPrefix:      getEnv("REDIS_QUEUE_STREAM_PREFIX", "flights"),
			QueueBlockTimeout:      5 * time.Second,
			QueueVisibilityTimeout: 2 * time.Minute,
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
