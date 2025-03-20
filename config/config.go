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
	Port           string
	Environment    string
	PostgresConfig PostgresConfig
	Neo4jConfig    Neo4jConfig
	RedisConfig    RedisConfig
	WorkerConfig   WorkerConfig
	WorkerEnabled  bool
}

// PostgresConfig holds PostgreSQL connection configuration
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
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
	_ = godotenv.Load()
	
	port := getEnv("PORT", "8080")
	environment := getEnv("ENVIRONMENT", "development")
	workerEnabled, _ := strconv.ParseBool(getEnv("WORKER_ENABLED", "true"))

	postgresConfig := PostgresConfig{
		Host:     getEnv("DB_HOST", "postgres"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "flights"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "flights"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
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
		PostgresConfig: postgresConfig,
		Neo4jConfig:    neo4jConfig,
		RedisConfig:    redisConfig,
		WorkerConfig:   workerConfig,
		WorkerEnabled:  workerEnabled,
	}, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(strings.TrimSpace(value)) == 0 {
		return defaultValue
	}
	return value
}
