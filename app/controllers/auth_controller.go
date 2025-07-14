// app/controllers/auth_controller.go
package controllers

import (
	"context"
	"gofiber/app/middlewares"
	"gofiber/app/models"
	"gofiber/app/services"
	"gofiber/app/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AuthController handles authentication-related HTTP requests
type AuthController struct {
	usersCollection    *mongo.Collection
	sessionsCollection *mongo.Collection
	authService        *services.AuthService
}

// NewAuthController creates a new auth controller instance
func NewAuthController(usersCollection, sessionsCollection *mongo.Collection) *AuthController {
	return &AuthController{
		usersCollection:    usersCollection,
		sessionsCollection: sessionsCollection,
	}
}

// SetAuthService sets the auth service for the controller
func (c *AuthController) SetAuthService(authService *services.AuthService) {
	c.authService = authService
}

// Login handles user login requests
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	var loginReq models.LoginRequest
	if err := ctx.BodyParser(&loginReq); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
		})
	}

	// Validate required fields
	if loginReq.MobileNo == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Mobile number is required",
		})
	}

	if loginReq.DeviceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Device ID is required",
		})
	}

	if loginReq.FCMToken == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "FCM token is required",
		})
	}

	// Generate OTP
	otp := c.authService.GenerateOTP()

	// Store OTP in database
	_, err := c.authService.GetLatestOTP(loginReq.MobileNo, "login")
	if err != nil {
		// Store new OTP using auth service's cassandra session
		err = c.authService.GetCassandraSession().Query(`
			INSERT INTO otp_store (phone_or_email, otp_code, purpose, created_at, expires_at, is_verified, attempt_count)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`).Bind(loginReq.MobileNo, strconv.Itoa(otp), "login", time.Now().Format(time.RFC3339), time.Now().Add(5*time.Minute).Format(time.RFC3339), false, 0).Exec()
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to generate OTP",
			})
		}
	}

	// Generate session token
	sessionToken, err := c.authService.GenerateSessionToken()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate session token",
		})
	}

	// Create session
	session := models.Session{
		SessionToken: sessionToken,
		MobileNo:     loginReq.MobileNo,
		DeviceID:     loginReq.DeviceID,
		FCMToken:     loginReq.FCMToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		IsActive:     true,
	}

	// Store session in database
	_, err = c.sessionsCollection.InsertOne(context.Background(), session)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create session",
		})
	}

	// Check if user exists
	var existingUser models.User
	err = c.usersCollection.FindOne(context.Background(), bson.M{"mobile_no": loginReq.MobileNo}).Decode(&existingUser)
	isNewUser := err == mongo.ErrNoDocuments

	// If user doesn't exist, create new user
	if isNewUser {
		newUser := models.User{
			ID:           uuid.New().String(),
			MobileNo:     loginReq.MobileNo,
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LanguageCode: "en",
		}

		_, err = c.usersCollection.InsertOne(context.Background(), newUser)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to create user",
			})
		}
	}

	// Return response
	response := models.LoginResponse{
		Status:       "success",
		Message:      "OTP sent successfully",
		MobileNo:     loginReq.MobileNo,
		SessionToken: sessionToken,
		OTP:          otp, // In production, this should not be returned
		IsNewUser:    isNewUser,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}

	return ctx.JSON(response)
}

