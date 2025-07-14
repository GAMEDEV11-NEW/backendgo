// config/config.go
package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Configuration constants for the application
var (
	// Cassandra configuration
	CassandraHost     string
	CassandraUsername string
	CassandraPassword string
	CassandraKeyspace string
	CassandraPort     int

	// Redis configuration
	RedisURL      string
	RedisPassword string
	RedisDB       int

	// ServerPort is the port on which the server will run
	ServerPort int

	// Application configuration
	AppName    = "GOSOCKET"
	AppVersion = "1.0.0"
)

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
	}

	// Cassandra configuration
	CassandraHost = getEnv("CASSANDRA_HOST", "localhost")
	CassandraUsername = getEnv("CASSANDRA_USERNAME", "cassandra")
	CassandraPassword = getEnv("CASSANDRA_PASSWORD", "cassandra")
	CassandraKeyspace = getEnv("CASSANDRA_KEYSPACE", "myapp")

	portStr := getEnv("CASSANDRA_PORT", "9042")
	if port, err := strconv.Atoi(portStr); err == nil {
		CassandraPort = port
	} else {
		CassandraPort = 9042
	}

	// Server configuration
	portStr = getEnv("SERVER_PORT", "8088")
	if port, err := strconv.Atoi(portStr); err == nil {
		ServerPort = port
	} else {
		ServerPort = 8088
	}

	// Redis configuration
	RedisURL = getEnv("REDIS_URL", "localhost:6379")
	RedisPassword = getEnv("REDIS_PASSWORD", "")
	redisDBStr := getEnv("REDIS_DB", "0")
	if db, err := strconv.Atoi(redisDBStr); err == nil {
		RedisDB = db
	} else {
		RedisDB = 0
	}

}

// getEnv gets environment variable with fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
