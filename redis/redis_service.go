package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis configuration constants
const (
	RedisURL      = "localhost:6379"
	RedisPassword = ""
	RedisDB       = 0
)

// Service handles all Redis-related operations
type Service struct {
	client *redis.Client
	ctx    context.Context
}

// NewService creates a new Redis service instance
func NewService() *Service {
	client := redis.NewClient(&redis.Options{
		Addr:     RedisURL,
		Password: RedisPassword,
		DB:       RedisDB,
	})

	ctx := context.Background()

	// Test the connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("⚠️ Warning: Redis connection failed: %v", err)
		log.Printf("💡 Make sure Redis is running on %s", RedisURL)
	} else {
		log.Printf("✅ Redis connected successfully to %s", RedisURL)
	}

	return &Service{
		client: client,
		ctx:    ctx,
	}
}

// Close closes the Redis connection
func (r *Service) Close() error {
	return r.client.Close()
}

// Set stores a key-value pair in Redis
func (r *Service) Set(key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}

	err = r.client.Set(r.ctx, key, jsonValue, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %v", key, err)
	}

	log.Printf("📝 Redis SET: %s (expires in %v)", key, expiration)
	return nil
}

// Get retrieves a value from Redis
func (r *Service) Get(key string, dest interface{}) error {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key %s not found", key)
		}
		return fmt.Errorf("failed to get key %s: %v", key, err)
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %v", key, err)
	}

	log.Printf("📖 Redis GET: %s", key)
	return nil
}

// Delete removes a key from Redis
func (r *Service) Delete(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %v", key, err)
	}

	log.Printf("🗑️ Redis DELETE: %s", key)
	return nil
}

// Exists checks if a key exists in Redis
func (r *Service) Exists(key string) (bool, error) {
	result, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %v", key, err)
	}

	return result > 0, nil
}

// SetExpiration sets the expiration time for a key
func (r *Service) SetExpiration(key string, expiration time.Duration) error {
	err := r.client.Expire(r.ctx, key, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration for key %s: %v", key, err)
	}

	log.Printf("⏰ Redis EXPIRE: %s (expires in %v)", key, expiration)
	return nil
}

// GetTTL gets the remaining time to live for a key
func (r *Service) GetTTL(key string) (time.Duration, error) {
	ttl, err := r.client.TTL(r.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %v", key, err)
	}

	return ttl, nil
}

// CacheSession stores user session data in Redis
func (r *Service) CacheSession(sessionID string, sessionData map[string]interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Set(key, sessionData, expiration)
}

// GetSession retrieves user session data from Redis
func (r *Service) GetSession(sessionID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	var sessionData map[string]interface{}
	err := r.Get(key, &sessionData)
	if err != nil {
		return nil, err
	}
	return sessionData, nil
}

// DeleteSession removes user session data from Redis
func (r *Service) DeleteSession(sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Delete(key)
}

// CacheGameList stores game list data in Redis
func (r *Service) CacheGameList(data map[string]interface{}, expiration time.Duration) error {
	key := "gamelist:current"
	return r.Set(key, data, expiration)
}

func (r *Service) CacheListContest(data map[string]interface{}, expiration time.Duration) error {
	key := "listcontest:current"
	return r.Set(key, data, expiration)
}

// GetGameList retrieves game list data from Redis
func (r *Service) GetGameList() (map[string]interface{}, error) {
	key := "gamelist:current"
	var gameList map[string]interface{}
	err := r.Get(key, &gameList)
	if err != nil {
		return nil, err
	}
	return gameList, nil
}

func (r *Service) GetListContest() (map[string]interface{}, error) {
	key := "listcontest:current"
	var listContest map[string]interface{}
	err := r.Get(key, &listContest)
	if err != nil {
		return nil, err
	}
	return listContest, nil
}

// CacheUserData stores user data in Redis
func (r *Service) CacheUserData(userID string, userData map[string]interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("user:%s", userID)
	return r.Set(key, userData, expiration)
}

// GetUserData retrieves user data from Redis
func (r *Service) GetUserData(userID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("user:%s", userID)
	var userData map[string]interface{}
	err := r.Get(key, &userData)
	if err != nil {
		return nil, err
	}
	return userData, nil
}

// DeleteUserData removes user data from Redis
func (r *Service) DeleteUserData(userID string) error {
	key := fmt.Sprintf("user:%s", userID)
	return r.Delete(key)
}

// IncrementCounter increments a counter in Redis
func (r *Service) IncrementCounter(key string) (int64, error) {
	result, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment counter %s: %v", key, err)
	}

	log.Printf("🔢 Redis INCR: %s = %d", key, result)
	return result, nil
}

// GetCounter gets the current value of a counter
func (r *Service) GetCounter(key string) (int64, error) {
	result, err := r.client.Get(r.ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get counter %s: %v", key, err)
	}

	return result, nil
}

// SetCounter sets the value of a counter
func (r *Service) SetCounter(key string, value int64, expiration time.Duration) error {
	err := r.client.Set(r.ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set counter %s: %v", key, err)
	}

	log.Printf("🔢 Redis SET Counter: %s = %d", key, value)
	return nil
}

// FlushAll clears all data from Redis (use with caution)
func (r *Service) FlushAll() error {
	err := r.client.FlushAll(r.ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush all: %v", err)
	}

	log.Printf("🧹 Redis FLUSHALL: All data cleared")
	return nil
}

// GetClient returns the Redis client for advanced operations
func (r *Service) GetClient() *redis.Client {
	return r.client
}

// GetContext returns the Redis context
func (r *Service) GetContext() context.Context {
	return r.ctx
}
