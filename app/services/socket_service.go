package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"gofiber/app/models"
	"gofiber/app/utils"
	"gofiber/redis"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SocketService handles all socket-related business logic
type SocketService struct {
	usersCollection    *mongo.Collection
	sessionsCollection *mongo.Collection
	redisService       *redis.Service
}

// NewSocketService creates a new socket service instance
func NewSocketService(usersCollection, sessionsCollection *mongo.Collection) *SocketService {
	redisService := redis.NewService()

	return &SocketService{
		usersCollection:    usersCollection,
		sessionsCollection: sessionsCollection,
		redisService:       redisService,
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

	// Create context for database operations
	ctx := context.Background()

	// Check if user exists
	var existingUser models.User
	err = s.usersCollection.FindOne(ctx, bson.M{"mobile_no": loginReq.MobileNo}).Decode(&existingUser)

	var userID string
	var isNewUser bool
	if err == mongo.ErrNoDocuments {
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

		_, err = s.usersCollection.InsertOne(ctx, user)
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
	err = s.sessionsCollection.FindOne(ctx, bson.M{
		"mobile_no":  loginReq.MobileNo,
		"device_id":  loginReq.DeviceID,
		"is_active":  true,
		"expires_at": bson.M{"$gt": time.Now()},
	}).Decode(&existingSession)

	if err == nil {
		// Existing session found, update it with new session token
		log.Printf("Updating existing session for mobile: %s, device: %s", loginReq.MobileNo, loginReq.DeviceID)

		update := bson.M{
			"$set": bson.M{
				"session_token": sessionToken,
				"fcm_token":     loginReq.FCMToken,
				"updated_at":    time.Now(),
				"expires_at":    time.Now().Add(24 * time.Hour),
				"is_active":     true,
			},
		}

		_, err = s.sessionsCollection.UpdateOne(
			ctx,
			bson.M{"_id": existingSession.ID},
			update,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update existing session: %v", err)
		}

		log.Printf("Existing session updated for %s", loginReq.MobileNo)
	} else if err == mongo.ErrNoDocuments {
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

		_, err = s.sessionsCollection.InsertOne(ctx, session)
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

	// Create context for database operations
	ctx := context.Background()

	// Validate session
	var session models.Session
	err := s.sessionsCollection.FindOne(ctx, bson.M{
		"session_token": otpReq.SessionToken,
		"mobile_no":     otpReq.MobileNo,
		"is_active":     true,
		"expires_at":    bson.M{"$gt": time.Now()},
	}).Decode(&session)

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
	err = s.usersCollection.FindOne(ctx, bson.M{"mobile_no": otpReq.MobileNo}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Determine user status and update if needed
	userStatus := user.Status
	if user.Status == "new_user" {
		// Update user status to existing_user
		_, err = s.usersCollection.UpdateOne(
			ctx,
			bson.M{"mobile_no": otpReq.MobileNo},
			bson.M{"$set": bson.M{"status": "existing_user", "updated_at": time.Now()}},
		)
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
	_, err = s.sessionsCollection.UpdateOne(
		ctx,
		bson.M{"session_token": otpReq.SessionToken},
		bson.M{"$set": bson.M{
			"jwt_token":  jwtToken,
			"updated_at": time.Now(),
		}},
	)
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

	// Create context for database operations
	ctx := context.Background()

	// Validate session
	var session models.Session
	err := s.sessionsCollection.FindOne(ctx, bson.M{
		"session_token": profileReq.SessionToken,
		"mobile_no":     profileReq.MobileNo,
		"is_active":     true,
		"expires_at":    bson.M{"$gt": time.Now()},
	}).Decode(&session)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Update user profile
	update := bson.M{
		"$set": bson.M{
			"full_name":    profileReq.FullName,
			"state":        profileReq.State,
			"referred_by":  profileReq.ReferredBy,
			"profile_data": profileReq.ProfileData,
			"updated_at":   time.Now(),
		},
	}

	result, err := s.usersCollection.UpdateOne(
		ctx,
		bson.M{"mobile_no": profileReq.MobileNo},
		update,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %v", err)
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("user not found")
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

	// Create context for database operations
	ctx := context.Background()

	// Validate session
	var session models.Session
	err := s.sessionsCollection.FindOne(ctx, bson.M{
		"session_token": langReq.SessionToken,
		"mobile_no":     langReq.MobileNo,
		"is_active":     true,
		"expires_at":    bson.M{"$gt": time.Now()},
	}).Decode(&session)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Update user language preferences
	update := bson.M{
		"$set": bson.M{
			"language_code":    langReq.LanguageCode,
			"language_name":    langReq.LanguageName,
			"region_code":      langReq.RegionCode,
			"timezone":         langReq.Timezone,
			"user_preferences": langReq.UserPreferences,
			"updated_at":       time.Now(),
		},
	}

	result, err := s.usersCollection.UpdateOne(
		ctx,
		bson.M{"mobile_no": langReq.MobileNo},
		update,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update language preferences: %v", err)
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("user not found")
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

	// Create context for database operations
	ctx := context.Background()

	// Validate session
	var session models.Session
	err := s.sessionsCollection.FindOne(ctx, bson.M{
		"session_token": actionReq.SessionToken,
		"mobile_no":     actionReq.PlayerID,
		"is_active":     true,
		"expires_at":    bson.M{"$gt": time.Now()},
	}).Decode(&session)

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
	ctx := context.Background()
	var session models.Session
	err := s.sessionsCollection.FindOne(ctx, bson.M{
		"session_token": sessionToken,
		"mobile_no":     mobileNo,
		"is_active":     true,
		"expires_at":    bson.M{"$gt": time.Now()},
	}).Decode(&session)

	return err == nil
}

// CleanupExpiredSessions removes expired sessions
func (s *SocketService) CleanupExpiredSessions() error {
	ctx := context.Background()
	_, err := s.sessionsCollection.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	return err
}

// HandleStaticMessage handles static message requests including game list
func (s *SocketService) HandleStaticMessage(staticReq models.StaticMessageRequest) (*models.StaticMessageResponse, error) {
	log.Printf("Static message request received for mobile: %s, type: %s", staticReq.MobileNo, staticReq.MessageType)

	ctx := context.Background()

	// Validate session
	var session models.Session
	err := s.sessionsCollection.FindOne(ctx, bson.M{
		"session_token": staticReq.SessionToken,
		"mobile_no":     staticReq.MobileNo,
		"is_active":     true,
		"expires_at":    bson.M{"$gt": time.Now()},
	}).Decode(&session)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Prepare response data based on message type
	var responseData map[string]interface{}

	switch staticReq.MessageType {
	case "game_list":
		responseData = s.getGameListData()
	default:
		return nil, fmt.Errorf("unknown message type: %s", staticReq.MessageType)
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
			"game_id":          "simple_1",
			"game_name":        "Simple Ludo",
			"game_type":        "board",
			"active_gamepalye": 1500,
			"livegameplaye":    800,
			"status":           "active",
			"created_at":       time.Now().UTC().Format(time.RFC3339),
		},
		{
			"game_id":          "2",
			"game_name":        "Test Rummy",
			"game_type":        "card",
			"active_gamepalye": 1000,
			"livegameplaye":    600,
			"status":           "active",
			"created_at":       time.Now().UTC().Format(time.RFC3339),
		},
		{
			"game_id":          "test_game_3",
			"game_name":        "Test Teen Patti",
			"game_type":        "card",
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

// getGameListData returns sample game list data with Redis caching
func (s *SocketService) getGameSubListData() map[string]interface{} {
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
			"active_gamepalye": 12313,
			"livegameplaye":    12313,
			"game name":        "newgame",
		},
		{
			"active_gamepalye": 12313,
			"livegameplaye":    12313,
			"game name":        "newgame",
		},
		{
			"active_gamepalye": 12313,
			"livegameplaye":    12313,
			"game name":        "newgame",
		},
		{
			"active_gamepalye": 12313,
			"livegameplaye":    12313,
			"game name":        "newgame",
		},
		{
			"active_gamepalye": 12313,
			"livegameplaye":    12313,
			"game name":        "newgame",
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

// HandleMainScreen handles main screen requests with authentication validation
func (s *SocketService) HandleMainScreen(mainReq models.MainScreenRequest) (*models.MainScreenResponse, error) {
	// Create context for database operations
	ctx := context.Background()

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
	err = s.usersCollection.FindOne(ctx, bson.M{"mobile_no": tokenMobileNo}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.sessionsCollection.FindOne(ctx, bson.M{
		"mobile_no":  tokenMobileNo,
		"device_id":  tokenDeviceID,
		"is_active":  true,
		"expires_at": bson.M{"$gt": time.Now()},
	}).Decode(&session)

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

	// Prepare response data based on message type
	var responseData map[string]interface{}

	switch mainReq.MessageType {
	case "game_list":
		responseData = s.getGameListData()
	case "sub_list":
		responseData = s.getGameSubListData()
	default:
		return nil, fmt.Errorf("unknown message type: %s", mainReq.MessageType)
	}

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
