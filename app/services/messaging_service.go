package services

import (
	"fmt"
	"time"

	socketio "github.com/doquangtan/socket.io/v4"
)

// MessagingService handles server-to-client messaging using session data
type MessagingService struct {
	sessionService *SessionService
	io             *socketio.Io
}

// NewMessagingService creates a new messaging service instance
func NewMessagingService(sessionService *SessionService, io *socketio.Io) *MessagingService {
	return &MessagingService{
		sessionService: sessionService,
		io:             io,
	}
}

// MessageData represents a message to be sent to clients
type MessageData struct {
	Event       string                 `json:"event"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   string                 `json:"timestamp"`
	Priority    string                 `json:"priority,omitempty"` // high, normal, low
	Expiry      *time.Time             `json:"expiry,omitempty"`
	TargetType  string                 `json:"target_type,omitempty"`  // user, mobile, all, specific
	TargetValue string                 `json:"target_value,omitempty"` // user_id, mobile_no, socket_id
}

// SendMessageToSocket sends a message to a specific socket connection
func (m *MessagingService) SendMessageToSocket(socketID string, message MessageData) error {
	// Find session by socket ID
	sessions, err := m.sessionService.GetAllActiveSessions()
	if err != nil {
		return fmt.Errorf("failed to get sessions: %v", err)
	}

	var targetSession *SessionData
	for _, session := range sessions {
		if session.SocketID == socketID && session.IsActive {
			targetSession = &session
			break
		}
	}

	if targetSession == nil {
		return fmt.Errorf("socket %s not found or not active", socketID)
	}

	// Update last seen timestamp
	err = m.sessionService.UpdateSessionLastSeen(targetSession.SessionToken)
	if err != nil {
		// Log error but don't fail the message sending
	}

	// Send message via Socket.IO
	if m.io != nil {
		// Get all connected sockets
		sockets := m.io.Sockets()

		// Find the specific socket and emit the message
		for _, socket := range sockets {
			if socket.Id == socketID {
				socket.Emit(message.Event, message.Data)
				return nil
			}
		}

		// If socket not found in connected sockets, try broadcasting
		// This handles cases where the socket might be in a different namespace
		m.io.Emit(message.Event, message.Data)
		return nil
	}

	return fmt.Errorf("socket not found or not connected")
}

// SendMessageToUser sends a message to all connections of a specific user
func (m *MessagingService) SendMessageToUser(userID string, message MessageData) error {
	sessions, err := m.sessionService.GetSessionsByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %v", err)
	}

	if len(sessions) == 0 {
		return fmt.Errorf("no active sessions found for user %s", userID)
	}

	var sentCount int
	for _, session := range sessions {
		if session.SocketID != "" {
			err := m.SendMessageToSocket(session.SocketID, message)
			if err == nil {
				sentCount++
			} else {
			}
		}
	}

	return nil
}

// SendMessageToMobile sends a message to all connections of a specific mobile number
func (m *MessagingService) SendMessageToMobile(mobileNo string, message MessageData) error {
	sessions, err := m.sessionService.GetSessionsByMobileNo(mobileNo)
	if err != nil {
		return fmt.Errorf("failed to get mobile sessions: %v", err)
	}

	if len(sessions) == 0 {
		return fmt.Errorf("no active sessions found for mobile %s", mobileNo)
	}

	var sentCount int
	for _, session := range sessions {
		if session.SocketID != "" {
			err := m.SendMessageToSocket(session.SocketID, message)
			if err == nil {
				sentCount++
			} else {
			}
		}
	}

	return nil
}

// BroadcastMessage sends a message to all active connections
func (m *MessagingService) BroadcastMessage(message MessageData) error {
	sessions, err := m.sessionService.GetAllActiveSessions()
	if err != nil {
		return fmt.Errorf("failed to get active sessions: %v", err)
	}

	if len(sessions) == 0 {
		return fmt.Errorf("no active sessions found")
	}

	var sentCount int
	for _, session := range sessions {
		if session.SocketID != "" {
			err := m.SendMessageToSocket(session.SocketID, message)
			if err == nil {
				sentCount++
			} else {
			}
		}
	}

	return nil
}

// SendNotification sends a notification message to a specific target
func (m *MessagingService) SendNotification(targetType, targetValue, title, body string, data map[string]interface{}) error {
	message := MessageData{
		Event:     "notification",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Priority:  "normal",
		Data: map[string]interface{}{
			"title": title,
			"body":  body,
			"data":  data,
		},
	}

	switch targetType {
	case "user":
		return m.SendMessageToUser(targetValue, message)
	case "mobile":
		return m.SendMessageToMobile(targetValue, message)
	case "socket":
		return m.SendMessageToSocket(targetValue, message)
	case "all":
		return m.BroadcastMessage(message)
	default:
		return fmt.Errorf("invalid target type: %s", targetType)
	}
}

// SendGameUpdate sends a game update message to a specific user
func (m *MessagingService) SendGameUpdate(userID string, gameData map[string]interface{}) error {
	message := MessageData{
		Event:     "game:update",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Priority:  "high",
		Data:      gameData,
	}

	return m.SendMessageToUser(userID, message)
}

// SendContestUpdate sends a contest update message to a specific user
func (m *MessagingService) SendContestUpdate(userID string, contestData map[string]interface{}) error {
	message := MessageData{
		Event:     "contest:update",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Priority:  "high",
		Data:      contestData,
	}

	return m.SendMessageToUser(userID, message)
}

// SendSystemAlert sends a system alert to all users
func (m *MessagingService) SendSystemAlert(alertType, message string, severity string) error {
	alertData := map[string]interface{}{
		"type":     alertType,
		"message":  message,
		"severity": severity,
	}

	alertMessage := MessageData{
		Event:     "system:alert",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Priority:  "high",
		Data:      alertData,
	}

	return m.BroadcastMessage(alertMessage)
}

// GetActiveConnectionsCount returns the number of active connections
func (m *MessagingService) GetActiveConnectionsCount() (int64, error) {
	return m.sessionService.GetActiveSessionsCount()
}

// GetUserConnections returns all active connections for a user
func (m *MessagingService) GetUserConnections(userID string) ([]SessionData, error) {
	return m.sessionService.GetSessionsByUserID(userID)
}

// GetMobileConnections returns all active connections for a mobile number
func (m *MessagingService) GetMobileConnections(mobileNo string) ([]SessionData, error) {
	return m.sessionService.GetSessionsByMobileNo(mobileNo)
}

// CleanupInactiveConnections removes inactive connections
func (m *MessagingService) CleanupInactiveConnections(maxIdleTime time.Duration) error {
	return m.sessionService.CleanupInactiveSessions(maxIdleTime)
}

// StoreConnectionData stores connection data when a user connects
func (m *MessagingService) StoreConnectionData(socketID, userID, mobileNo, sessionToken, deviceID, fcmToken, userAgent, ipAddress, namespace string) error {
	// This is now handled by the session service
	// Connection data is stored as part of the session data
	return nil
}

// RemoveConnectionData removes connection data when a user disconnects
func (m *MessagingService) RemoveConnectionData(socketID string) error {
	// Find session by socket ID and mark as inactive
	sessions, err := m.sessionService.GetAllActiveSessions()
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if session.SocketID == socketID {
			// Mark session as inactive using session service
			updates := map[string]interface{}{
				"is_active": false,
			}
			return m.sessionService.UpdateSession(session.SessionToken, updates)
		}
	}

	return nil
}

// UpdateConnectionData updates connection data (e.g., last seen)
func (m *MessagingService) UpdateConnectionData(socketID string) error {
	// Find session by socket ID and update last seen
	sessions, err := m.sessionService.GetAllActiveSessions()
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if session.SocketID == socketID {
			return m.sessionService.UpdateSessionLastSeen(session.SessionToken)
		}
	}

	return fmt.Errorf("session not found for socket %s", socketID)
}

// GetSessionService returns the session service instance
func (m *MessagingService) GetSessionService() *SessionService {
	return m.sessionService
}
