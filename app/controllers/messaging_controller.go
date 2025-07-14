package controllers

import (
	"gofiber/app/services"
	"time"

	"github.com/gofiber/fiber/v2"
)

// MessagingController handles HTTP endpoints for server-to-client messaging
type MessagingController struct {
	messagingService *services.MessagingService
}

// NewMessagingController creates a new messaging controller instance
func NewMessagingController(messagingService *services.MessagingService) *MessagingController {
	return &MessagingController{
		messagingService: messagingService,
	}
}

// SendNotificationRequest represents a notification request
type SendNotificationRequest struct {
	TargetType  string                 `json:"target_type"`  // user, mobile, all, socket
	TargetValue string                 `json:"target_value"` // user_id, mobile_no, socket_id
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// SendGameUpdateRequest represents a game update request
type SendGameUpdateRequest struct {
	UserID   string                 `json:"user_id"`
	GameData map[string]interface{} `json:"game_data"`
}

// SendContestUpdateRequest represents a contest update request
type SendContestUpdateRequest struct {
	UserID      string                 `json:"user_id"`
	ContestData map[string]interface{} `json:"contest_data"`
}

// SendSystemAlertRequest represents a system alert request
type SendSystemAlertRequest struct {
	AlertType string `json:"alert_type"`
	Message   string `json:"message"`
	Severity  string `json:"severity"` // info, warning, error
}

// SendNotification sends a notification to a specific target
func (c *MessagingController) SendNotification(ctx *fiber.Ctx) error {
	var req SendNotificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate required fields
	if req.TargetType == "" || req.TargetValue == "" || req.Title == "" || req.Body == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing required fields: target_type, target_value, title, body",
		})
	}

	// Send notification
	err := c.messagingService.SendNotification(req.TargetType, req.TargetValue, req.Title, req.Body, req.Data)
	if err != nil {

		return ctx.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to send notification",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":    "success",
		"message":   "Notification sent successfully",
		"target":    req.TargetType + ":" + req.TargetValue,
		"title":     req.Title,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// SendGameUpdate sends a game update to a specific user
func (c *MessagingController) SendGameUpdate(ctx *fiber.Ctx) error {
	var req SendGameUpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate required fields
	if req.UserID == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing required field: user_id",
		})
	}

	// Send game update
	err := c.messagingService.SendGameUpdate(req.UserID, req.GameData)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to send game update",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":    "success",
		"message":   "Game update sent successfully",
		"user_id":   req.UserID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// SendContestUpdate sends a contest update to a specific user
func (c *MessagingController) SendContestUpdate(ctx *fiber.Ctx) error {
	var req SendContestUpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate required fields
	if req.UserID == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing required field: user_id",
		})
	}

	// Send contest update
	err := c.messagingService.SendContestUpdate(req.UserID, req.ContestData)
	if err != nil {

		return ctx.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to send contest update",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":    "success",
		"message":   "Contest update sent successfully",
		"user_id":   req.UserID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// SendSystemAlert sends a system alert to all users
func (c *MessagingController) SendSystemAlert(ctx *fiber.Ctx) error {
	var req SendSystemAlertRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate required fields
	if req.AlertType == "" || req.Message == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing required fields: alert_type, message",
		})
	}

	// Set default severity if not provided
	if req.Severity == "" {
		req.Severity = "info"
	}

	// Send system alert
	err := c.messagingService.SendSystemAlert(req.AlertType, req.Message, req.Severity)
	if err != nil {

		return ctx.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to send system alert",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":     "success",
		"message":    "System alert sent successfully",
		"alert_type": req.AlertType,
		"severity":   req.Severity,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

// GetConnectionStats returns connection statistics
func (c *MessagingController) GetConnectionStats(ctx *fiber.Ctx) error {
	// Get active connections count
	count, err := c.messagingService.GetActiveConnectionsCount()
	if err != nil {

		return ctx.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to get connection statistics",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":             "success",
		"active_connections": count,
		"timestamp":          time.Now().UTC().Format(time.RFC3339),
	})
}

// GetUserConnections returns all active connections for a specific user
func (c *MessagingController) GetUserConnections(ctx *fiber.Ctx) error {
	userID := ctx.Params("user_id")
	if userID == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing user_id parameter",
		})
	}

	connections, err := c.messagingService.GetUserConnections(userID)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to get user connections",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":      "success",
		"user_id":     userID,
		"connections": connections,
		"count":       len(connections),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}

// GetMobileConnections returns all active connections for a specific mobile number
func (c *MessagingController) GetMobileConnections(ctx *fiber.Ctx) error {
	mobileNo := ctx.Params("mobile_no")
	if mobileNo == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing mobile_no parameter",
		})
	}

	connections, err := c.messagingService.GetMobileConnections(mobileNo)
	if err != nil {

		return ctx.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to get mobile connections",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":      "success",
		"mobile_no":   mobileNo,
		"connections": connections,
		"count":       len(connections),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}

// GetUserSessionStatus returns the session status for a specific user
func (c *MessagingController) GetUserSessionStatus(ctx *fiber.Ctx) error {
	mobileNo := ctx.Params("mobile_no")
	if mobileNo == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing mobile_no parameter",
		})
	}

	// Get active session count
	count, err := c.messagingService.GetSessionService().GetActiveSessionCount(mobileNo)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to get session status",
			"error":   err.Error(),
		})
	}

	// Get active session details
	activeSession, err := c.messagingService.GetSessionService().GetActiveSessionForUser(mobileNo)
	hasActiveSession := err == nil

	return ctx.JSON(fiber.Map{
		"status":                  "success",
		"mobile_no":               mobileNo,
		"has_active_session":      hasActiveSession,
		"active_sessions_count":   count,
		"single_session_enforced": true,
		"active_session":          activeSession,
		"timestamp":               time.Now().UTC().Format(time.RFC3339),
	})
}

// ForceLogoutUser deactivates all sessions for a user (force logout)
func (c *MessagingController) ForceLogoutUser(ctx *fiber.Ctx) error {
	mobileNo := ctx.Params("mobile_no")
	if mobileNo == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing mobile_no parameter",
		})
	}

	// Deactivate all existing sessions
	err := c.messagingService.GetSessionService().DeactivateExistingSessions(mobileNo)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to force logout user",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":    "success",
		"message":   "User force logged out successfully",
		"mobile_no": mobileNo,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
