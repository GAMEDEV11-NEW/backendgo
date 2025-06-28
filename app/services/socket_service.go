package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"gofiber/app/models"
	"gofiber/app/utils"
	"gofiber/redis"
	"log"
	"math"
	"math/big"
	"sort"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SocketService handles all socket-related business logic
// Refactored to use Cassandra
type SocketService struct {
	cassandraSession *gocql.Session
	redisService     *redis.Service
}

// NewSocketService creates a new socket service instance using Cassandra
func NewSocketService(cassandraSession *gocql.Session) *SocketService {
	redisService := redis.NewService()
	return &SocketService{
		cassandraSession: cassandraSession,
		redisService:     redisService,
	}
}

// GenerateSessionToken generates a unique session token
func (s *SocketService) GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateOTP generates a 6-digit OTP
func (s *SocketService) GenerateOTP() int {
	// Generate a random number between 100000 and 999999
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return int(n.Int64()) + 100000
}

// HandleDeviceInfo processes device information from client
func (s *SocketService) HandleDeviceInfo(deviceInfo models.DeviceInfo, socketID string) models.DeviceInfoResponse {
	log.Printf("📱 Device info received: %+v", deviceInfo)

	return models.DeviceInfoResponse{
		Status:    "success",
		Message:   "Device info received and validated",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SocketID:  socketID,
		Event:     "device:info:ack",
	}
}

// HandleLogin processes login request and generates OTP
func (s *SocketService) HandleLogin(loginReq models.LoginRequest) (*models.LoginResponse, error) {
	log.Printf("Login request received for mobile: %s", loginReq.MobileNo)

	// Validate mobile number (basic validation)
	if len(loginReq.MobileNo) < 10 {
		return nil, fmt.Errorf("invalid mobile number")
	}

	// Validate FCM token length (should be at least 100 characters as per test)
	if len(loginReq.FCMToken) < 100 {
		return nil, fmt.Errorf("FCM token too short")
	}

	// Generate session token
	sessionToken, err := s.GenerateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %v", err)
	}

	// Generate OTP
	otp := s.GenerateOTP()

	// Check if user exists
	var existingUser models.User
	err = s.cassandraSession.Query(`
		SELECT mobile_no, email, status, created_at, updated_at
		FROM users
		WHERE mobile_no = ?
	`).Bind(loginReq.MobileNo).Scan(&existingUser.MobileNo, &existingUser.Email, &existingUser.Status, &existingUser.CreatedAt, &existingUser.UpdatedAt)

	var userID string
	var isNewUser bool
	if err == gocql.ErrNotFound {
		// User doesn't exist, create new user
		userID = primitive.NewObjectID().Hex()
		user := models.User{
			ID:        userID,
			MobileNo:  loginReq.MobileNo,
			Email:     loginReq.Email,
			Status:    "new_user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = s.cassandraSession.Query(`
			INSERT INTO users (id, mobile_no, email, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`).Bind(userID, user.MobileNo, user.Email, user.Status, user.CreatedAt, user.UpdatedAt).Exec()
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %v", err)
		}
		log.Printf("New user created: %s", loginReq.MobileNo)
		isNewUser = true
	} else if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	} else {
		log.Printf("Existing user found: %s", loginReq.MobileNo)
		userID = existingUser.ID
		isNewUser = false
	}

	log.Printf("isNewUser: %v", isNewUser)

	// Check if there's an existing active session for this mobile number and device
	var existingSession models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, fcm_token, expires_at, is_active
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
	`).Bind(loginReq.MobileNo, loginReq.DeviceID, time.Now()).Scan(&existingSession.SessionToken, &existingSession.FCMToken, &existingSession.ExpiresAt, &existingSession.IsActive)

	if err == nil {
		// Existing session found, update it with new session token
		log.Printf("Updating existing session for mobile: %s, device: %s", loginReq.MobileNo, loginReq.DeviceID)
		err = s.cassandraSession.Query(`
			UPDATE sessions
			SET session_token = ?, fcm_token = ?, expires_at = ?, is_active = ?
			WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
		`).Bind(sessionToken, loginReq.FCMToken, time.Now().Add(24*time.Hour), true, loginReq.MobileNo, loginReq.DeviceID, time.Now()).Exec()
		if err != nil {
			return nil, fmt.Errorf("failed to update existing session: %v", err)
		}
		log.Printf("Existing session updated for %s", loginReq.MobileNo)
	} else if err == gocql.ErrNotFound {
		// No existing session, create new one (without JWT token yet)
		log.Printf("Creating new session for mobile: %s, device: %s", loginReq.MobileNo, loginReq.DeviceID)
		session := models.Session{
			ID:           primitive.NewObjectID().Hex(),
			UserID:       userID,
			SessionToken: sessionToken,
			JWTToken:     "", // Will be set after OTP verification
			MobileNo:     loginReq.MobileNo,
			DeviceID:     loginReq.DeviceID,
			FCMToken:     loginReq.FCMToken,
			CreatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour), // 24 hour expiry
			IsActive:     true,
		}
		err = s.cassandraSession.Query(`
			INSERT INTO sessions (id, user_id, session_token, jwt_token, mobile_no, device_id, fcm_token, created_at, expires_at, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`).Bind(session.ID, session.UserID, session.SessionToken, session.JWTToken, session.MobileNo, session.DeviceID, session.FCMToken, session.CreatedAt, session.ExpiresAt, session.IsActive).Exec()
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %v", err)
		}
		log.Printf("New session created for %s", loginReq.MobileNo)
	} else {
		// Database error
		return nil, fmt.Errorf("database error checking existing session: %v", err)
	}

	log.Printf("OTP generated for %s: %d", loginReq.MobileNo, otp)

	return &models.LoginResponse{
		Status:       "success",
		Message:      "OTP sent successfully",
		MobileNo:     loginReq.MobileNo,
		DeviceID:     loginReq.DeviceID,
		SessionToken: sessionToken,
		OTP:          otp,
		IsNewUser:    isNewUser,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		SocketID:     "",
		Event:        "otp:sent",
	}, nil
}

