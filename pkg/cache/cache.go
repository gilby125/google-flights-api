package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache interface defines caching operations
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
}

// RedisCache implements Cache using Redis
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

// prefixKey adds the cache prefix to a key
func (c *RedisCache) prefixKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

// Get retrieves a value from cache
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, c.prefixKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}
	return []byte(val), nil
}

// Set stores a value in cache with TTL
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := c.client.Set(ctx, c.prefixKey(key), value, ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

// Delete removes a value from cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, c.prefixKey(key)).Err()
	if err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}
	return nil
}

// Exists checks if a key exists in cache
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, c.prefixKey(key)).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists error: %w", err)
	}
	return count > 0, nil
}

// Clear removes all keys with the cache prefix
func (c *RedisCache) Clear(ctx context.Context) error {
	pattern := c.prefixKey("*")
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("redis clear error: %w", err)
		}
	}
	return iter.Err()
}

// CacheManager provides high-level caching operations
type CacheManager struct {
	cache Cache
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cache Cache) *CacheManager {
	return &CacheManager{cache: cache}
}

// GetJSON retrieves and unmarshals JSON data from cache
func (cm *CacheManager) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := cm.cache.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// SetJSON marshals and stores JSON data in cache
func (cm *CacheManager) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}
	return cm.cache.Set(ctx, key, data, ttl)
}

// GetOrSet retrieves from cache, or executes function and caches result
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	var result interface{}
	err := cm.GetJSON(ctx, key, &result)
	if err == nil {
		return result, nil
	}
	if err != ErrCacheMiss {
		// Log cache error but continue with function execution
		return nil, fmt.Errorf("cache get error: %w", err)
	}

	// Cache miss - execute function
	result, err = fn()
	if err != nil {
		return nil, err
	}

	// Store in cache (don't fail if cache set fails)
	if setErr := cm.SetJSON(ctx, key, result, ttl); setErr != nil {
		// Log cache set error but return the result
		// In a real implementation, you'd use your logger here
	}

	return result, nil
}

// Delete removes a key from cache
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	return cm.cache.Delete(ctx, key)
}

// Exists checks if a key exists in cache
func (cm *CacheManager) Exists(ctx context.Context, key string) (bool, error) {
	return cm.cache.Exists(ctx, key)
}

// Clear removes all cached data
func (cm *CacheManager) Clear(ctx context.Context) error {
	return cm.cache.Clear(ctx)
}

// Cache policies and TTLs
const (
	ShortTTL  = 5 * time.Minute
	MediumTTL = 1 * time.Hour
	LongTTL   = 24 * time.Hour
)

// Cache key generators
func AirportsKey() string {
	return "airports:all"
}

func AirlinesKey() string {
	return "airlines:all"
}

func FlightSearchKey(origin, destination, date, class string, adults int) string {
	return fmt.Sprintf("flight_search:%s:%s:%s:%s:%d", origin, destination, date, class, adults)
}

func PriceHistoryKey(origin, destination string) string {
	return fmt.Sprintf("price_history:%s:%s", origin, destination)
}

// Error definitions
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)
