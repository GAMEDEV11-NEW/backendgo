package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gofiber/app/models"
	"gofiber/app/utils"
	"gofiber/redis"
	"math"
	"math/big"
	"sort"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/google/uuid"
)

// SocketService handles all socket-related business logic
// Refactored to use Cassandra
type SocketService struct {
	cassandraSession *gocql.Session
	redisService     *redis.Service
}

// NewSocketService creates a new socket service instance using Cassandra
func NewSocketService(cassandraSession *gocql.Session) *SocketService {
	if cassandraSession == nil {
		panic("Cassandra session cannot be nil")
	}
	redisService := redis.NewService()
	service := &SocketService{
		cassandraSession: cassandraSession,
		redisService:     redisService,
	}
	return service
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
	return models.DeviceInfoResponse{
		Status:    "success",
		Message:   "Device info received and validated",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
			SocketID:  socketID,
			Event:     "device:info:ack",
	}
}

// HandleLogin handles user login and OTP generation
func (s *SocketService) HandleLogin(loginReq models.LoginRequest) (*models.LoginResponse, error) {
	// Validate mobile number format
	if len(loginReq.MobileNo) != 10 {
		return nil, fmt.Errorf("invalid mobile number format")
	}

	// Check if user exists
	var existingUser models.User
	err := s.cassandraSession.Query(`
		SELECT mobile_no, status, created_at
		FROM users
		WHERE mobile_no = ?
		ALLOW FILTERING
	`).Bind(loginReq.MobileNo).Scan(&existingUser.MobileNo, &existingUser.Status, &existingUser.CreatedAt)

	if err != nil {
		// Create new user
		userID := uuid.New().String()
		now := time.Now()
		err = s.cassandraSession.Query(`
			INSERT INTO users (id, mobile_no, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
		`).Bind(userID, loginReq.MobileNo, "new_user", now, now).Exec()
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %v", err)
		}
		existingUser.Status = "new_user"
	}

	// Generate OTP
	otp := s.GenerateOTP()

	// Store OTP in database with expiry
	expiryTime := time.Now().Add(5 * time.Minute)
	err = s.cassandraSession.Query(`
		INSERT INTO otp_store (phone_or_email, otp_code, purpose, created_at, expires_at, is_verified, attempt_count)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`).Bind(loginReq.MobileNo, strconv.Itoa(otp), "login", time.Now().Format(time.RFC3339), expiryTime.Format(time.RFC3339), false, 0).Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to store OTP: %v", err)
	}

	// Create session
	sessionToken := uuid.New().String()
	sessionExpiry := time.Now().Add(24 * time.Hour)
	err = s.cassandraSession.Query(`
		INSERT INTO sessions (session_token, mobile_no, device_id, fcm_token, created_at, expires_at, is_active, jwt_token)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`).Bind(sessionToken, loginReq.MobileNo, loginReq.DeviceID, loginReq.FCMToken, time.Now(), sessionExpiry, true, "").Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	// Send OTP via SMS (mock implementation)
	// In a real implementation, you would integrate with an SMS service here

	// Create response
	response := &models.LoginResponse{
		Status:       "success",
		Message:      "OTP sent successfully",
		MobileNo:     loginReq.MobileNo,
		DeviceID:     loginReq.DeviceID,
		SessionToken: sessionToken,
		OTP:          otp,
		IsNewUser:    existingUser.Status == "new_user",
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		SocketID:     "",
		Event:        "login:success",
	}

	return response, nil
}

