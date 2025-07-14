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

	// Initialize socket handler
	socketHandler := config.NewSocketHandler(socketService)

	// Initialize messaging service with Session Service and Socket.IO
	messagingService := services.NewMessagingService(socketService.GetSessionService(), socketHandler.GetIo())
	socketHandler.SetMessagingService(messagingService)

	// Initialize cron service for matchmaking
	cronService := services.NewCronService(database.CassandraSession)

	// Start matchmaking cron job (runs every 30 seconds)
	cronService.StartMatchmakingCron(3 * time.Second)

	// Start cleanup cron job (runs every 5 minutes, cleans matches older than 24 hours)
	cronService.RunCleanupCron(5*time.Minute, 24*time.Hour)

	// Setup Socket.IO routes (this should be before regular routes)
	socketHandler.SetupSocketRoutes(app)

	// Initialize regular routes
	routes.SetupRoutes(app, messagingService)

	port := config.ServerPort
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
