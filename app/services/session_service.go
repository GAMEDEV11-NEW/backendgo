package services

import (
	"fmt"
	"gofiber/app/models"
	"gofiber/redis"
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
	// Connection data integrated into session
	ConnectedAt time.Time `json:"connected_at,omitempty"`
	LastSeen    time.Time `json:"last_seen,omitempty"`
	UserAgent   string    `json:"user_agent,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	Namespace   string    `json:"namespace,omitempty"`
}

// DeactivateExistingSessions deactivates all existing sessions for a user (mobile number)
func (s *SessionService) DeactivateExistingSessions(mobileNo string) error {
	// Get all active sessions for this mobile number
	sessions, err := s.GetSessionsByMobileNo(mobileNo)
	if err != nil {
		return fmt.Errorf("failed to get existing sessions: %v", err)
	}

	var deactivatedCount int
	for _, session := range sessions {
		if session.IsActive {
			// Deactivate session in Redis
			sessionKey := fmt.Sprintf("session:%s", session.SessionToken)
			session.IsActive = false
			err := s.redisService.Set(sessionKey, session, 24*time.Hour)
			if err != nil {
			} else {
				deactivatedCount++
			}

			// Deactivate session in Cassandra
			err = s.cassandraSession.Query(`
				UPDATE sessions
				SET is_active = false
				WHERE mobile_no = ? AND device_id = ? AND created_at = ?
			`).Bind(session.MobileNo, session.DeviceID, session.CreatedAt).Exec()
			if err != nil {
			}
		}
	}

	if deactivatedCount > 0 {
	}

	return nil
}

// CreateSession creates a new session and stores it in both Redis and Cassandra
// Enforces single session per user by deactivating existing sessions
func (s *SessionService) CreateSession(sessionData SessionData) error {
	// Deactivate any existing sessions for this user
	err := s.DeactivateExistingSessions(sessionData.MobileNo)
	if err != nil {
	}

	// Set connection timestamps if not set
	if sessionData.ConnectedAt.IsZero() {
		sessionData.ConnectedAt = time.Now()
	}
	if sessionData.LastSeen.IsZero() {
		sessionData.LastSeen = time.Now()
	}

	// Store in Redis (primary storage)
	sessionKey := fmt.Sprintf("session:%s", sessionData.SessionToken)
	err = s.redisService.Set(sessionKey, sessionData, 24*time.Hour)
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
	}

	// Store socket mapping in Cassandra
	if sessionData.SocketID != "" {
		err = s.cassandraSession.Query(`
			INSERT INTO sessions_by_socket (socket_id, mobile_no, user_id, session_token, created_at)
			VALUES (?, ?, ?, ?, ?)
		`).Bind(sessionData.SocketID, sessionData.MobileNo, sessionData.UserID, sessionData.SessionToken, sessionData.CreatedAt).Exec()
		if err != nil {
		}
	}

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
			return &sessionData, nil
		} else {
			// Session expired, remove from Redis
			s.redisService.Delete(sessionKey)
			return nil, fmt.Errorf("session expired")
		}
	}

	// Fallback to Cassandra
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
	if userAgent, ok := updates["user_agent"].(string); ok {
		sessionData.UserAgent = userAgent
	}
	if ipAddress, ok := updates["ip_address"].(string); ok {
		sessionData.IPAddress = ipAddress
	}
	if namespace, ok := updates["namespace"].(string); ok {
		sessionData.Namespace = namespace
	}

	// Always update last seen timestamp
	sessionData.LastSeen = time.Now()

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
	}

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
	}

	// Mark as inactive in Cassandra
	err = s.cassandraSession.Query(`
		UPDATE sessions
		SET is_active = false
		WHERE session_token = ?
	`).Bind(sessionToken).Exec()
	if err != nil {
	}

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

	// Update session with new socket ID and connection data
	updates := map[string]interface{}{
		"socket_id": newSocketID,
		"last_seen": time.Now(),
	}

	err = s.UpdateSession(sessionToken, updates)
	if err != nil {
		return fmt.Errorf("failed to update session: %v", err)
	}

	// Update socket mapping in Cassandra
	err = s.cassandraSession.Query(`
		INSERT INTO sessions_by_socket (socket_id, mobile_no, user_id, session_token, created_at)
		VALUES (?, ?, ?, ?, ?)
	`).Bind(newSocketID, sessionData.MobileNo, sessionData.UserID, sessionToken, sessionData.CreatedAt).Exec()
	if err != nil {
		return fmt.Errorf("failed to update socket mapping in Cassandra: %v", err)
	}

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

	return nil
}

// GetSessionsByUserID retrieves all active sessions for a specific user
func (s *SessionService) GetSessionsByUserID(userID string) ([]SessionData, error) {
	// Get all session keys from Redis
	pattern := "session:*"
	keys, err := s.redisService.GetClient().Keys(s.redisService.GetContext(), pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session keys: %v", err)
	}

	var sessions []SessionData
	for _, key := range keys {
		var sessionData SessionData
		err := s.redisService.Get(key, &sessionData)
		if err == nil && sessionData.UserID == userID && sessionData.IsActive && time.Now().Before(sessionData.ExpiresAt) {
			sessions = append(sessions, sessionData)
		}
	}

	return sessions, nil
}

// GetSessionsByMobileNo retrieves all active sessions for a specific mobile number
func (s *SessionService) GetSessionsByMobileNo(mobileNo string) ([]SessionData, error) {
	// Get all session keys from Redis
	pattern := "session:*"
	keys, err := s.redisService.GetClient().Keys(s.redisService.GetContext(), pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session keys: %v", err)
	}

	var sessions []SessionData
	for _, key := range keys {
		var sessionData SessionData
		err := s.redisService.Get(key, &sessionData)
		if err == nil && sessionData.MobileNo == mobileNo && sessionData.IsActive && time.Now().Before(sessionData.ExpiresAt) {
			sessions = append(sessions, sessionData)
		}
	}

	return sessions, nil
}

// GetAllActiveSessions retrieves all active sessions
func (s *SessionService) GetAllActiveSessions() ([]SessionData, error) {
	// Get all session keys from Redis
	pattern := "session:*"
	keys, err := s.redisService.GetClient().Keys(s.redisService.GetContext(), pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session keys: %v", err)
	}

	var sessions []SessionData
	for _, key := range keys {
		var sessionData SessionData
		err := s.redisService.Get(key, &sessionData)
		if err == nil && sessionData.IsActive && time.Now().Before(sessionData.ExpiresAt) {
			sessions = append(sessions, sessionData)
		}
	}

	return sessions, nil
}

// GetActiveSessionsCount returns the total number of active sessions
func (s *SessionService) GetActiveSessionsCount() (int64, error) {
	sessions, err := s.GetAllActiveSessions()
	if err != nil {
		return 0, err
	}
	return int64(len(sessions)), nil
}

// UpdateSessionLastSeen updates the last seen timestamp for a session
func (s *SessionService) UpdateSessionLastSeen(sessionToken string) error {
	sessionData, err := s.GetSession(sessionToken)
	if err != nil {
		return err
	}

	sessionData.LastSeen = time.Now()

	// Update in Redis
	sessionKey := fmt.Sprintf("session:%s", sessionToken)
	err = s.redisService.Set(sessionKey, sessionData, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to update session last seen: %v", err)
	}

	return nil
}

// CleanupInactiveSessions removes sessions that haven't been seen recently
func (s *SessionService) CleanupInactiveSessions(maxIdleTime time.Duration) error {
	sessions, err := s.GetAllActiveSessions()
	if err != nil {
		return err
	}

	cutoffTime := time.Now().Add(-maxIdleTime)
	var cleanedCount int64

	for _, sessionData := range sessions {
		if sessionData.LastSeen.Before(cutoffTime) {
			// Mark session as inactive
			sessionData.IsActive = false
			sessionKey := fmt.Sprintf("session:%s", sessionData.SessionToken)
			s.redisService.Set(sessionKey, sessionData, 24*time.Hour)
			cleanedCount++
		}
	}

	return nil
}

// GetActiveSessionForUser returns the current active session for a user (mobile number)
func (s *SessionService) GetActiveSessionForUser(mobileNo string) (*SessionData, error) {
	sessions, err := s.GetSessionsByMobileNo(mobileNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions for user: %v", err)
	}

	// Find the active session
	for _, session := range sessions {
		if session.IsActive && time.Now().Before(session.ExpiresAt) {
			return &session, nil
		}
	}

	return nil, fmt.Errorf("no active session found for mobile %s", mobileNo)
}

// HasActiveSession checks if a user has an active session
func (s *SessionService) HasActiveSession(mobileNo string) bool {
	_, err := s.GetActiveSessionForUser(mobileNo)
	return err == nil
}

// GetActiveSessionCount returns the number of active sessions for a user
func (s *SessionService) GetActiveSessionCount(mobileNo string) (int, error) {
	sessions, err := s.GetSessionsByMobileNo(mobileNo)
	if err != nil {
		return 0, fmt.Errorf("failed to get sessions for user: %v", err)
	}

	count := 0
	for _, session := range sessions {
		if session.IsActive && time.Now().Before(session.ExpiresAt) {
			count++
		}
	}

	return count, nil
}