// HandleOTPVerification verifies OTP and returns user status
func (s *SocketService) HandleOTPVerification(otpReq models.OTPVerificationRequest) (*models.OTPVerificationResponse, error) {
	log.Printf("OTP verification for mobile: %s", otpReq.MobileNo)

	// Validate session
	var session models.Session
	err := s.cassandraSession.Query(`
		SELECT session_token, fcm_token, expires_at, is_active
		FROM sessions
		WHERE session_token = ? AND mobile_no = ? AND is_active = true AND expires_at > ?
	`).Bind(otpReq.SessionToken, otpReq.MobileNo, time.Now()).Scan(&session.SessionToken, &session.FCMToken, &session.ExpiresAt, &session.IsActive)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// In a real application, you would verify the OTP against what was sent
	// For this demo, we'll accept any 6-digit OTP
	otpInt, err := strconv.Atoi(otpReq.OTP)
	if err != nil || otpInt < 100000 || otpInt > 999999 {
		return nil, fmt.Errorf("invalid OTP format")
	}

	// Check user status
	var user models.User
	err = s.cassandraSession.Query(`
		SELECT status
		FROM users
		WHERE mobile_no = ?
	`).Bind(otpReq.MobileNo).Scan(&user.Status)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Determine user status and update if needed
	userStatus := user.Status
	if user.Status == "new_user" {
		// Update user status to existing_user
		err = s.cassandraSession.Query(`
			UPDATE users
			SET status = ?
			WHERE mobile_no = ?
		`).Bind("existing_user", otpReq.MobileNo).Exec()
		if err != nil {
			return nil, fmt.Errorf("failed to update user status: %v", err)
		}
	} else {
		userStatus = "existing_user"
	}

	// Generate simple JWT token with only mobile_no, device_id, and fcm_token
	jwtToken, err := utils.GenerateSimpleJWTToken(otpReq.MobileNo, session.DeviceID, session.FCMToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT token: %v", err)
	}

	// Update session with JWT token
	err = s.cassandraSession.Query(`
		UPDATE sessions
		SET jwt_token = ?
		WHERE session_token = ?
	`).Bind(jwtToken, otpReq.SessionToken).Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to update session with JWT token: %v", err)
	}

	log.Printf("✅ JWT token generated and stored for user: %s", otpReq.MobileNo)

	return &models.OTPVerificationResponse{
		Status:       "success",
		Message:      "OTP verified successfully",
		MobileNo:     otpReq.MobileNo,
		DeviceID:     session.DeviceID,
		SessionToken: otpReq.SessionToken,
		JWTToken:     jwtToken,
		UserStatus:   userStatus,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		SocketID:     "",
		Event:        "otp:verified",
	}, nil
}

// HandleSetProfile sets up user profile
func (s *SocketService) HandleSetProfile(profileReq models.SetProfileRequest) (*models.SetProfileResponse, error) {
	log.Printf("Setting profile for mobile: %s", profileReq.MobileNo)

	// Validate session
	var session models.Session
	err := s.cassandraSession.Query(`
		SELECT session_token, mobile_no, is_active, expires_at
		FROM sessions
		WHERE session_token = ? AND mobile_no = ? AND is_active = true AND expires_at > ?
	`).Bind(profileReq.SessionToken, profileReq.MobileNo, true, time.Now()).Scan(&session.SessionToken, &session.MobileNo, &session.IsActive, &session.ExpiresAt)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Update user profile
	err = s.cassandraSession.Query(`
		UPDATE users
		SET full_name = ?, state = ?, referred_by = ?, profile_data = ?
		WHERE mobile_no = ?
	`).Bind(profileReq.FullName, profileReq.State, profileReq.ReferredBy, profileReq.ProfileData, profileReq.MobileNo).Exec()

	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %v", err)
	}

	log.Printf("Profile set successfully for %s", profileReq.MobileNo)

	return &models.SetProfileResponse{
		Status:         "success",
		Message:        "User profile updated successfully! 🎉",
		MobileNo:       profileReq.MobileNo,
		SessionToken:   profileReq.SessionToken,
		FullName:       profileReq.FullName,
		State:          profileReq.State,
		ReferralCode:   profileReq.ReferralCode,
		ReferredBy:     profileReq.ReferredBy,
		ProfileData:    profileReq.ProfileData,
		WelcomeMessage: fmt.Sprintf("Welcome %s! Your profile has been set up successfully.", profileReq.FullName),
		NextSteps:      "You can now proceed to set your language preferences.",
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		SocketID:       "",
		Event:          "profile:set",
	}, nil
}

