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
	
	app := fiber.New(fiber.Config{
		Prefork:       false,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Fiber",
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
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

	// Initialize database first
	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ Failed to connect to the database: %v", err)
	}

	// Initialize socket service with Cassandra session
	socketService := services.NewSocketService(database.CassandraSession)
	
	socketHandler := config.NewSocketHandler(socketService)
	
	// Setup Socket.IO routes (this should be before regular routes)
	socketHandler.SetupSocketRoutes(app)
	
	// Initialize regular routes
	routes.SetupRoutes(app)
	
	// Start background cleanup goroutine
	go func() {
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
	
	port := config.ServerPort
	log.Printf("🚀 DEBUG: Server starting on port :%d", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
