package config

import (
	"gofiber/app/models"
	"gofiber/app/services"
	"log"
	"time"

	socketio "github.com/doquangtan/socket.io/v4"
	"github.com/gofiber/fiber/v2"
)

// SocketIoHandler handles all Socket.IO related functionality
type SocketIoHandler struct {
	io                *socketio.Io
	socketService     *services.SocketService
	authHandler       *AuthSocketHandler
	gameHandler       *GameSocketHandler
	systemHandler     *SystemSocketHandler
}

// NewSocketHandler creates a new Socket.IO handler instance
func NewSocketHandler(socketService *services.SocketService) *SocketIoHandler {
	io := socketio.New()

	handler := &SocketIoHandler{
		io:                io,
		socketService:     socketService,
		authHandler:       NewAuthSocketHandler(socketService),
		gameHandler:       NewGameSocketHandler(socketService),
		systemHandler:     NewSystemSocketHandler(socketService),
	}

	handler.setupSocketHandlers()
	return handler
}

// setupSocketHandlers configures all Socket.IO event handlers
func (h *SocketIoHandler) setupSocketHandlers() {
	// Authorization handler
	h.io.OnAuthorization(func(params map[string]string) bool {
		log.Printf("Authorization attempt with params: %v", params)
		// For now, allow all connections
		// In production, you would validate tokens here
		return true
	})

	// Main connection handler
	h.io.OnConnection(func(socket *socketio.Socket) {
		log.Printf("✅ Socket connected: %s (namespace: %s)", socket.Id, socket.Nps)

		// Send welcome message
		welcome := h.socketService.HandleWelcome()
		socket.Emit("connect_response", welcome)

		// Send connect response with token
		connectResp := models.ConnectResponse{
			Token:     123456, // Generate random token
			Message:   "Welcome to the Game Admin Server!",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			SocketID:  socket.Id,
			Status:    "connected",
			Event:     "connect",
		}
		socket.Emit("connect", connectResp)

		// Setup authentication handlers
		h.authHandler.SetupAuthHandlers(socket)

		// Setup game handlers
		h.gameHandler.SetupGameHandlers(socket)

		// Setup system handlers
		h.systemHandler.SetupSystemHandlers(socket)
	})

	// Setup gameplay namespace "/gameplay"
	h.io.Of("/gameplay").OnConnection(func(socket *socketio.Socket) {
		// Send welcome message for gameplay
		socket.Emit("main:screen", map[string]interface{}{
			"success": true,
			"message": "Welcome to gameplay!",
			"server_info": map[string]interface{}{
				"version":  "1.0.0",
				"features": []string{"real-time", "multiplayer", "gameplay"},
			},
		})

		// Setup gameplay handlers
		h.gameHandler.SetupGameplayHandlers(socket)

		// Setup gameplay system handlers
		h.systemHandler.SetupGameplaySystemHandlers(socket)
	})
}

// GetIo returns the Socket.IO instance
func (h *SocketIoHandler) GetIo() *socketio.Io {
	return h.io
}

// SetupSocketRoutes configures Socket.IO routes for the Fiber app
func (h *SocketIoHandler) SetupSocketRoutes(app *fiber.App) {
	// Serve static files from WEBSITE directory if it exists
	app.Static("/", "./WEBSITE")

	// Setup Socket.IO middleware and routes
	app.Use("/", h.io.Middleware)
	app.Route("/socket.io", h.io.FiberRoute)
}
