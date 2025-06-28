package config

import (
	"gofiber/app/services"
	"log"
	

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
func (h *SystemSocketHandler) SetupSystemHandlers(socket *socketio.Socket) {
	// Heartbeat handler
	socket.On("heartbeat", func(event *socketio.EventPayload) {
		log.Printf("💓 Heartbeat received from %s", socket.Id)
		response := h.socketService.HandleHeartbeat()
		socket.Emit("heartbeat", response)
	})

	// Health check handler
	socket.On("health_check", func(event *socketio.EventPayload) {
		log.Printf("🏥 Health check received from %s", socket.Id)
		response := h.socketService.HandleHealthCheck()
		socket.Emit("health_check:ack", response)
	})

	// Ping handler
	socket.On("ping", func(event *socketio.EventPayload) {
		log.Printf("🏓 Ping received from %s", socket.Id)
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
		log.Printf("🔌 Socket disconnected: %s (namespace: %s)", socket.Id, socket.Nps)
	})
}

// SetupGameplaySystemHandlers configures system handlers for gameplay namespace
func (h *SystemSocketHandler) SetupGameplaySystemHandlers(socket *socketio.Socket) {
	// Gameplay disconnect handler
	socket.On("disconnect", func(event *socketio.EventPayload) {
		log.Printf("🎮 Gameplay socket disconnected: %s", socket.Id)
	})
} 