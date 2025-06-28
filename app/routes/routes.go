// app/routes/routes.go
package routes

import (
	"gofiber/database"
	"gofiber/redis"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
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

	// API version endpoint
	app.Get("/api/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version":   "1.0.0",
			"name":      "GOSOCKET",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})
}
