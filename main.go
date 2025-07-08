// main.go
package main

import (
	"fmt"
	"gofiber/app/routes"
	"gofiber/app/services"
	"gofiber/config"
	"gofiber/database"
	"log"

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

	port := config.ServerPort
	log.Printf("🚀 DEBUG: Server starting on port :%d", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
