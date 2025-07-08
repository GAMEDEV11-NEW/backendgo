package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gofiber/app/models"
	"gofiber/app/utils"
	"gofiber/redis"
	"math/big"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

// AuthService handles all authentication and user management business logic
type AuthService struct {
	cassandraSession *gocql.Session
	redisService     *redis.Service
	sessionService   *SessionService
}

// NewAuthService creates a new auth service instance using Cassandra
func NewAuthService(cassandraSession *gocql.Session) *AuthService {
	if cassandraSession == nil {
		panic("Cassandra session cannot be nil")
	}
	redisService := redis.NewService()
	service := &AuthService{
		cassandraSession: cassandraSession,
		redisService:     redisService,
		sessionService:   NewSessionService(cassandraSession),
	}
	return service
}

// GenerateSessionToken generates a unique session token
func (s *AuthService) GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateOTP generates a 6-digit OTP
func (s *AuthService) GenerateOTP() int {
	// Generate a random number between 100000 and 999999
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return int(n.Int64()) + 100000
}

// HandleDeviceInfo processes device information from client
func (s *AuthService) HandleDeviceInfo(deviceInfo models.DeviceInfo, socketID string) models.DeviceInfoResponse {
	return models.DeviceInfoResponse{
		Status:    "success",
		Message:   "Device info received and validated",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SocketID:  socketID,
		Event:     "device:info:ack",
	}
}

// HandleLogin handles user login and OTP generation
func (s *AuthService) HandleLogin(loginReq models.LoginRequest) (*models.LoginResponse, error) {
	// Validate mobile number format
	if len(loginReq.MobileNo) != 10 {
		return nil, fmt.Errorf("invalid mobile number format")
	}

	// Check if user exists
	var existingUser models.User
	err := s.cassandraSession.Query(`
		SELECT id, mobile_no, status, created_at
		FROM users
		WHERE mobile_no = ?
		ALLOW FILTERING
	`).Bind(loginReq.MobileNo).Scan(&existingUser.ID, &existingUser.MobileNo, &existingUser.Status, &existingUser.CreatedAt)

	userID := existingUser.ID

	if err != nil {
		// Create new user
		userID = uuid.New().String()
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

	// Create session using SessionService (Redis + Cassandra)
	sessionToken := uuid.New().String()
	sessionExpiry := time.Now().Add(24 * time.Hour)

	sessionData := SessionData{
		SessionToken: sessionToken,
		MobileNo:     loginReq.MobileNo,
		UserID:       userID,
		DeviceID:     loginReq.DeviceID,
		FCMToken:     loginReq.FCMToken,
		JWTToken:     "", // Will be set after OTP verification
		SocketID:     loginReq.SocketID,
		IsActive:     true,
		CreatedAt:    time.Now(),
		ExpiresAt:    sessionExpiry,
		UserStatus:   existingUser.Status,
	}

	err = s.sessionService.CreateSession(sessionData)
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
func (s *AuthService) HandleOTPVerification(otpReq models.OTPVerificationRequest) (*models.OTPVerificationResponse, error) {
	// Validate session using SessionService (Redis + Cassandra)
	sessionData, err := s.sessionService.GetSession(otpReq.SessionToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session: %v", err)
	}

	// Verify mobile number matches
	if sessionData.MobileNo != otpReq.MobileNo {
		return nil, fmt.Errorf("session mobile number mismatch")
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
	jwtToken, err := utils.GenerateSimpleJWTToken(otpReq.MobileNo, sessionData.DeviceID, sessionData.FCMToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT token: %v", err)
	}

	// Update session with JWT token using SessionService
	updates := map[string]interface{}{
		"jwt_token":   jwtToken,
		"user_status": userStatus,
	}

	err = s.sessionService.UpdateSession(otpReq.SessionToken, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update session with JWT token: %v", err)
	}

	return &models.OTPVerificationResponse{
		Status:       "success",
		Message:      "OTP verified successfully",
		MobileNo:     otpReq.MobileNo,
		DeviceID:     sessionData.DeviceID,
		SessionToken: otpReq.SessionToken,
		JWTToken:     jwtToken,
		UserStatus:   userStatus,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		SocketID:     sessionData.SocketID,
		Event:        "otp:verified",
	}, nil
}

// HandleSetProfile sets up user profile
func (s *AuthService) HandleSetProfile(profileReq models.SetProfileRequest) (*models.SetProfileResponse, error) {
	// Validate session using SessionService (Redis + Cassandra)
	sessionData, err := s.sessionService.GetSession(profileReq.SessionToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session: %v", err)
	}

	// Verify mobile number matches
	if sessionData.MobileNo != profileReq.MobileNo {
		return nil, fmt.Errorf("session mobile number mismatch")
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
func (s *AuthService) HandleSetLanguage(langReq models.SetLanguageRequest) (*models.SetLanguageResponse, error) {
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

// ValidateSession validates if a session is active and not expired
func (s *AuthService) ValidateSession(sessionToken, mobileNo string) bool {
	var session models.Session
	err := s.cassandraSession.Query(`
		SELECT session_token, mobile_no, is_active, expires_at
		FROM sessions
		WHERE session_token = ? AND mobile_no = ? AND is_active = true AND expires_at > ?
	`).Bind(sessionToken, mobileNo, true, time.Now()).Scan(&session.SessionToken, &session.MobileNo, &session.IsActive, &session.ExpiresAt)

	return err == nil
}

// CleanupExpiredSessions removes expired sessions
func (s *AuthService) CleanupExpiredSessions() error {
	err := s.cassandraSession.Query(`
		DELETE FROM sessions
		WHERE expires_at < ?
	`).Bind(time.Now()).Exec()
	return err
}

// CleanupExpiredOTPs removes expired OTPs from the database
func (s *AuthService) CleanupExpiredOTPs() error {
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
func (s *AuthService) GetLatestOTP(phoneOrEmail, purpose string) (*models.OTPData, error) {
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
func (s *AuthService) ResendOTP(mobileNo string) (int, error) {
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

// GetCassandraSession returns the Cassandra session for external access
func (s *AuthService) GetCassandraSession() *gocql.Session {
	return s.cassandraSession
}
