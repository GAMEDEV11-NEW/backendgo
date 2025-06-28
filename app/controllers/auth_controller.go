package controllers

import (
	"gofiber/app/middlewares"
	"gofiber/app/models"
	"gofiber/app/services"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AuthController handles HTTP authentication endpoints
type AuthController struct {
	usersCollection    *mongo.Collection
	sessionsCollection *mongo.Collection
	socketService      *services.SocketService
}

// NewAuthController creates a new auth controller instance
func NewAuthController(usersCollection, sessionsCollection *mongo.Collection) *AuthController {
	socketService := services.NewSocketService(usersCollection, sessionsCollection)

	return &AuthController{
		usersCollection:    usersCollection,
		sessionsCollection: sessionsCollection,
		socketService:      socketService,
	}
}

// Login handles HTTP login requests
func (ac *AuthController) Login(c *fiber.Ctx) error {
	var loginReq models.LoginRequest

	if err := c.BodyParser(&loginReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if loginReq.MobileNo == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Mobile number is required",
		})
	}

	if loginReq.DeviceID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Device ID is required",
		})
	}

	if loginReq.FCMToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "FCM token is required",
		})
	}

	// Process login using socket service
	response, err := ac.socketService.HandleLogin(loginReq)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Login failed",
			"details": err.Error(),
		})
	}

	return c.JSON(response)
}

// VerifyOTP handles HTTP OTP verification requests
func (ac *AuthController) VerifyOTP(c *fiber.Ctx) error {
	var otpReq models.OTPVerificationRequest

	if err := c.BodyParser(&otpReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if otpReq.MobileNo == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Mobile number is required",
		})
	}

	if otpReq.SessionToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Session token is required",
		})
	}

	if otpReq.OTP == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "OTP is required",
		})
	}

	// Process OTP verification using socket service
	response, err := ac.socketService.HandleOTPVerification(otpReq)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "OTP verification failed",
			"details": err.Error(),
		})
	}

	return c.JSON(response)
}

// SetProfile handles HTTP profile setup requests
func (ac *AuthController) SetProfile(c *fiber.Ctx) error {
	var profileReq models.SetProfileRequest

	if err := c.BodyParser(&profileReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if profileReq.MobileNo == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Mobile number is required",
		})
	}

	if profileReq.SessionToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Session token is required",
		})
	}

	if profileReq.FullName == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Full name is required",
		})
	}

	// Process profile setup using socket service
	response, err := ac.socketService.HandleSetProfile(profileReq)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Profile setup failed",
			"details": err.Error(),
		})
	}

	return c.JSON(response)
}

// SetLanguage handles HTTP language setup requests
func (ac *AuthController) SetLanguage(c *fiber.Ctx) error {
	var langReq models.SetLanguageRequest

	if err := c.BodyParser(&langReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if langReq.MobileNo == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Mobile number is required",
		})
	}

	if langReq.SessionToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Session token is required",
		})
	}

	if langReq.LanguageCode == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Language code is required",
		})
	}

	// Process language setup using socket service
	response, err := ac.socketService.HandleSetLanguage(langReq)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Language setup failed",
			"details": err.Error(),
		})
	}

	return c.JSON(response)
}

// MainScreen handles HTTP main screen requests (protected)
func (ac *AuthController) MainScreen(c *fiber.Ctx) error {
	var mainReq models.MainScreenRequest

	if err := c.BodyParser(&mainReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if mainReq.MobileNo == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Mobile number is required",
		})
	}

	if mainReq.JWTToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "JWT token is required",
		})
	}

	if mainReq.DeviceID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Device ID is required",
		})
	}

	if mainReq.MessageType == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Message type is required",
		})
	}

	// Process main screen request using socket service
	response, err := ac.socketService.HandleMainScreen(mainReq)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Main screen request failed",
			"details": err.Error(),
		})
	}

	return c.JSON(response)
}

// StaticMessage handles HTTP static message requests (protected)
func (ac *AuthController) StaticMessage(c *fiber.Ctx) error {
	var staticReq models.StaticMessageRequest

	if err := c.BodyParser(&staticReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if staticReq.MobileNo == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Mobile number is required",
		})
	}

	if staticReq.SessionToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Session token is required",
		})
	}

	if staticReq.MessageType == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Message type is required",
		})
	}

	// Process static message request using socket service
	response, err := ac.socketService.HandleStaticMessage(staticReq)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Static message request failed",
			"details": err.Error(),
		})
	}

	return c.JSON(response)
}

// Logout handles HTTP logout requests (protected)
func (ac *AuthController) Logout(c *fiber.Ctx) error {
	// Get user from context (set by middleware)
	user, err := middlewares.GetUserFromContext(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "User not authenticated",
		})
	}

	// Get session from context
	session, err := middlewares.GetSessionFromContext(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Session not found",
		})
	}

	// Deactivate session
	_, err = ac.sessionsCollection.UpdateOne(
		c.Context(),
		bson.M{"_id": session.ID},
		bson.M{"$set": bson.M{"is_active": false}},
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to logout",
		})
	}

	return c.JSON(fiber.Map{
		"status":    "success",
		"message":   "Logout successful",
		"mobile_no": user.MobileNo,
	})
}

// GetProfile handles HTTP profile retrieval requests (protected)
func (ac *AuthController) GetProfile(c *fiber.Ctx) error {
	// Get user from context (set by middleware)
	user, err := middlewares.GetUserFromContext(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "User not authenticated",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Profile retrieved successfully",
		"profile": user,
	})
}

// UpdateProfile handles HTTP profile update requests (protected)
func (ac *AuthController) UpdateProfile(c *fiber.Ctx) error {
	// Get user from context (set by middleware)
	user, err := middlewares.GetUserFromContext(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "User not authenticated",
		})
	}

	var updateData map[string]interface{}
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Update user profile
	_, err = ac.usersCollection.UpdateOne(
		c.Context(),
		bson.M{"mobile_no": user.MobileNo},
		bson.M{"$set": updateData},
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update profile",
		})
	}

	return c.JSON(fiber.Map{
		"status":    "success",
		"message":   "Profile updated successfully",
		"mobile_no": user.MobileNo,
	})
}
