package database

import (
	"fmt"
	"gofiber/config"
	"log"
	"time"

	"github.com/gocql/gocql"
)

var (
	// Cassandra session instance
	CassandraSession *gocql.Session
)

// InitDB initializes Cassandra connection
func InitDB() error {
	if err := InitCassandra(); err != nil {
		return fmt.Errorf("failed to initialize Cassandra: %v", err)
	}
	fmt.Println("✅ Database services initialized successfully")
	return nil
}

// InitCassandra initializes the Cassandra session
func InitCassandra() error {
	// Create cluster configuration
	cluster := gocql.NewCluster(config.CassandraHost)
	cluster.Port = config.CassandraPort
	cluster.Keyspace = config.CassandraKeyspace
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.CassandraUsername,
		Password: config.CassandraPassword,
	}

	// Set consistency and timeout
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second

	// Enable retry policy
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{
		NumRetries: 3,
	}

	// Enable connection pooling
	cluster.NumConns = 10
	cluster.MaxWaitSchemaAgreement = 2 * time.Minute

	log.Printf("🔌 Connecting to Cassandra at %s:%d...", config.CassandraHost, config.CassandraPort)

	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to connect to Cassandra: %v", err)
	}

	// Test the connection
	if err := session.Query("SELECT release_version FROM system.local").Exec(); err != nil {
		return fmt.Errorf("failed to test Cassandra connection: %v", err)
	}

	CassandraSession = session
	log.Printf("✅ Cassandra session initialized successfully")
	log.Printf("📊 Connected to keyspace: %s", config.CassandraKeyspace)

	return nil
}

// CloseAllConnections closes Cassandra connection
func CloseAllConnections() {
	if CassandraSession != nil {
		CassandraSession.Close()
		log.Println("✅ Cassandra connection closed")
	}
}

// GetSession returns the current Cassandra session
func GetSession() *gocql.Session {
	return CassandraSession
}

// HealthCheck performs a health check on the database
func HealthCheck() error {
	if CassandraSession == nil {
		return fmt.Errorf("Cassandra session is not initialized")
	}

	// Simple health check query
	return CassandraSession.Query("SELECT release_version FROM system.local").Exec()
}