// VerifyOTP handles OTP verification
func (c *AuthController) VerifyOTP(ctx *fiber.Ctx) error {
	var otpReq models.OTPVerificationRequest
	if err := ctx.BodyParser(&otpReq); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
		})
	}

	// Validate required fields
	if otpReq.MobileNo == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Mobile number is required",
		})
	}

	if otpReq.SessionToken == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Session token is required",
		})
	}

	if otpReq.OTP == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "OTP is required",
		})
	}

	// Verify OTP
	otpData, err := c.authService.GetLatestOTP(otpReq.MobileNo, "login")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid OTP",
		})
	}

	// Check if OTP is expired
	expiresAt, err := time.Parse(time.RFC3339, otpData.ExpiresAt)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid OTP expiry format",
		})
	}

	if time.Now().After(expiresAt) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "OTP has expired",
		})
	}

	// Check if OTP matches
	if otpData.OTPCode != otpReq.OTP {
		// Increment attempt count - update in database
		err = c.authService.GetCassandraSession().Query(`
			UPDATE otp_store
			SET attempt_count = attempt_count + 1
			WHERE phone_or_email = ? AND purpose = ?
		`).Bind(otpReq.MobileNo, "login").Exec()
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to update OTP attempt count",
			})
		}

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid OTP",
		})
	}

	// Mark OTP as verified
	err = c.authService.GetCassandraSession().Query(`
		UPDATE otp_store
		SET is_verified = true
		WHERE phone_or_email = ? AND purpose = ?
	`).Bind(otpReq.MobileNo, "login").Exec()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to mark OTP as verified",
		})
	}

	// Get session data from MongoDB
	var sessionData models.Session
	err = c.sessionsCollection.FindOne(context.Background(), bson.M{"session_token": otpReq.SessionToken}).Decode(&sessionData)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid session token",
		})
	}

	// Generate JWT token
	jwtToken, err := utils.GenerateSimpleJWTToken(otpReq.MobileNo, sessionData.DeviceID, sessionData.FCMToken)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate JWT token",
		})
	}

	// Update session with JWT token - using direct database update
	_, err = c.sessionsCollection.UpdateOne(
		context.Background(),
		bson.M{"session_token": otpReq.SessionToken},
		bson.M{"$set": bson.M{"jwt_token": jwtToken}},
	)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update session",
		})
	}

	// Return response
	response := models.OTPVerificationResponse{
		Status:       "success",
		Message:      "OTP verified successfully",
		MobileNo:     otpReq.MobileNo,
		SessionToken: otpReq.SessionToken,
		JWTToken:     jwtToken,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}

	return ctx.JSON(response)
}

// SetProfile handles user profile setup
func (c *AuthController) SetProfile(ctx *fiber.Ctx) error {
	// Get user from context (set by middleware)
	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User not authenticated",
		})
	}

	var profileReq models.SetProfileRequest
	if err := ctx.BodyParser(&profileReq); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
		})
	}

	// Update user profile
	update := bson.M{
		"$set": bson.M{
			"full_name":  profileReq.FullName,
			"state":      profileReq.State,
			"updated_at": time.Now(),
		},
	}

	_, err = c.usersCollection.UpdateOne(
		context.Background(),
		bson.M{"mobile_no": user.MobileNo},
		update,
	)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update profile",
		})
	}

	response := models.SetProfileResponse{
		Status:    "success",
		Message:   "Profile updated successfully",
		MobileNo:  user.MobileNo,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	return ctx.JSON(response)
}

// SetLanguage handles language preference setup
func (c *AuthController) SetLanguage(ctx *fiber.Ctx) error {
	// Get user from context (set by middleware)
	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User not authenticated",
		})
	}

	var langReq models.SetLanguageRequest
	if err := ctx.BodyParser(&langReq); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
		})
	}

	// Update user language preference
	update := bson.M{
		"$set": bson.M{
			"language_code": langReq.LanguageCode,
			"updated_at":    time.Now(),
		},
	}

	_, err = c.usersCollection.UpdateOne(
		context.Background(),
		bson.M{"mobile_no": user.MobileNo},
		update,
	)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update language preference",
		})
	}

	response := models.SetLanguageResponse{
		Status:       "success",
		Message:      "Language preference updated successfully",
		MobileNo:     user.MobileNo,
		LanguageCode: langReq.LanguageCode,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}

	return ctx.JSON(response)
}

// Logout handles user logout
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	// Get session from context (set by middleware)
	session, err := middlewares.GetSessionFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User not authenticated",
		})
	}

	// Deactivate session
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	_, err = c.sessionsCollection.UpdateOne(
		context.Background(),
		bson.M{"session_token": session.SessionToken},
		update,
	)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to logout",
		})
	}

	response := fiber.Map{
		"status":    "success",
		"message":   "Logged out successfully",
		"mobile_no": session.MobileNo,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	return ctx.JSON(response)
}

// GetProfile handles getting user profile
func (c *AuthController) GetProfile(ctx *fiber.Ctx) error {
	// Get user from context (set by middleware)
	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User not authenticated",
		})
	}

	response := fiber.Map{
		"status":    "success",
		"message":   "Profile retrieved successfully",
		"user":      user,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	return ctx.JSON(response)
}