// HandleOTPVerification verifies OTP and returns user status
func (s *SocketService) HandleOTPVerification(otpReq models.OTPVerificationRequest) (*models.OTPVerificationResponse, error) {
	// Validate session
	var session models.Session
	err := s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active, created_at
		FROM sessions
		WHERE session_token = ? AND mobile_no = ? AND is_active = true AND expires_at > ?
		ALLOW FILTERING
	`).Bind(otpReq.SessionToken, otpReq.MobileNo, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive, &session.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Validate OTP format
	otpInt, err := strconv.Atoi(otpReq.OTP)
	if err != nil || otpInt < 100000 || otpInt > 999999 {
		return nil, fmt.Errorf("invalid OTP format")
	}

	// Verify OTP against stored OTP in database
	var storedOTP struct {
		OTPCode      string `json:"otp_code"`
		ExpiresAt    string `json:"expires_at"`
		IsVerified   bool   `json:"is_verified"`
		AttemptCount int    `json:"attempt_count"`
		CreatedAt    string `json:"created_at"`
	}

	err = s.cassandraSession.Query(`
		SELECT otp_code, expires_at, is_verified, attempt_count, created_at
		FROM otp_store
		WHERE phone_or_email = ? AND purpose = ?
		ORDER BY created_at DESC
		LIMIT 1
		ALLOW FILTERING
	`).Bind(otpReq.MobileNo, "login").Scan(&storedOTP.OTPCode, &storedOTP.ExpiresAt, &storedOTP.IsVerified, &storedOTP.AttemptCount, &storedOTP.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("OTP not found or expired")
	}

	// Check if OTP is already verified
	if storedOTP.IsVerified {
		return nil, fmt.Errorf("OTP already verified")
	}

	// Check if OTP is expired
	expiryTime, err := time.Parse(time.RFC3339, storedOTP.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("invalid OTP expiry time")
	}
	if time.Now().After(expiryTime) {
		return nil, fmt.Errorf("OTP has expired")
	}

	// Check attempt count
	if storedOTP.AttemptCount >= 3 {
		return nil, fmt.Errorf("maximum OTP attempts exceeded")
	}

	// Verify OTP
	if storedOTP.OTPCode != otpReq.OTP {
		// Increment attempt count
		err = s.cassandraSession.Query(`
			UPDATE otp_store
			SET attempt_count = ?
			WHERE phone_or_email = ? AND purpose = ? AND created_at = ?
		`).Bind(storedOTP.AttemptCount+1, otpReq.MobileNo, "login", storedOTP.CreatedAt).Exec()
		if err != nil {
			return nil, fmt.Errorf("failed to update attempt count: %v", err)
		}
		return nil, fmt.Errorf("invalid OTP")
	}

	// Mark OTP as verified
	err = s.cassandraSession.Query(`
		UPDATE otp_store
		SET is_verified = ?
		WHERE phone_or_email = ? AND purpose = ? AND created_at = ?
	`).Bind(true, otpReq.MobileNo, "login", storedOTP.CreatedAt).Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to mark OTP as verified: %v", err)
	}

	// Check user status
	var user models.User
	err = s.cassandraSession.Query(`
		SELECT id, status
		FROM users
		WHERE mobile_no = ?
		ALLOW FILTERING
	`).Bind(otpReq.MobileNo).Scan(&user.ID, &user.Status)
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
			WHERE id = ?
		`).Bind("existing_user", user.ID).Exec()
		if err != nil {
			return nil, fmt.Errorf("failed to update user status: %v", err)
		}
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
		WHERE mobile_no = ? AND device_id = ? AND created_at = ?
	`).Bind(jwtToken, otpReq.MobileNo, session.DeviceID, session.CreatedAt).Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to update session with JWT token: %v", err)
	}

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
	// Validate session
	var session models.Session
	err := s.cassandraSession.Query(`
		SELECT session_token, mobile_no, is_active, expires_at
		FROM sessions
		WHERE session_token = ? AND mobile_no = ? AND is_active = true AND expires_at > ?
		ALLOW FILTERING
	`).Bind(profileReq.SessionToken, profileReq.MobileNo, time.Now()).Scan(&session.SessionToken, &session.MobileNo, &session.IsActive, &session.ExpiresAt)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Get user ID first
	var user models.User
	err = s.cassandraSession.Query(`
		SELECT id
		FROM users
		WHERE mobile_no = ?
		ALLOW FILTERING
	`).Bind(profileReq.MobileNo).Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update user profile
	
	// Convert ProfileData to JSON string
	profileDataJSON := ""
	if profileReq.ProfileData.Avatar != "" || profileReq.ProfileData.Bio != "" || len(profileReq.ProfileData.Preferences) > 0 {
		profileDataBytes, err := json.Marshal(profileReq.ProfileData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal profile data: %v", err)
		}
		profileDataJSON = string(profileDataBytes)
	}
	
	err = s.cassandraSession.Query(`
		UPDATE users
		SET full_name = ?, state = ?, referred_by = ?, profile_data = ?
		WHERE id = ?
	`).Bind(profileReq.FullName, profileReq.State, profileReq.ReferredBy, profileDataJSON, user.ID).Exec()

	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %v", err)
	}

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
	// Validate session
	var session models.Session
	err := s.cassandraSession.Query(`
		SELECT session_token, mobile_no, is_active, expires_at
		FROM sessions
		WHERE session_token = ? AND mobile_no = ? AND is_active = true AND expires_at > ?
			ALLOW FILTERING
	`).Bind(langReq.SessionToken, langReq.MobileNo, time.Now()).Scan(&session.SessionToken, &session.MobileNo, &session.IsActive, &session.ExpiresAt)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Get user ID first
	var user models.User
	err = s.cassandraSession.Query(`
		SELECT id
		FROM users
		WHERE mobile_no = ?
		ALLOW FILTERING
	`).Bind(langReq.MobileNo).Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update user language preferences
	
	// Convert UserPreferences to JSON string
	userPreferencesJSON := ""
	if langReq.UserPreferences.DateFormat != "" || langReq.UserPreferences.TimeFormat != "" || langReq.UserPreferences.Currency != "" {
		userPreferencesBytes, err := json.Marshal(langReq.UserPreferences)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal user preferences: %v", err)
		}
		userPreferencesJSON = string(userPreferencesBytes)
	}
	
	err = s.cassandraSession.Query(`
		UPDATE users
		SET language_code = ?, language_name = ?, region_code = ?, timezone = ?, user_preferences = ?
		WHERE id = ?
	`).Bind(langReq.LanguageCode, langReq.LanguageName, langReq.RegionCode, langReq.Timezone, userPreferencesJSON, user.ID).Exec()

	if err != nil {
		return nil, fmt.Errorf("failed to update language preferences: %v", err)
	}

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
		_ = math.Sqrt(float64(actionReq.Coordinates.X*actionReq.Coordinates.X + actionReq.Coordinates.Y*actionReq.Coordinates.Y))

	case "attack":
		// Handle attack action
		_ = actionReq.GameState.Health

	case "collect":
		// Handle collect action
		_ = actionReq.GameState.Level

	default:
		return nil, fmt.Errorf("unknown action type: %s", actionReq.ActionType)
	}

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

// CleanupExpiredOTPs removes expired OTPs from the database
func (s *SocketService) CleanupExpiredOTPs() error {
	err := s.cassandraSession.Query(`
		DELETE FROM otp_store
		WHERE expires_at < ?
	`).Bind(time.Now().Format(time.RFC3339)).Exec()
	if err != nil {
		return err
	}
	return nil
}

// GetLatestOTP retrieves the latest OTP for a given phone number and purpose
func (s *SocketService) GetLatestOTP(phoneOrEmail, purpose string) (*models.OTPData, error) {
	var otpData models.OTPData
	err := s.cassandraSession.Query(`
		SELECT phone_or_email, otp_code, created_at, expires_at, purpose, is_verified, attempt_count
		FROM otp_store
		WHERE phone_or_email = ? AND purpose = ?
		ORDER BY created_at DESC
		LIMIT 1
	`).Bind(phoneOrEmail, purpose).Scan(
		&otpData.PhoneOrEmail,
		&otpData.OTPCode,
		&otpData.CreatedAt,
		&otpData.ExpiresAt,
		&otpData.Purpose,
		&otpData.IsVerified,
		&otpData.AttemptCount,
	)
	if err != nil {
		return nil, err
	}
	return &otpData, nil
}

// ResendOTP generates and stores a new OTP for the given phone number
func (s *SocketService) ResendOTP(mobileNo string) (int, error) {
	// Generate new OTP
	otp := s.GenerateOTP()
	
	// Store new OTP in database
	otpExpiry := time.Now().Add(10 * time.Minute) // OTP expires in 10 minutes
	err := s.cassandraSession.Query(`
		INSERT INTO otp_store (phone_or_email, otp_code, created_at, expires_at, purpose, is_verified, attempt_count)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`).Bind(
		mobileNo,
		strconv.Itoa(otp),
		time.Now().Format(time.RFC3339),
		otpExpiry.Format(time.RFC3339),
		"login",
		false,
		0,
	).Exec()
	if err != nil {
		return 0, fmt.Errorf("failed to store new OTP: %v", err)
	}
	
	return otp, nil
}

// HandleStaticMessage handles static message requests including game list
func (s *SocketService) HandleStaticMessage(staticReq models.StaticMessageRequest) (*models.StaticMessageResponse, error) {
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
	responseData := s.getGameSubListData()

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
			return cachedData
		}
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
			return gameListData
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
			return gameListData
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
		ALLOW FILTERING
	`).Bind(tokenMobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active, jwt_token
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
		ALLOW FILTERING
	`).Bind(tokenMobileNo, tokenDeviceID, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive, &session.JWTToken)

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
		ALLOW FILTERING
	`).Bind(tokenMobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active, jwt_token
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
		ALLOW FILTERING
	`).Bind(tokenMobileNo, tokenDeviceID, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive, &session.JWTToken)

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
	// Validate mobile number from token
	if len(joinReq.MobileNo) < 10 {
		return nil, fmt.Errorf("invalid mobile number in JWT token")
	}

	// Validate device ID from token
	if len(joinReq.DeviceID) < 1 {
		return nil, fmt.Errorf("invalid device ID in JWT token")
	}

	// Check if user exists and is active using token mobile number
	var user models.User
	err := s.cassandraSession.Query(`
		SELECT id, mobile_no, full_name, status, language_code
		FROM users
		WHERE mobile_no = ?
		ALLOW FILTERING
	`).Bind(joinReq.MobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active, jwt_token
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
		ALLOW FILTERING
	`).Bind(joinReq.MobileNo, joinReq.DeviceID, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive, &session.JWTToken)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Verify JWT token matches stored token
	if session.JWTToken != joinReq.JWTToken {
		return nil, fmt.Errorf("JWT token mismatch with stored token")
	}

	// Verify FCM token from JWT token matches the one in request
	if joinReq.FCMToken != session.FCMToken {
		return nil, fmt.Errorf("FCM token mismatch - JWT token contains: %s, request contains: %s",
			session.FCMToken[:20]+"...", joinReq.FCMToken[:20]+"...")
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

	return &models.ContestJoinResponse{
		Status:    "success",
		Message:   "Successfully joined contest",
		MobileNo:  joinReq.MobileNo,
		DeviceID:  joinReq.DeviceID,
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
		ALLOW FILTERING
	`).Bind(tokenMobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active, jwt_token
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
		ALLOW FILTERING
	`).Bind(tokenMobileNo, tokenDeviceID, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive, &session.JWTToken)

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

	if gamelist, exists := allContests["gamelist"]; exists {
		var contests []map[string]interface{}

		// Handle different possible types
		switch v := gamelist.(type) {
		case []map[string]interface{}:
			contests = v
		case []interface{}:
			// Convert []interface{} to []map[string]interface{}
			contests = make([]map[string]interface{}, 0, len(v))
			for _, item := range v {
				if contest, ok := item.(map[string]interface{}); ok {
					contests = append(contests, contest)
				}
			}
		default:
			return gapData
		}

		if len(contests) > 0 {
			var winPrices []float64
			var entryFees []float64

			// Extract all prices and fees
			for _, contest := range contests {
				if winPrice, exists := contest["contest_win_price"]; exists {
					price := s.convertToFloat(winPrice)
					winPrices = append(winPrices, price)
				}

				if entryFee, exists := contest["contest_entryfee"]; exists {
					fee := s.convertToFloat(entryFee)
					entryFees = append(entryFees, fee)
				}
			}

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
		}
	}

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

// GetCassandraSession returns the Cassandra session for external access
func (s *SocketService) GetCassandraSession() *gocql.Session {
	return s.cassandraSession
}
