package services

import (
	"fmt"
	"gofiber/app/models"
	"gofiber/redis"
	"log"
	"time"

	"github.com/gocql/gocql"
)

// SessionService handles session management using Redis and Cassandra
type SessionService struct {
	cassandraSession *gocql.Session
	redisService     *redis.Service
}

// NewSessionService creates a new session service instance
func NewSessionService(cassandraSession *gocql.Session) *SessionService {
	return &SessionService{
		cassandraSession: cassandraSession,
		redisService:     redis.NewService(),
	}
}

// SessionData represents session data stored in Redis
type SessionData struct {
	SessionToken string    `json:"session_token"`
	MobileNo     string    `json:"mobile_no"`
	UserID       string    `json:"user_id"`
	DeviceID     string    `json:"device_id"`
	FCMToken     string    `json:"fcm_token"`
	JWTToken     string    `json:"jwt_token"`
	SocketID     string    `json:"socket_id"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserStatus   string    `json:"user_status"`
}

// CreateSession creates a new session and stores it in both Redis and Cassandra
func (s *SessionService) CreateSession(sessionData SessionData) error {
	// Store in Redis (primary storage)
	sessionKey := fmt.Sprintf("session:%s", sessionData.SessionToken)
	err := s.redisService.Set(sessionKey, sessionData, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to store session in Redis: %v", err)
	}

	// Store in Cassandra (backup storage)
	err = s.cassandraSession.Query(`
		INSERT INTO sessions (session_token, mobile_no, device_id, fcm_token, created_at, expires_at, is_active, jwt_token)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`).Bind(sessionData.SessionToken, sessionData.MobileNo, sessionData.DeviceID, sessionData.FCMToken,
		sessionData.CreatedAt, sessionData.ExpiresAt, sessionData.IsActive, sessionData.JWTToken).Exec()
	if err != nil {
		log.Printf("Warning: Failed to store session in Cassandra: %v", err)
	}

	// Store socket mapping in Cassandra
	if sessionData.SocketID != "" {
		err = s.cassandraSession.Query(`
			INSERT INTO sessions_by_socket (socket_id, mobile_no, user_id, session_token, created_at)
			VALUES (?, ?, ?, ?, ?)
		`).Bind(sessionData.SocketID, sessionData.MobileNo, sessionData.UserID, sessionData.SessionToken, sessionData.CreatedAt).Exec()
		if err != nil {
			log.Printf("Warning: Failed to store socket mapping in Cassandra: %v", err)
		}
	}

	log.Printf("✅ Session created: %s (Redis + Cassandra)", sessionData.SessionToken)
	return nil
}

// GetSession retrieves session data from Redis (with Cassandra fallback)
func (s *SessionService) GetSession(sessionToken string) (*SessionData, error) {
	// Try Redis first
	sessionKey := fmt.Sprintf("session:%s", sessionToken)
	var sessionData SessionData
	err := s.redisService.Get(sessionKey, &sessionData)
	if err == nil {
		// Check if session is still valid
		if time.Now().Before(sessionData.ExpiresAt) && sessionData.IsActive {
			log.Printf("📖 Session retrieved from Redis: %s", sessionToken)
			return &sessionData, nil
		} else {
			// Session expired, remove from Redis
			s.redisService.Delete(sessionKey)
			return nil, fmt.Errorf("session expired")
		}
	}

	// Fallback to Cassandra
	log.Printf("🔄 Session not found in Redis, trying Cassandra: %s", sessionToken)
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, mobile_no, device_id, fcm_token, expires_at, is_active, jwt_token, created_at
		FROM sessions
		WHERE session_token = ? AND is_active = true AND expires_at > ?
	`).Bind(sessionToken, time.Now()).Scan(&session.SessionToken, &session.MobileNo, &session.DeviceID,
		&session.FCMToken, &session.ExpiresAt, &session.IsActive, &session.JWTToken, &session.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("session not found: %v", err)
	}

	// Convert to SessionData
	sessionData = SessionData{
		SessionToken: session.SessionToken,
		MobileNo:     session.MobileNo,
		DeviceID:     session.DeviceID,
		FCMToken:     session.FCMToken,
		JWTToken:     session.JWTToken,
		IsActive:     session.IsActive,
		CreatedAt:    session.CreatedAt,
		ExpiresAt:    session.ExpiresAt,
	}

	// Store back in Redis for future access
	s.redisService.Set(sessionKey, sessionData, 24*time.Hour)

	log.Printf("📖 Session retrieved from Cassandra and cached in Redis: %s", sessionToken)
	return &sessionData, nil
}

