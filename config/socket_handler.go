package config

import (
	"fmt"
	"gofiber/app/models"
	"gofiber/app/services"
	"time"

	socketio "github.com/doquangtan/socket.io/v4"
	"github.com/gofiber/fiber/v2"
)

// AuthenticationError represents authentication errors
type AuthenticationError struct {
	*models.ConnectionError
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// SocketIoHandler handles all Socket.IO related functionality
type SocketIoHandler struct {
	io               *socketio.Io
	socketService    *services.SocketService
	messagingService *services.MessagingService
	authHandler      *AuthSocketHandler
	gameHandler      *GameSocketHandler
	gameboardHandler *GameboardSocketHandler
	systemHandler    *SystemSocketHandler
}

// NewSocketHandler creates a new Socket.IO handler instance
func NewSocketHandler(socketService *services.SocketService) *SocketIoHandler {
	io := socketio.New()

	handler := &SocketIoHandler{
		io:               io,
		socketService:    socketService,
		authHandler:      NewAuthSocketHandler(socketService),
		gameHandler:      NewGameSocketHandler(socketService),
		gameboardHandler: NewGameboardSocketHandler(socketService),
		systemHandler:    NewSystemSocketHandler(socketService),
	}

	// Set the Socket.IO instance in the socket service
	socketService.SetIo(io)

	handler.setupSocketHandlers()
	return handler
}

// SetMessagingService sets the messaging service for the socket handler
func (h *SocketIoHandler) SetMessagingService(messagingService *services.MessagingService) {
	h.messagingService = messagingService
	h.socketService.SetMessagingService(messagingService)
}

// authenticateUser validates user authentication for all events
func (h *SocketIoHandler) authenticateUser(socket *socketio.Socket, eventName string) (*models.User, error) {
	// Skip authentication for auth-related events
	authEvents := map[string]bool{
		"device:info":      true,
		"login":            true,
		"verify:otp":       true,
		"restore:session":  true,
		"logout":           true,
		"connect":          true,
		"disconnect":       true,
		"connect_response": true,
	}

	if authEvents[eventName] {
		return nil, nil // Skip authentication for auth events
	}

	// Get session data using SessionService (Redis + Cassandra)
	sessionData, err := h.socketService.GetSessionService().GetSessionBySocket(socket.Id)
	if err != nil {
		return nil, &AuthenticationError{
			ConnectionError: &models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidSession,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "session",
				Message:   "User not authenticated. Please login first.",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "authentication_error",
			},
		}
	}

	// Get user details
	var user models.User
	err = h.socketService.GetCassandraSession().Query(`
		SELECT id, mobile_no, full_name, status, language_code
		FROM users
		WHERE mobile_no = ?
		ALLOW FILTERING
	`).Bind(sessionData.MobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)

	if err != nil {
		return nil, &AuthenticationError{
			ConnectionError: &models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidSession,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "user",
				Message:   "User not found in database",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "authentication_error",
			},
		}
	}

	// Session is already validated by SessionService, just verify it's still active
	if !sessionData.IsActive || time.Now().After(sessionData.ExpiresAt) {
		return nil, &AuthenticationError{
			ConnectionError: &models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidSession,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "session",
				Message:   "Session expired or invalid",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "authentication_error",
			},
		}
	}

	return &user, nil
}

// setupSocketHandlers configures all Socket.IO event handlers
func (h *SocketIoHandler) setupSocketHandlers() {
	// Authorization handler
	h.io.OnAuthorization(func(params map[string]string) bool {
		// For now, allow all connections
		// In production, you would validate tokens here
		return true
	})

	// Main connection handler
	h.io.OnConnection(func(socket *socketio.Socket) {
		// Store basic connection data (will be updated with user data after authentication)
		if h.messagingService != nil {
			// Store basic connection info - will be updated when user authenticates
		}

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

		// Setup authentication handlers with authentication middleware
		h.authHandler.SetupAuthHandlers(socket, h.authenticateUser)

		// Setup game handlers with authentication middleware
		h.gameHandler.SetupGameHandlers(socket, h.authenticateUser)

		// Setup gameboard handlers with authentication middleware
		h.gameboardHandler.SetupGameboardHandlers(socket, h.authenticateUser)

		// Setup system handlers with authentication middleware
		h.systemHandler.SetupSystemHandlers(socket, h.authenticateUser)
	})

}

// GetIo returns the Socket.IO instance
func (h *SocketIoHandler) GetIo() *socketio.Io {
	return h.io
}

// SendMessageToSocket sends a message to a specific socket
func (h *SocketIoHandler) SendMessageToSocket(socketID string, event string, data interface{}) error {
	// Get all connected sockets
	sockets := h.io.Sockets()

	// Find the specific socket
	for _, socket := range sockets {
		if socket.Id == socketID {
			socket.Emit(event, data)
			return nil
		}
	}
	return fmt.Errorf("socket %s not found", socketID)
}

// SetupSocketRoutes configures Socket.IO routes for the Fiber app
func (h *SocketIoHandler) SetupSocketRoutes(app *fiber.App) {
	// Serve static files from WEBSITE directory if it exists
	app.Static("/", "./WEBSITE")

	// Setup Socket.IO middleware and routes
	app.Use("/", h.io.Middleware)
	app.Route("/socket.io", h.io.FiberRoute)
}