// HandleSetLanguage sets user language preferences
func (s *SocketService) HandleSetLanguage(langReq models.SetLanguageRequest) (*models.SetLanguageResponse, error) {
	log.Printf("Setting language for mobile: %s", langReq.MobileNo)

	// Validate session
	var session models.Session
	err := s.cassandraSession.Query(`
		SELECT session_token, mobile_no, is_active, expires_at
		FROM sessions
		WHERE session_token = ? AND mobile_no = ? AND is_active = true AND expires_at > ?
	`).Bind(langReq.SessionToken, langReq.MobileNo, true, time.Now()).Scan(&session.SessionToken, &session.MobileNo, &session.IsActive, &session.ExpiresAt)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Update user language preferences
	err = s.cassandraSession.Query(`
		UPDATE users
		SET language_code = ?, language_name = ?, region_code = ?, timezone = ?, user_preferences = ?
		WHERE mobile_no = ?
	`).Bind(langReq.LanguageCode, langReq.LanguageName, langReq.RegionCode, langReq.Timezone, langReq.UserPreferences, langReq.MobileNo).Exec()

	if err != nil {
		return nil, fmt.Errorf("failed to update language preferences: %v", err)
	}

	log.Printf("Language set successfully for %s", langReq.MobileNo)

	return &models.SetLanguageResponse{
		Status:          "success",
		Message:         "Welcome to Game Admin! 🎮",
		MobileNo:        langReq.MobileNo,
		SessionToken:    langReq.SessionToken,
		LanguageCode:    langReq.LanguageCode,
		LanguageName:    langReq.LanguageName,
		RegionCode:      langReq.RegionCode,
		Timezone:        langReq.Timezone,
		UserPreferences: langReq.UserPreferences,
		LocalizedMessages: models.LocalizedMessages{
			Welcome:       "Welcome to Game Admin! 🎮",
			SetupComplete: "Setup completed successfully! ✅",
			ReadyToPlay:   "You're all set to start gaming! 🚀",
			NextSteps:     "Explore the dashboard and start managing your game experience.",
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SocketID:  "",
		Event:     "language:set",
	}, nil
}

// HandlePlayerAction processes player actions in gameplay
func (s *SocketService) HandlePlayerAction(actionReq models.PlayerActionRequest) (*models.PlayerActionResponse, error) {
	log.Printf("Player action received: %s from %s", actionReq.ActionType, actionReq.PlayerID)

	// Validate session
	var session models.Session
	err := s.cassandraSession.Query(`
		SELECT session_token, mobile_no, is_active, expires_at
		FROM sessions
		WHERE session_token = ? AND mobile_no = ? AND is_active = true AND expires_at > ?
	`).Bind(actionReq.SessionToken, actionReq.PlayerID, true, time.Now()).Scan(&session.SessionToken, &session.MobileNo, &session.IsActive, &session.ExpiresAt)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Generate action ID
	actionID := primitive.NewObjectID().Hex()

	// Process action based on type
	switch actionReq.ActionType {
	case "move":
		// Validate coordinates
		if actionReq.Coordinates.X < 0 || actionReq.Coordinates.Y < 0 {
			return nil, fmt.Errorf("invalid coordinates")
		}

		// Calculate distance moved (for demo purposes)
		distance := math.Sqrt(float64(actionReq.Coordinates.X*actionReq.Coordinates.X + actionReq.Coordinates.Y*actionReq.Coordinates.Y))
		log.Printf("Player moved distance: %.2f", distance)

	case "attack":
		// Handle attack action
		log.Printf("Player attacked with health: %d", actionReq.GameState.Health)

	case "collect":
		// Handle collect action
		log.Printf("Player collected item at level: %d", actionReq.GameState.Level)

	default:
		return nil, fmt.Errorf("unknown action type: %s", actionReq.ActionType)
	}

	log.Printf("Player action processed successfully: %s", actionID)

	return &models.PlayerActionResponse{
		Success:  true,
		Message:  "Action processed successfully",
		ActionID: actionID,
	}, nil
}

// HandleHeartbeat processes heartbeat from client
func (s *SocketService) HandleHeartbeat() models.HeartbeatResponse {
	return models.HeartbeatResponse{
		Success:   true,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// HandleWelcome sends welcome message to client
func (s *SocketService) HandleWelcome() models.WelcomeResponse {
	return models.WelcomeResponse{
		Success: true,
		Status:  "connected",
		Message: "Welcome to the game server!",
		ServerInfo: map[string]interface{}{
			"version":     "1.0.0",
			"server_time": time.Now().Format(time.RFC3339),
			"features":    []string{"real-time", "multiplayer", "chat"},
		},
	}
}

// HandleHealthCheck processes health check request
func (s *SocketService) HandleHealthCheck() models.HealthCheckResponse {
	return models.HealthCheckResponse{
		Success:   true,
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// ValidateSession validates if a session is active and not expired
func (s *SocketService) ValidateSession(sessionToken, mobileNo string) bool {
	var session models.Session
	err := s.cassandraSession.Query(`
		SELECT session_token, mobile_no, is_active, expires_at
		FROM sessions
		WHERE session_token = ? AND mobile_no = ? AND is_active = true AND expires_at > ?
	`).Bind(sessionToken, mobileNo, true, time.Now()).Scan(&session.SessionToken, &session.MobileNo, &session.IsActive, &session.ExpiresAt)

	return err == nil
}

// CleanupExpiredSessions removes expired sessions
func (s *SocketService) CleanupExpiredSessions() error {
	err := s.cassandraSession.Query(`
		DELETE FROM sessions
		WHERE expires_at < ?
	`).Bind(time.Now()).Exec()
	return err
}

// HandleStaticMessage handles static message requests including game list
func (s *SocketService) HandleStaticMessage(staticReq models.StaticMessageRequest) (*models.StaticMessageResponse, error) {
	log.Printf("Static message request received for mobile: %s, type: %s", staticReq.MobileNo, staticReq.MessageType)

	// Validate session
	var session models.Session
	err := s.cassandraSession.Query(`
		SELECT session_token, mobile_no, is_active, expires_at
		FROM sessions
		WHERE session_token = ? AND mobile_no = ? AND is_active = true AND expires_at > ?
	`).Bind(staticReq.SessionToken, staticReq.MobileNo, true, time.Now()).Scan(&session.SessionToken, &session.MobileNo, &session.IsActive, &session.ExpiresAt)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Prepare response data - only use sub_list
	log.Printf("🔍 Calling getGameSubListData() from HandleStaticMessage")
	responseData := s.getGameSubListData()
	log.Printf("📊 Response data received in HandleStaticMessage: %+v", responseData)
	log.Printf("🔍 Response data type: %T", responseData)
	if responseData != nil {
		log.Printf("📊 Response data keys: %v", getMapKeys(responseData))
		if gamelist, exists := responseData["gamelist"]; exists {
			if list, ok := gamelist.([]map[string]interface{}); ok {
				log.Printf("📊 Number of contests in response: %d", len(list))
			}
		}
	}

	log.Printf("Static message processed successfully for %s", staticReq.MobileNo)

	return &models.StaticMessageResponse{
		Status:       "success",
		Message:      "Static message retrieved successfully",
		MobileNo:     staticReq.MobileNo,
		SessionToken: staticReq.SessionToken,
		MessageType:  staticReq.MessageType,
		Data:         responseData,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		SocketID:     "",
		Event:        "static_message",
	}, nil
}

// GetGameListData returns sample game list data with Redis caching
func (s *SocketService) getGameListData() map[string]interface{} {
	// Try to get from Redis cache first
	if s.redisService != nil {
		cachedData, err := s.redisService.GetGameList()
		if err == nil {
			log.Printf("📖 Game list retrieved from Redis cache")
			return cachedData
		}
		log.Printf("📝 Game list not found in cache, generating fresh data")
	}

	// Generate fresh game list data
	gamelist := []map[string]interface{}{
		{
			"game_id":          "1",
			"game_name":        "Simple Ludo",
			"game_type":        "board",
			"active_gamepalye": 1500,
			"livegameplaye":    800,
			"status":           "active",
			"created_at":       time.Now().UTC().Format(time.RFC3339),
		},
		{
			"game_id":          "2",
			"game_name":        "Classic Ludo",
			"game_type":        "board",
			"active_gamepalye": 1000,
			"livegameplaye":    600,
			"status":           "active",
			"created_at":       time.Now().UTC().Format(time.RFC3339),
		},
		{
			"game_id":          "3",
			"game_name":        "Quick Ludo",
			"game_type":        "board",
			"active_gamepalye": 2000,
			"livegameplaye":    1200,
			"status":           "active",
			"created_at":       time.Now().UTC().Format(time.RFC3339),
		},
	}

	gameListData := map[string]interface{}{
		"gamelist": gamelist,
	}

	// Cache the data in Redis for 5 minutes
	if s.redisService != nil {
		err := s.redisService.CacheGameList(gameListData, 5*time.Minute)
		if err != nil {
			log.Printf("⚠️ Failed to cache game list in Redis: %v", err)
		} else {
			log.Printf("📝 Game list cached in Redis for 5 minutes")
		}
	}

	return gameListData
}

// GetGameListFromRedis returns game list data from Redis
func (s *SocketService) GetGameListFromRedis() (map[string]interface{}, error) {
	if s.redisService == nil {
		return nil, fmt.Errorf("Redis service not available")
	}
	return s.redisService.GetGameList()
}

// GetGameListDataPublic returns game list data (public method)
func (s *SocketService) GetGameListDataPublic() map[string]interface{} {
	return s.getGameListData()
}

// convertContestDataTypes converts float64 values back to integers in contest data
func (s *SocketService) convertContestDataTypes(data map[string]interface{}) map[string]interface{} {
	if gamelist, exists := data["gamelist"]; exists {
		if contests, ok := gamelist.([]interface{}); ok {
			for _, contest := range contests {
				if contestMap, ok := contest.(map[string]interface{}); ok {
					// Convert contest_id
					if contestID, exists := contestMap["contest_id"]; exists {
						if floatID, ok := contestID.(float64); ok {
							contestMap["contest_id"] = int(floatID)
						}
					}

					// Convert contest_joinuser
					if joinUser, exists := contestMap["contest_joinuser"]; exists {
						if floatJoin, ok := joinUser.(float64); ok {
							contestMap["contest_joinuser"] = int(floatJoin)
						}
					}

					// Convert contest_activeuser
					if activeUser, exists := contestMap["contest_activeuser"]; exists {
						if floatActive, ok := activeUser.(float64); ok {
							contestMap["contest_activeuser"] = int(floatActive)
						}
					}

					// Convert contest_win_price (keep as interface{} to support both int and float)
					if winPrice, exists := contestMap["contest_win_price"]; exists {
						if floatPrice, ok := winPrice.(float64); ok {
							// Check if it's a whole number
							if floatPrice == float64(int(floatPrice)) {
								contestMap["contest_win_price"] = int(floatPrice)
							} else {
								contestMap["contest_win_price"] = floatPrice
							}
						}
					}

					// Convert contest_entryfee (keep as interface{} to support both int and float)
					if entryFee, exists := contestMap["contest_entryfee"]; exists {
						if floatFee, ok := entryFee.(float64); ok {
							// Check if it's a whole number
							if floatFee == float64(int(floatFee)) {
								contestMap["contest_entryfee"] = int(floatFee)
							} else {
								contestMap["contest_entryfee"] = floatFee
							}
						}
					}
				}
			}
		}
	}
	return data
}

// getGameListData returns sample game list data with Redis caching
func (s *SocketService) getGameSubListData() map[string]interface{} {

	// Try to get from Redis cache first
	if s.redisService != nil {
		cachedData, err := s.redisService.GetListContest()
		if err == nil {
			// Convert data types after retrieving from cache
			convertedData := s.convertContestDataTypes(cachedData)
			return convertedData
		}
	} else {
	}

	// Generate fresh game list data with proper integer types
	gamelist := []map[string]interface{}{
		{
			"contest_id":         1,
			"contest_name":       "Weekly Algorithm Challenge",
			"contest_win_price":  5000,
			"contest_entryfee":   50,
			"contest_joinuser":   1000,
			"contest_activeuser": 847,
			"contest_starttime":  "2025-07-01T09:00:00Z",
			"contest_endtime":    "2025-07-07T23:59:59Z",
		},
		{
			"contest_id":         2,
			"contest_name":       "Data Science Hackathon",
			"contest_win_price":  10000,
			"contest_entryfee":   100,
			"contest_joinuser":   500,
			"contest_activeuser": 423,
			"contest_starttime":  "2025-06-15T10:00:00Z",
			"contest_endtime":    "2025-06-20T18:00:00Z",
		},
		{
			"contest_id":         3,
			"contest_name":       "Frontend Development Sprint",
			"contest_win_price":  3000,
			"contest_entryfee":   30,
			"contest_joinuser":   300,
			"contest_activeuser": 298,
			"contest_starttime":  "2025-06-10T08:00:00Z",
			"contest_endtime":    "2025-06-12T20:00:00Z",
		},
		{
			"contest_id":         4,
			"contest_name":       "Mobile App Innovation",
			"contest_win_price":  8000,
			"contest_entryfee":   80,
			"contest_joinuser":   800,
			"contest_activeuser": 156,
			"contest_starttime":  "2025-07-15T09:00:00Z",
			"contest_endtime":    "2025-07-25T23:59:59Z",
		},
		{
			"contest_id":         5,
			"contest_name":       "Blockchain Smart Contract Challenge",
			"contest_win_price":  5.0,
			"contest_entryfee":   0.1,
			"contest_joinuser":   400,
			"contest_activeuser": 89,
			"contest_starttime":  "2025-07-05T10:00:00Z",
			"contest_endtime":    "2025-07-10T18:00:00Z",
		},
		{
			"contest_id":         6,
			"contest_name":       "AI Chatbot Competition",
			"contest_win_price":  4000,
			"contest_entryfee":   40,
			"contest_joinuser":   600,
			"contest_activeuser": 567,
			"contest_starttime":  "2025-06-25T09:00:00Z",
			"contest_endtime":    "2025-06-27T17:00:00Z",
		},
	}

	gameListData := map[string]interface{}{
		"gamelist": gamelist,
	}

	// Cache the data in Redis for 5 minutes
	if s.redisService != nil {
		err := s.redisService.CacheListContest(gameListData, 5*time.Minute)

		if err != nil {
			log.Printf("⚠️ Failed to cache contest list in Redis: %v", err)
		} else {
			log.Printf("📝 Contest list cached in Redis for 5 minutes")
		}
	}
	return gameListData
}

// Helper function to get map keys for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// HandleMainScreen handles main screen requests with authentication validation
func (s *SocketService) HandleMainScreen(mainReq models.MainScreenRequest) (*models.MainScreenResponse, error) {
	// Decrypt the simple JWT token to get the original values used during token creation
	simpleJWTData, err := utils.ValidateSimpleJWTToken(mainReq.JWTToken)
	if err != nil {
		return nil, fmt.Errorf("simple JWT token validation failed: %v", err)
	}

	// Extract values from the decrypted JWT token
	tokenMobileNo := simpleJWTData.MobileNo
	tokenDeviceID := simpleJWTData.DeviceID
	tokenFCMToken := simpleJWTData.FCMToken

	// Validate mobile number from token
	if len(tokenMobileNo) < 10 {
		return nil, fmt.Errorf("invalid mobile number in JWT token")
	}

	// Validate device ID from token
	if len(tokenDeviceID) < 1 {
		return nil, fmt.Errorf("invalid device ID in JWT token")
	}

	// Check if user exists and is active using token mobile number
	var user models.User
	err = s.cassandraSession.Query(`
		SELECT id, mobile_no, full_name, status, language_code
		FROM users
		WHERE mobile_no = ?
	`).Bind(tokenMobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
	`).Bind(tokenMobileNo, tokenDeviceID, true, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Verify JWT token matches stored token
	if session.JWTToken != mainReq.JWTToken {
		return nil, fmt.Errorf("JWT token mismatch with stored token")
	}

	// Verify FCM token from JWT token matches the one in request
	if tokenFCMToken != mainReq.FCMToken {
		return nil, fmt.Errorf("FCM token mismatch - JWT token contains: %s, request contains: %s",
			tokenFCMToken[:20]+"...", mainReq.FCMToken[:20]+"...")
	}

	// Validate FCM token length
	if len(mainReq.FCMToken) < 100 {
		return nil, fmt.Errorf("FCM token too short or invalid")
	}

	// Prepare response data - only use sub_list
	responseData := s.getGameListData()

	log.Printf("Main screen processed successfully for %s", tokenMobileNo)

	return &models.MainScreenResponse{
		Status:      "success",
		Message:     "Main screen data retrieved successfully",
		MobileNo:    tokenMobileNo, // Use token mobile number
		DeviceID:    tokenDeviceID, // Use token device ID
		MessageType: mainReq.MessageType,
		Data:        responseData,
		UserInfo: map[string]interface{}{
			"user_id":   user.ID,
			"mobile_no": user.MobileNo,
			"full_name": user.FullName,
			"status":    user.Status,
			"language":  user.LanguageCode,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SocketID:  "",
		Event:     "main:screen:response",
	}, nil
}

// HandleContestList handles contest list requests with authentication validation
func (s *SocketService) HandleContestList(contestReq models.ContestRequest) (*models.ContestResponse, error) {
	// Decrypt the simple JWT token to get the original values used during token creation
	simpleJWTData, err := utils.ValidateSimpleJWTToken(contestReq.JWTToken)
	if err != nil {
		return nil, fmt.Errorf("simple JWT token validation failed: %v", err)
	}

	// Extract values from the decrypted JWT token
	tokenMobileNo := simpleJWTData.MobileNo
	tokenDeviceID := simpleJWTData.DeviceID
	tokenFCMToken := simpleJWTData.FCMToken

	// Validate mobile number from token
	if len(tokenMobileNo) < 10 {
		return nil, fmt.Errorf("invalid mobile number in JWT token")
	}

	// Validate device ID from token
	if len(tokenDeviceID) < 1 {
		return nil, fmt.Errorf("invalid device ID in JWT token")
	}

	// Check if user exists and is active using token mobile number
	var user models.User
	err = s.cassandraSession.Query(`
		SELECT id, mobile_no, full_name, status, language_code
		FROM users
		WHERE mobile_no = ?
	`).Bind(tokenMobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
	`).Bind(tokenMobileNo, tokenDeviceID, true, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Verify JWT token matches stored token
	if session.JWTToken != contestReq.JWTToken {
		return nil, fmt.Errorf("JWT token mismatch with stored token")
	}

	// Verify FCM token from JWT token matches the one in request
	if tokenFCMToken != contestReq.FCMToken {
		return nil, fmt.Errorf("FCM token mismatch - JWT token contains: %s, request contains: %s",
			tokenFCMToken[:20]+"...", contestReq.FCMToken[:20]+"...")
	}

	// Validate FCM token length
	if len(contestReq.FCMToken) < 100 {
		return nil, fmt.Errorf("FCM token too short or invalid")
	}

	// Get contest list data specifically

	responseData := s.getGameSubListData()
	if responseData != nil {
		log.Printf("📊 Response data keys: %v", getMapKeys(responseData))
		if gamelist, exists := responseData["gamelist"]; exists {
			if list, ok := gamelist.([]map[string]interface{}); ok {
				log.Printf("📊 Number of contests in response: %d", len(list))
			}
		}
	}

	return &models.ContestResponse{
		Status:      "success",
		Message:     "Contest list data retrieved successfully",
		MobileNo:    tokenMobileNo, // Use token mobile number
		DeviceID:    tokenDeviceID, // Use token device ID
		MessageType: contestReq.MessageType,
		Data:        responseData,
		UserInfo: map[string]interface{}{
			"user_id":   user.ID,
			"mobile_no": user.MobileNo,
			"full_name": user.FullName,
			"status":    user.Status,
			"language":  user.LanguageCode,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SocketID:  "",
		Event:     "contest:list:response",
	}, nil
}

// HandleContestJoin handles contest join requests
func (s *SocketService) HandleContestJoin(joinReq models.ContestJoinRequest) (*models.ContestJoinResponse, error) {
	// Decrypt the simple JWT token to get the original values used during token creation
	simpleJWTData, err := utils.ValidateSimpleJWTToken(joinReq.JWTToken)
	if err != nil {
		return nil, fmt.Errorf("simple JWT token validation failed: %v", err)
	}

	// Extract values from the decrypted JWT token
	tokenMobileNo := simpleJWTData.MobileNo
	tokenDeviceID := simpleJWTData.DeviceID
	tokenFCMToken := simpleJWTData.FCMToken

	// Validate mobile number from token
	if len(tokenMobileNo) < 10 {
		return nil, fmt.Errorf("invalid mobile number in JWT token")
	}

	// Validate device ID from token
	if len(tokenDeviceID) < 1 {
		return nil, fmt.Errorf("invalid device ID in JWT token")
	}

	// Check if user exists and is active using token mobile number
	var user models.User
	err = s.cassandraSession.Query(`
		SELECT id, mobile_no, full_name, status, language_code
		FROM users
		WHERE mobile_no = ?
	`).Bind(tokenMobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
	`).Bind(tokenMobileNo, tokenDeviceID, true, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Verify JWT token matches stored token
	if session.JWTToken != joinReq.JWTToken {
		return nil, fmt.Errorf("JWT token mismatch with stored token")
	}

	// Verify FCM token from JWT token matches the one in request
	if tokenFCMToken != joinReq.FCMToken {
		return nil, fmt.Errorf("FCM token mismatch - JWT token contains: %s, request contains: %s",
			tokenFCMToken[:20]+"...", joinReq.FCMToken[:20]+"...")
	}

	// Validate FCM token length
	if len(joinReq.FCMToken) < 100 {
		return nil, fmt.Errorf("FCM token too short or invalid")
	}

	// Validate contest ID
	if joinReq.ContestID == "" {
		return nil, fmt.Errorf("contest ID is required")
	}

	// Generate team ID if team name is provided
	teamID := ""
	if joinReq.TeamName != "" {
		teamID = fmt.Sprintf("team_%s_%s", joinReq.ContestID, time.Now().Format("20060102150405"))
	}

	log.Printf("Contest join processed successfully for %s - Contest: %s", tokenMobileNo, joinReq.ContestID)

	return &models.ContestJoinResponse{
		Status:    "success",
		Message:   "Successfully joined contest",
		MobileNo:  tokenMobileNo,
		DeviceID:  tokenDeviceID,
		ContestID: joinReq.ContestID,
		TeamID:    teamID,
		JoinTime:  time.Now().UTC().Format(time.RFC3339),
		Data: map[string]interface{}{
			"contest_id":  joinReq.ContestID,
			"team_name":   joinReq.TeamName,
			"team_size":   joinReq.TeamSize,
			"join_status": "confirmed",
			"next_steps":  "Wait for contest start",
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SocketID:  "",
		Event:     "contest:join:response",
	}, nil
}

// HandleListContestScreen handles contest list screen requests with authentication validation
func (s *SocketService) HandleListContestScreen(mainReq models.MainScreenRequest) (*models.MainScreenResponse, error) {
	// Convert MainScreenRequest to ContestRequest for consistency
	contestReq := models.ContestRequest{
		MobileNo:    mainReq.MobileNo,
		FCMToken:    mainReq.FCMToken,
		JWTToken:    mainReq.JWTToken,
		DeviceID:    mainReq.DeviceID,
		MessageType: mainReq.MessageType,
	}

	// Use the dedicated contest handler
	contestResponse, err := s.HandleContestList(contestReq)
	if err != nil {
		return nil, err
	}

	// Convert ContestResponse back to MainScreenResponse for backward compatibility
	return &models.MainScreenResponse{
		Status:      contestResponse.Status,
		Message:     contestResponse.Message,
		MobileNo:    contestResponse.MobileNo,
		DeviceID:    contestResponse.DeviceID,
		MessageType: contestResponse.MessageType,
		Data:        contestResponse.Data,
		UserInfo:    contestResponse.UserInfo,
		Timestamp:   contestResponse.Timestamp,
		SocketID:    contestResponse.SocketID,
		Event:       contestResponse.Event,
	}, nil
}

// HandleContestGap handles contest price gap requests with authentication validation
func (s *SocketService) HandleContestGap(gapReq models.ContestGapRequest) (*models.ContestGapResponse, error) {
	// Decrypt the simple JWT token to get the original values used during token creation
	simpleJWTData, err := utils.ValidateSimpleJWTToken(gapReq.JWTToken)
	if err != nil {
		return nil, fmt.Errorf("simple JWT token validation failed: %v", err)
	}

	// Extract values from the decrypted JWT token
	tokenMobileNo := simpleJWTData.MobileNo
	tokenDeviceID := simpleJWTData.DeviceID
	tokenFCMToken := simpleJWTData.FCMToken

	// Validate mobile number from token
	if len(tokenMobileNo) < 10 {
		return nil, fmt.Errorf("invalid mobile number in JWT token")
	}

	// Validate device ID from token
	if len(tokenDeviceID) < 1 {
		return nil, fmt.Errorf("invalid device ID in JWT token")
	}

	// Check if user exists and is active using token mobile number
	var user models.User
	err = s.cassandraSession.Query(`
		SELECT id, mobile_no, full_name, status, language_code
		FROM users
		WHERE mobile_no = ?
	`).Bind(tokenMobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
	`).Bind(tokenMobileNo, tokenDeviceID, true, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Verify JWT token matches stored token
	if session.JWTToken != gapReq.JWTToken {
		return nil, fmt.Errorf("JWT token mismatch with stored token")
	}

	// Verify FCM token from JWT token matches the one in request
	if tokenFCMToken != gapReq.FCMToken {
		return nil, fmt.Errorf("FCM token mismatch - JWT token contains: %s, request contains: %s",
			tokenFCMToken[:20]+"...", gapReq.FCMToken[:20]+"...")
	}

	// Validate FCM token length
	if len(gapReq.FCMToken) < 100 {
		return nil, fmt.Errorf("FCM token too short or invalid")
	}

	// Get contest list data and calculate price gaps
	allContests := s.getGameSubListData()

	// Calculate price gap data
	gapData := s.calculatePriceGapData(allContests, gapReq)

	log.Printf("💰 Contest price gap processed successfully for %s - Type: %s", tokenMobileNo, gapReq.MessageType)

	return &models.ContestGapResponse{
		Status:      "success",
		Message:     "Contest price gap data retrieved successfully",
		MobileNo:    tokenMobileNo, // Use token mobile number
		DeviceID:    tokenDeviceID, // Use token device ID
		MessageType: gapReq.MessageType,
		Data:        gapData,
		UserInfo: map[string]interface{}{
			"user_id":   user.ID,
			"mobile_no": user.MobileNo,
			"full_name": user.FullName,
			"status":    user.Status,
			"language":  user.LanguageCode,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SocketID:  "",
		Event:     "list:contest:gap:response",
	}, nil
}

// calculatePriceGapData calculates price gap information from contest data
func (s *SocketService) calculatePriceGapData(allContests map[string]interface{}, gapReq models.ContestGapRequest) map[string]interface{} {
	gapData := map[string]interface{}{
		"price_gaps": []map[string]interface{}{},
		"summary":    map[string]interface{}{},
	}

	log.Printf("🔍 Starting calculatePriceGapData - MessageType: %s", gapReq.MessageType)
	log.Printf("🔍 All contests keys: %v", getMapKeys(allContests))

	if gamelist, exists := allContests["gamelist"]; exists {
		log.Printf("🔍 Found gamelist, type: %T", gamelist)

		var contests []map[string]interface{}

		// Handle different possible types
		switch v := gamelist.(type) {
		case []map[string]interface{}:
			contests = v
			log.Printf("🔍 Direct []map[string]interface{} type, count: %d", len(contests))
		case []interface{}:
			// Convert []interface{} to []map[string]interface{}
			contests = make([]map[string]interface{}, 0, len(v))
			for _, item := range v {
				if contest, ok := item.(map[string]interface{}); ok {
					contests = append(contests, contest)
				}
			}
			log.Printf("🔍 Converted []interface{} to []map[string]interface{}, count: %d", len(contests))
		default:
			log.Printf("🔍 Unknown gamelist type: %T", gamelist)
			return gapData
		}

		if len(contests) > 0 {
			var winPrices []float64
			var entryFees []float64

			log.Printf("🔍 Processing %d contests", len(contests))

			// Extract all prices and fees
			for i, contest := range contests {
				log.Printf("🔍 Contest %d keys: %v", i, getMapKeys(contest))

				if winPrice, exists := contest["contest_win_price"]; exists {
					price := s.convertToFloat(winPrice)
					winPrices = append(winPrices, price)
					log.Printf("🔍 Contest %d win price: %v -> %f", i, winPrice, price)
				}

				if entryFee, exists := contest["contest_entryfee"]; exists {
					fee := s.convertToFloat(entryFee)
					entryFees = append(entryFees, fee)
					log.Printf("🔍 Contest %d entry fee: %v -> %f", i, entryFee, fee)
				}
			}

			log.Printf("🔍 Extracted %d win prices, %d entry fees", len(winPrices), len(entryFees))

			// Calculate gaps based on message type
			switch gapReq.MessageType {
			case "win_price_gap":
				gapData["price_gaps"] = s.calculateWinPriceGaps(winPrices)
			case "entry_fee_gap":
				gapData["price_gaps"] = s.calculateEntryFeeGaps(entryFees)
			case "price_gap":
				gapData["price_gaps"] = s.calculateCombinedPriceGaps(winPrices, entryFees)
			default:
				gapData["price_gaps"] = s.calculateAllPriceGaps(winPrices, entryFees)
			}

			// Add summary statistics
			gapData["summary"] = map[string]interface{}{
				"total_contests":  len(contests),
				"filter_type":     gapReq.MessageType,
				"win_price_range": s.calculateRange(winPrices),
				"entry_fee_range": s.calculateRange(entryFees),
				"avg_win_price":   s.calculateAverage(winPrices),
				"avg_entry_fee":   s.calculateAverage(entryFees),
			}
		} else {
			log.Printf("🔍 No contests found in gamelist")
		}
	} else {
		log.Printf("🔍 No gamelist found in allContests")
	}

	log.Printf("🔍 Final gap data: %+v", gapData)
	return gapData
}

// calculateWinPriceGaps calculates win price gap ranges
func (s *SocketService) calculateWinPriceGaps(prices []float64) []map[string]interface{} {
	if len(prices) == 0 {
		return []map[string]interface{}{}
	}

	// Sort prices
	sort.Float64s(prices)

	var gaps []map[string]interface{}

	// Create price ranges
	ranges := []struct {
		min  float64
		max  float64
		name string
	}{
		{0, 1000, "Low"},
		{1000, 3000, "Medium"},
		{3000, 5000, "High"},
		{5000, 10000, "Premium"},
		{10000, 999999, "Elite"},
	}

	for _, r := range ranges {
		count := 0
		for _, price := range prices {
			if price >= r.min && price < r.max {
				count++
			}
		}

		if count > 0 {
			gaps = append(gaps, map[string]interface{}{
				"type":          "win_price",
				"range_name":    r.name,
				"min_price":     r.min,
				"max_price":     r.max,
				"contest_count": count,
				"percentage":    float64(count) / float64(len(prices)) * 100,
			})
		}
	}

	return gaps
}

// calculateEntryFeeGaps calculates entry fee gap ranges
func (s *SocketService) calculateEntryFeeGaps(fees []float64) []map[string]interface{} {
	if len(fees) == 0 {
		return []map[string]interface{}{}
	}

	// Sort fees
	sort.Float64s(fees)

	var gaps []map[string]interface{}

	// Create fee ranges
	ranges := []struct {
		min  float64
		max  float64
		name string
	}{
		{0, 10, "Free"},
		{10, 50, "Low"},
		{50, 100, "Medium"},
		{100, 200, "High"},
		{200, 999999, "Premium"},
	}

	for _, r := range ranges {
		count := 0
		for _, fee := range fees {
			if fee >= r.min && fee < r.max {
				count++
			}
		}

		if count > 0 {
			gaps = append(gaps, map[string]interface{}{
				"type":          "entry_fee",
				"range_name":    r.name,
				"min_price":     r.min,
				"max_price":     r.max,
				"contest_count": count,
				"percentage":    float64(count) / float64(len(fees)) * 100,
			})
		}
	}

	return gaps
}

// calculateCombinedPriceGaps calculates both win price and entry fee gaps
func (s *SocketService) calculateCombinedPriceGaps(prices, fees []float64) []map[string]interface{} {
	winGaps := s.calculateWinPriceGaps(prices)
	entryGaps := s.calculateEntryFeeGaps(fees)

	var combined []map[string]interface{}
	combined = append(combined, winGaps...)
	combined = append(combined, entryGaps...)

	return combined
}

// calculateAllPriceGaps calculates all possible price gaps
func (s *SocketService) calculateAllPriceGaps(prices, fees []float64) []map[string]interface{} {
	return s.calculateCombinedPriceGaps(prices, fees)
}

// calculateRange calculates min, max, and range for a slice of values
func (s *SocketService) calculateRange(values []float64) map[string]interface{} {
	if len(values) == 0 {
		return map[string]interface{}{
			"min":   0,
			"max":   0,
			"range": 0,
		}
	}

	sort.Float64s(values)
	min := values[0]
	max := values[len(values)-1]

	return map[string]interface{}{
		"min":   min,
		"max":   max,
		"range": max - min,
	}
}

// calculateAverage calculates the average of a slice of values
func (s *SocketService) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

// convertToFloat converts interface{} to float64 for price comparisons
func (s *SocketService) convertToFloat(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}
