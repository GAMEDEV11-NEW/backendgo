// app/routes/routes.go
package routes

import (
	"gofiber/app/controllers"
	"gofiber/app/services"
	"gofiber/database"
	"gofiber/redis"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, messagingService *services.MessagingService) {
	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		health := map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"services":  map[string]string{},
		}

		// Check Cassandra connection
		if err := database.HealthCheck(); err != nil {
			health["services"].(map[string]string)["cassandra"] = "error: " + err.Error()
		} else {
			health["services"].(map[string]string)["cassandra"] = "ok"
		}

		// Check Redis connection
		redisService := redis.NewService()
		if _, err := redisService.GetClient().Ping(redisService.GetContext()).Result(); err != nil {
			health["services"].(map[string]string)["redis"] = "error: " + err.Error()
		} else {
			health["services"].(map[string]string)["redis"] = "ok"
		}

		return c.JSON(health)
	})

	// Matchmaking status endpoint
	app.Get("/api/matchmaking/status", func(c *fiber.Ctx) error {
		cronService := services.NewCronService(database.CassandraSession)

		stats, err := cronService.GetMatchmakingStats()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":     "Failed to get matchmaking stats",
				"message":   err.Error(),
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		}

		return c.JSON(fiber.Map{
			"status":       "success",
			"cron_running": cronService.IsRunning(),
			"stats":        stats,
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API version endpoint
	app.Get("/api/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version":   "1.0.0",
			"name":      "GOSOCKET",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Setup messaging routes
	setupMessagingRoutes(app, messagingService)
}

// setupMessagingRoutes sets up all messaging-related endpoints
func setupMessagingRoutes(app *fiber.App, messagingService *services.MessagingService) {
	// Initialize messaging controller
	messagingController := controllers.NewMessagingController(messagingService)

	// Messaging API routes
	messagingGroup := app.Group("/api/messaging")

	// Send messages
	messagingGroup.Post("/send/notification", messagingController.SendNotification)
	messagingGroup.Post("/send/game-update", messagingController.SendGameUpdate)
	messagingGroup.Post("/send/contest-update", messagingController.SendContestUpdate)
	messagingGroup.Post("/send/system-alert", messagingController.SendSystemAlert)

	// Get connection information
	messagingGroup.Get("/connections/stats", messagingController.GetConnectionStats)
	messagingGroup.Get("/connections/user/:user_id", messagingController.GetUserConnections)
	messagingGroup.Get("/connections/mobile/:mobile_no", messagingController.GetMobileConnections)

	// Session management (single session enforcement)
	messagingGroup.Get("/session/status/:mobile_no", messagingController.GetUserSessionStatus)
	messagingGroup.Post("/session/logout/:mobile_no", messagingController.ForceLogoutUser)
}
