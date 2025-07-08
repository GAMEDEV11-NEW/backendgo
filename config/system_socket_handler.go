package config

import (
	"gofiber/app/models"
	"gofiber/app/services"
	"log"
	"time"

	socketio "github.com/doquangtan/socket.io/v4"
)

// SystemSocketHandler handles all system-related Socket.IO events
type SystemSocketHandler struct {
	socketService *services.SocketService
}

// NewSystemSocketHandler creates a new system socket handler instance
func NewSystemSocketHandler(socketService *services.SocketService) *SystemSocketHandler {
	return &SystemSocketHandler{
		socketService: socketService,
	}
}

// SetupSystemHandlers configures all system-related Socket.IO event handlers
func (h *SystemSocketHandler) SetupSystemHandlers(socket *socketio.Socket, authFunc func(socket *socketio.Socket, eventName string) (*models.User, error)) {
	// Heartbeat handler
	socket.On("heartbeat", func(event *socketio.EventPayload) {
		log.Printf("💓 Heartbeat received from %s", socket.Id)

		// Authenticate user
		_, err := authFunc(socket, "heartbeat")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

		response := h.socketService.HandleHeartbeat()
		socket.Emit("heartbeat", response)
	})

	// Health check handler
	socket.On("health_check", func(event *socketio.EventPayload) {
		log.Printf("🏥 Health check received from %s", socket.Id)

		// Authenticate user
		_, err := authFunc(socket, "health_check")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

		response := h.socketService.HandleHealthCheck()
		socket.Emit("health_check:ack", response)
	})

	// Ping handler
	socket.On("ping", func(event *socketio.EventPayload) {
		log.Printf("🏓 Ping received from %s", socket.Id)

		// Authenticate user
		_, err := authFunc(socket, "ping")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

		socket.Emit("pong", map[string]interface{}{
			"success":   true,
			"message":   "pong",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	// Disconnect handlers
	socket.On("disconnecting", func(event *socketio.EventPayload) {
		log.Printf("🔌 Socket disconnecting: %s (namespace: %s)", socket.Id, socket.Nps)
	})

	socket.On("disconnect", func(event *socketio.EventPayload) {

		// Handle the disconnect by cleaning up sessions and updating league_joins
		if err := h.socketService.HandleSocketDisconnect(socket.Id); err != nil {

		} else {
			// Broadcast disconnect notification to other clients
			disconnectMessage := map[string]interface{}{
				"event":     "user:disconnected",
				"socket_id": socket.Id,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"message":   "A user has disconnected",
			}
			socket.Emit("user:disconnected", disconnectMessage)
		}

	})
}

// SetupGameplaySystemHandlers configures system handlers for gameplay namespace
func (h *SystemSocketHandler) SetupGameplaySystemHandlers(socket *socketio.Socket) {
	// Gameplay disconnect handler
	socket.On("disconnect", func(event *socketio.EventPayload) {
		// Handle the disconnect by cleaning up sessions and updating league_joins
		if err := h.socketService.HandleSocketDisconnect(socket.Id); err != nil {
		} else {
		}
	})
}
