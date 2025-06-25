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
	// app := fiber.New()
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
			// You can customize the error response format here
			return ctx.JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Initialize database first
	fmt.Println("🔌 Initializing database connection...")
	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ Failed to connect to the database: %v", err)
	}
	fmt.Println("✅ Database initialized successfully")

	// Initialize socket service with database collections
	fmt.Println("🔧 Initializing socket service...")
	usersCollection := database.GetUsersCollection()
	sessionsCollection := database.GetSessionsCollection()
	socketService := services.NewSocketService(usersCollection, sessionsCollection)
	fmt.Println("✅ Socket service initialized")

	// Initialize Socket.IO handler with socket service
	fmt.Println("🔌 Initializing Socket.IO handler...")
	socketHandler := config.NewSocketHandler(socketService)
	fmt.Println("✅ Socket.IO handler initialized")

	// Setup Socket.IO routes (this should be before regular routes)
	socketHandler.SetupSocketRoutes(app)

	// Initialize regular routes
	routes.SetupRoutes(app)

	port := config.ServerPort
	fmt.Printf("🚀 Server starting on port :%d\n", port)
	fmt.Printf("🔌 Socket.IO server available at :%d/socket.io\n", port)
	fmt.Printf("🎮 Gameplay namespace available at :%d/socket.io/gameplay\n", port)
	fmt.Printf("📊 MongoDB URL: %s\n", config.MongoDBURL)
	fmt.Printf("🗄️  Database: %s\n", config.DatabaseName)

	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
