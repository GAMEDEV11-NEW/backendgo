// app/middlewares/middleware_example.go
package middlewares

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"gofiber/app/models"
	"gofiber/app/utils"
)

// JWTMiddleware validates JWT tokens using encrypted validation
func JWTMiddleware(usersCollection, sessionsCollection *mongo.Collection) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"status":  "error",
				"message": "Authorization header is required",
			})
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid authorization header format",
			})
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Get mobile number from request (could be in body, query, or header)
		mobileNo := c.Query("mobile_no")
		if mobileNo == "" {
			mobileNo = c.Get("X-Mobile-No")
		}
		if mobileNo == "" {
			// Try to get from request body
			var body map[string]interface{}
			if err := c.BodyParser(&body); err == nil {
				if mobile, ok := body["mobile_no"].(string); ok {
					mobileNo = mobile
				}
			}
		}

		// Validate JWT token with mobile number matching
		if mobileNo != "" {
			err := utils.ValidateMobileNumberInToken(tokenString, mobileNo)
			if err != nil {
				log.Printf("JWT validation failed for mobile %s: %v", mobileNo, err)
				return c.Status(401).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid JWT token",
					"details": err.Error(),
				})
			}
		}

		// Validate encrypted JWT token
		encryptedJWTData, err := utils.ValidateEncryptedJWTToken(tokenString)
		if err != nil {
			log.Printf("Encrypted JWT validation failed: %v", err)
			return c.Status(401).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid encrypted JWT token",
				"details": err.Error(),
			})
		}

		// Verify session exists and is active
		ctx := context.Background()
		var session models.Session
		err = sessionsCollection.FindOne(ctx, bson.M{
			"mobile_no":  encryptedJWTData.MobileNo,
			"device_id":  encryptedJWTData.DeviceID,
			"is_active":  true,
			"expires_at": bson.M{"$gt": time.Now()},
		}).Decode(&session)

		if err != nil {
			log.Printf("Session validation failed for mobile %s: %v", encryptedJWTData.MobileNo, err)
			return c.Status(401).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid or expired session",
			})
		}

		// Verify JWT token matches stored token
		if session.JWTToken != tokenString {
			log.Printf("JWT token mismatch for mobile %s", encryptedJWTData.MobileNo)
			return c.Status(401).JSON(fiber.Map{
				"status":  "error",
				"message": "JWT token mismatch with stored token",
			})
		}

		// Get user information
		var user models.User
		err = usersCollection.FindOne(ctx, bson.M{"mobile_no": encryptedJWTData.MobileNo}).Decode(&user)
		if err != nil {
			log.Printf("User not found for mobile %s: %v", encryptedJWTData.MobileNo, err)
			return c.Status(401).JSON(fiber.Map{
				"status":  "error",
				"message": "User not found",
			})
		}

		// Store user and session data in context for later use
		c.Locals("user", user)
		c.Locals("session", session)
		c.Locals("encrypted_jwt_data", encryptedJWTData)

		log.Printf("✅ JWT middleware validation successful for user: %s", encryptedJWTData.MobileNo)

		return c.Next()
	}
}

// OptionalJWTMiddleware validates JWT tokens but doesn't require them (for optional endpoints)
func OptionalJWTMiddleware(usersCollection, sessionsCollection *mongo.Collection) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			return c.Next()
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			// Invalid format, continue without authentication
			return c.Next()
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Try to validate the token
		encryptedJWTData, err := utils.ValidateEncryptedJWTToken(tokenString)
		if err != nil {
			// Token is invalid, continue without authentication
			log.Printf("Optional JWT validation failed: %v", err)
			return c.Next()
		}

		// Token is valid, verify session
		ctx := context.Background()
		var session models.Session
		err = sessionsCollection.FindOne(ctx, bson.M{
			"mobile_no":  encryptedJWTData.MobileNo,
			"device_id":  encryptedJWTData.DeviceID,
			"is_active":  true,
			"expires_at": bson.M{"$gt": time.Now()},
		}).Decode(&session)

		if err != nil || session.JWTToken != tokenString {
			// Session is invalid, continue without authentication
			return c.Next()
		}

		// Get user information
		var user models.User
		err = usersCollection.FindOne(ctx, bson.M{"mobile_no": encryptedJWTData.MobileNo}).Decode(&user)
		if err != nil {
			// User not found, continue without authentication
			return c.Next()
		}

		// Store user and session data in context
		c.Locals("user", user)
		c.Locals("session", session)
		c.Locals("encrypted_jwt_data", encryptedJWTData)

		log.Printf("✅ Optional JWT middleware validation successful for user: %s", encryptedJWTData.MobileNo)

		return c.Next()
	}
}

// GetUserFromContext retrieves user data from context
func GetUserFromContext(c *fiber.Ctx) (*models.User, error) {
	user, ok := c.Locals("user").(models.User)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}
	return &user, nil
}

// GetSessionFromContext retrieves session data from context
func GetSessionFromContext(c *fiber.Ctx) (*models.Session, error) {
	session, ok := c.Locals("session").(models.Session)
	if !ok {
		return nil, fmt.Errorf("session not found in context")
	}
	return &session, nil
}

// GetEncryptedJWTDataFromContext retrieves encrypted JWT data from context
func GetEncryptedJWTDataFromContext(c *fiber.Ctx) (*utils.EncryptedJWTData, error) {
	encryptedJWTData, ok := c.Locals("encrypted_jwt_data").(*utils.EncryptedJWTData)
	if !ok {
		return nil, fmt.Errorf("encrypted JWT data not found in context")
	}
	return encryptedJWTData, nil
}