// UpdateSession updates session data in both Redis and Cassandra
func (s *SessionService) UpdateSession(sessionToken string, updates map[string]interface{}) error {
	// Get current session data
	sessionData, err := s.GetSession(sessionToken)
	if err != nil {
		return fmt.Errorf("session not found: %v", err)
	}

	// Update fields
	if jwtToken, ok := updates["jwt_token"].(string); ok {
		sessionData.JWTToken = jwtToken
	}
	if userStatus, ok := updates["user_status"].(string); ok {
		sessionData.UserStatus = userStatus
	}
	if socketID, ok := updates["socket_id"].(string); ok {
		sessionData.SocketID = socketID
	}

	// Update in Redis
	sessionKey := fmt.Sprintf("session:%s", sessionToken)
	err = s.redisService.Set(sessionKey, sessionData, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to update session in Redis: %v", err)
	}

	// Update in Cassandra
	updateQuery := `
		UPDATE sessions
		SET jwt_token = ?
		WHERE mobile_no = ? AND device_id = ? AND created_at = ?
	`
	err = s.cassandraSession.Query(updateQuery, sessionData.JWTToken, sessionData.MobileNo,
		sessionData.DeviceID, sessionData.CreatedAt).Exec()
	if err != nil {
		log.Printf("Warning: Failed to update session in Cassandra: %v", err)
	}

	log.Printf("✅ Session updated: %s", sessionToken)
	return nil
}

// ValidateSession validates if a session is active and not expired
func (s *SessionService) ValidateSession(sessionToken, mobileNo string) bool {
	sessionData, err := s.GetSession(sessionToken)
	if err != nil {
		return false
	}

	return sessionData.MobileNo == mobileNo && sessionData.IsActive && time.Now().Before(sessionData.ExpiresAt)
}

// GetSessionBySocket retrieves session data using socket ID
func (s *SessionService) GetSessionBySocket(socketID string) (*SessionData, error) {
	// Try to get from Cassandra first (socket mapping)
	var mobileNo, userID, sessionToken string
	var createdAt time.Time

	err := s.cassandraSession.Query(`
		SELECT mobile_no, user_id, session_token, created_at
		FROM sessions_by_socket
		WHERE socket_id = ?
	`, socketID).Scan(&mobileNo, &userID, &sessionToken, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("socket session not found: %v", err)
	}

	// Get full session data
	return s.GetSession(sessionToken)
}

// DeleteSession removes session from both Redis and Cassandra
func (s *SessionService) DeleteSession(sessionToken string) error {
	// Delete from Redis
	sessionKey := fmt.Sprintf("session:%s", sessionToken)
	err := s.redisService.Delete(sessionKey)
	if err != nil {
		log.Printf("Warning: Failed to delete session from Redis: %v", err)
	}

	// Mark as inactive in Cassandra
	err = s.cassandraSession.Query(`
		UPDATE sessions
		SET is_active = false
		WHERE session_token = ?
	`).Bind(sessionToken).Exec()
	if err != nil {
		log.Printf("Warning: Failed to deactivate session in Cassandra: %v", err)
	}

	log.Printf("🗑️ Session deleted: %s", sessionToken)
	return nil
}

// UpdateSessionSocketID updates the socket ID for an existing session (for reconnections)
func (s *SessionService) UpdateSessionSocketID(sessionToken, newSocketID string) error {
	// Get current session data
	sessionData, err := s.GetSession(sessionToken)
	if err != nil {
		return fmt.Errorf("session not found: %v", err)
	}

	// Check if session is still active and not expired
	if !sessionData.IsActive || time.Now().After(sessionData.ExpiresAt) {
		return fmt.Errorf("session expired or inactive")
	}

	// Update socket ID in session data
	sessionData.SocketID = newSocketID

	// Update in Redis
	sessionKey := fmt.Sprintf("session:%s", sessionToken)
	err = s.redisService.Set(sessionKey, sessionData, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to update session in Redis: %v", err)
	}

	// Update socket mapping in Cassandra
	err = s.cassandraSession.Query(`
		INSERT INTO sessions_by_socket (socket_id, mobile_no, user_id, session_token, created_at)
		VALUES (?, ?, ?, ?, ?)
	`).Bind(newSocketID, sessionData.MobileNo, sessionData.UserID, sessionToken, sessionData.CreatedAt).Exec()
	if err != nil {
		return fmt.Errorf("failed to update socket mapping in Cassandra: %v", err)
	}

	log.Printf("✅ Session socket ID updated: %s -> %s", sessionToken, newSocketID)
	return nil
}

// CleanupExpiredSessions removes expired sessions from Redis
func (s *SessionService) CleanupExpiredSessions() error {
	// Redis automatically handles expiration
	// Just clean up Cassandra expired sessions
	err := s.cassandraSession.Query(`
		UPDATE sessions
		SET is_active = false
		WHERE expires_at < ?
	`).Bind(time.Now()).Exec()

	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %v", err)
	}

	log.Printf("🧹 Cleaned up expired sessions")
	return nil
}
