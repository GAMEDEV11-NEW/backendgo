// config/config.go
package config

// Configuration constants for the application
var (
	// MongoDBURL is the connection string for MongoDB
	// For local development, use: "mongodb://localhost:27017"
	// For production, use: "mongodb://username:password@host:port/database"
	MongoDBURL = "mongodb://localhost:27017"

	// Redis configuration
	RedisURL      = "localhost:6379"
	RedisPassword = ""
	RedisDB       = 0

	// ServerPort is the port on which the server will run
	ServerPort = 8088

	// DatabaseName is the name of the MongoDB database
	DatabaseName = "gosocket_db"
)
