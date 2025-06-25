package database

import (
	"context"
	"fmt"
	"gofiber/config"
	"log"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clientInstance *mongo.Client
	clientOnce     sync.Once
	dbName         = "gosocket_db" // Database name for the application
)

// InitDB initializes MongoDB connection
func InitDB() error {
	// Initialize MongoDB client first
	client := InitializeMongoClient(config.MongoDBURL)
	if client == nil {
		return fmt.Errorf("failed to initialize MongoDB client")
	}

	// Wait a moment to ensure connections are established
	time.Sleep(1 * time.Second)

	// Initialize collections and indexes
	if err := InitializeCollections(); err != nil {
		return fmt.Errorf("failed to initialize collections: %v", err)
	}

	fmt.Println("✅ Database services initialized successfully")
	return nil
}

// InitializeMongoClient initializes the MongoDB client
func InitializeMongoClient(mongoURI string) *mongo.Client {
	clientOnce.Do(func() {
		clientOptions := options.Client().ApplyURI(mongoURI)
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}

		// Ping the database to ensure connection is working
		if err := client.Ping(context.TODO(), nil); err != nil {
			log.Fatalf("Failed to ping MongoDB: %v", err)
		}

		fmt.Println("✅ Successfully connected and pinged MongoDB.")
		clientInstance = client
	})
	return clientInstance
}

// GetMongoClient returns the MongoDB client instance
func GetMongoClient() *mongo.Client {
	return clientInstance
}

// GetDatabase returns the database instance
func GetDatabase() *mongo.Database {
	if clientInstance == nil {
		log.Fatal("MongoDB client not initialized. Call InitDB() first.")
	}
	return clientInstance.Database(dbName)
}

// GetUsersCollection returns the users collection
func GetUsersCollection() *mongo.Collection {
	return GetDatabase().Collection("users")
}

// GetSessionsCollection returns the sessions collection
func GetSessionsCollection() *mongo.Collection {
	return GetDatabase().Collection("sessions")
}

// InitializeCollections creates necessary indexes for collections
func InitializeCollections() error {
	// Ensure client is initialized
	if clientInstance == nil {
		return fmt.Errorf("MongoDB client not initialized")
	}

	// Create indexes for users collection
	usersCollection := GetUsersCollection()

	// Create unique index on mobile_no
	_, err := usersCollection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "mobile_no", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		// Check if it's a duplicate key error (index already exists)
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create users index: %v", err)
		}
		fmt.Println("Users index already exists, skipping...")
	}

	// Create indexes for sessions collection
	sessionsCollection := GetSessionsCollection()

	// Create index on session_token
	_, err = sessionsCollection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "session_token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		// Check if it's a duplicate key error (index already exists)
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create sessions index: %v", err)
		}
		fmt.Println("Sessions index already exists, skipping...")
	}

	// Create TTL index for session expiry
	_, err = sessionsCollection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	)
	if err != nil {
		// Check if it's a duplicate key error (index already exists)
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create TTL index: %v", err)
		}
		fmt.Println("TTL index already exists, skipping...")
	}

	// Create compound index for session validation using BSON document
	compoundIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "session_token", Value: 1},
			{Key: "mobile_no", Value: 1},
			{Key: "is_active", Value: 1},
			{Key: "expires_at", Value: 1},
		},
		Options: options.Index().SetName("session_validation_idx"),
	}

	_, err = sessionsCollection.Indexes().CreateOne(
		context.Background(),
		compoundIndex,
	)
	if err != nil {
		// Check if it's a duplicate key error (index already exists)
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create compound index: %v", err)
		}
		fmt.Println("Compound index already exists, skipping...")
	}

	fmt.Println("✅ Database collections and indexes initialized successfully")
	return nil
}

// CloseAllConnections closes MongoDB connection
func CloseAllConnections() {
	// Close MongoDB connection
	if clientInstance != nil {
		clientInstance.Disconnect(context.Background())
		fmt.Println("✅ MongoDB connection closed")
	}
}
