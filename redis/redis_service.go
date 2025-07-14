package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Service handles all Redis-related operations
type Service struct {
	client *redis.Client
	ctx    context.Context
}

// getRedisConfig gets Redis configuration from environment variables
func getRedisConfig() (string, string, int) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	dbStr := os.Getenv("REDIS_DB")
	db := 0
	if dbStr != "" {
		if dbInt, err := strconv.Atoi(dbStr); err == nil {
			db = dbInt
		}
	}

	return url, password, db
}

// NewService creates a new Redis service instance
func NewService() *Service {
	url, password, db := getRedisConfig()

	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       db,
		// Connection pool settings
		PoolSize:     10,
		MinIdleConns: 5,
		// Timeout settings
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx := context.Background()

	// Test the connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		// Silent fail - Redis might not be available
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
	return nil
}

// Delete removes a key from Redis
func (r *Service) Delete(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %v", key, err)
	}
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
	return nil
}

// FlushAll clears all data from Redis (use with caution)
func (r *Service) FlushAll() error {
	err := r.client.FlushAll(r.ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush all: %v", err)
	}
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

// ConnectionData represents active connection data stored in Redis
type ConnectionData struct {
	SocketID     string    `json:"socket_id"`
	UserID       string    `json:"user_id"`
	MobileNo     string    `json:"mobile_no"`
	SessionToken string    `json:"session_token"`
	DeviceID     string    `json:"device_id"`
	FCMToken     string    `json:"fcm_token"`
	IsActive     bool      `json:"is_active"`
	ConnectedAt  time.Time `json:"connected_at"`
	LastSeen     time.Time `json:"last_seen"`
	UserAgent    string    `json:"user_agent,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	Namespace    string    `json:"namespace,omitempty"`
}

// CacheConnection stores connection data in Redis for server-to-client messaging
func (r *Service) CacheConnection(connectionData ConnectionData, expiration time.Duration) error {
	key := fmt.Sprintf("connection:%s", connectionData.SocketID)
	return r.Set(key, connectionData, expiration)
}

// GetConnection retrieves connection data from Redis
func (r *Service) GetConnection(socketID string) (*ConnectionData, error) {
	key := fmt.Sprintf("connection:%s", socketID)
	var connectionData ConnectionData
	err := r.Get(key, &connectionData)
	if err != nil {
		return nil, err
	}
	return &connectionData, nil
}

// DeleteConnection removes connection data from Redis
func (r *Service) DeleteConnection(socketID string) error {
	key := fmt.Sprintf("connection:%s", socketID)
	return r.Delete(key)
}

// UpdateConnectionLastSeen updates the last seen timestamp for a connection
func (r *Service) UpdateConnectionLastSeen(socketID string) error {
	connectionData, err := r.GetConnection(socketID)
	if err != nil {
		return err
	}

	connectionData.LastSeen = time.Now()
	key := fmt.Sprintf("connection:%s", socketID)
	return r.Set(key, connectionData, 24*time.Hour)
}

// GetConnectionsByUserID retrieves all active connections for a specific user
func (r *Service) GetConnectionsByUserID(userID string) ([]ConnectionData, error) {
	// Note: This is a simplified implementation
	// In production, you might want to maintain a separate index
	pattern := "connection:*"
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection keys: %v", err)
	}

	var connections []ConnectionData
	for _, key := range keys {
		var connectionData ConnectionData
		err := r.Get(key, &connectionData)
		if err == nil && connectionData.UserID == userID && connectionData.IsActive {
			connections = append(connections, connectionData)
		}
	}

	return connections, nil
}

// GetConnectionsByMobileNo retrieves all active connections for a specific mobile number
func (r *Service) GetConnectionsByMobileNo(mobileNo string) ([]ConnectionData, error) {
	pattern := "connection:*"
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection keys: %v", err)
	}

	var connections []ConnectionData
	for _, key := range keys {
		var connectionData ConnectionData
		err := r.Get(key, &connectionData)
		if err == nil && connectionData.MobileNo == mobileNo && connectionData.IsActive {
			connections = append(connections, connectionData)
		}
	}

	return connections, nil
}

// GetAllActiveConnections retrieves all active connections
func (r *Service) GetAllActiveConnections() ([]ConnectionData, error) {
	pattern := "connection:*"
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection keys: %v", err)
	}

	var connections []ConnectionData
	for _, key := range keys {
		var connectionData ConnectionData
		err := r.Get(key, &connectionData)
		if err == nil && connectionData.IsActive {
			connections = append(connections, connectionData)
		}
	}

	return connections, nil
}

// GetConnectionCount returns the total number of active connections
func (r *Service) GetConnectionCount() (int64, error) {
	pattern := "connection:*"
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get connection keys: %v", err)
	}

	var count int64
	for _, key := range keys {
		var connectionData ConnectionData
		err := r.Get(key, &connectionData)
		if err == nil && connectionData.IsActive {
			count++
		}
	}

	return count, nil
}

// CleanupInactiveConnections removes connections that haven't been seen recently
func (r *Service) CleanupInactiveConnections(maxIdleTime time.Duration) error {
	pattern := "connection:*"
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get connection keys: %v", err)
	}

	cutoffTime := time.Now().Add(-maxIdleTime)
	var cleanedCount int64

	for _, key := range keys {
		var connectionData ConnectionData
		err := r.Get(key, &connectionData)
		if err == nil && connectionData.LastSeen.Before(cutoffTime) {
			r.Delete(key)
			cleanedCount++
		}
	}
	return nil
}
