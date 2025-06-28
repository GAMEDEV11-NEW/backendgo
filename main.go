// main.go
package main

import (
	"fmt"
	"gofiber/app/routes"
	"gofiber/app/services"
	"gofiber/config"
	"gofiber/database"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	log.Printf("🚀 DEBUG: Starting server initialization...")
	
	app := fiber.New(fiber.Config{
		Prefork:       false,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Fiber",
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			log.Printf("❌ DEBUG: Fiber error handler called: %v", err)
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			ctx.Status(code)
			return ctx.JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})
	log.Printf("✅ DEBUG: Fiber app created successfully")

	// Initialize database first
	log.Printf("🔌 DEBUG: Initializing database connection...")
	if err := database.InitDB(); err != nil {
		log.Printf("❌ DEBUG: Failed to connect to the database: %v", err)
		log.Fatalf("❌ Failed to connect to the database: %v", err)
	}
	log.Printf("✅ DEBUG: Database initialized successfully")

	// Initialize socket service with Cassandra session
	log.Printf("🔧 DEBUG: Initializing socket service...")
	socketService := services.NewSocketService(database.CassandraSession)
	log.Printf("✅ DEBUG: Socket service initialized")

	// Initialize Socket.IO handler with socket service
	log.Printf("🔌 DEBUG: Initializing Socket.IO handler...")
	socketHandler := config.NewSocketHandler(socketService)
	log.Printf("✅ DEBUG: Socket.IO handler initialized")

	// Setup Socket.IO routes (this should be before regular routes)
	log.Printf("🔌 DEBUG: Setting up Socket.IO routes...")
	socketHandler.SetupSocketRoutes(app)
	log.Printf("✅ DEBUG: Socket.IO routes setup completed")

	// Initialize regular routes
	log.Printf("🔌 DEBUG: Setting up regular routes...")
	routes.SetupRoutes(app)
	log.Printf("✅ DEBUG: Regular routes setup completed")

	// Start background cleanup goroutine
	log.Printf("🧹 DEBUG: Starting background cleanup service...")
	go func() {
		log.Printf("🧹 DEBUG: Background cleanup goroutine started")
		ticker := time.NewTicker(5 * time.Minute) // Run every 5 minutes
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				log.Printf("🧹 DEBUG: Running background cleanup...")
				// Cleanup expired sessions
				if err := socketService.CleanupExpiredSessions(); err != nil {
					log.Printf("⚠️ DEBUG: Failed to cleanup expired sessions: %v", err)
				} else {
					log.Printf("✅ DEBUG: Expired sessions cleanup completed")
				}

				// Cleanup expired OTPs
				if err := socketService.CleanupExpiredOTPs(); err != nil {
					log.Printf("⚠️ DEBUG: Failed to cleanup expired OTPs: %v", err)
				} else {
					log.Printf("✅ DEBUG: Expired OTPs cleanup completed")
				}
			}
		}
	}()
	log.Printf("🧹 DEBUG: Background cleanup service started successfully")

	port := config.ServerPort
	log.Printf("🚀 DEBUG: Server starting on port :%d", port)
	log.Printf("🔌 DEBUG: Socket.IO server available at :%d/socket.io", port)
	log.Printf("🎮 DEBUG: Gameplay namespace available at :%d/socket.io/gameplay", port)
	log.Printf("✅ DEBUG: Server initialization completed, starting to listen...")

	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
