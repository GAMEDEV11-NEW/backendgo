package config

import (
	"encoding/json"
	"gofiber/app/models"
	"gofiber/app/services"
	"log"
	"time"

	socketio "github.com/doquangtan/socket.io/v4"
	"github.com/gofiber/fiber/v2"
)

// SocketIoHandler handles all Socket.IO related functionality
type SocketIoHandler struct {
	io            *socketio.Io
	socketService *services.SocketService
}

// NewSocketHandler creates a new Socket.IO handler instance
func NewSocketHandler(socketService *services.SocketService) *SocketIoHandler {
	io := socketio.New()

	handler := &SocketIoHandler{
		io:            io,
		socketService: socketService,
	}

	handler.setupSocketHandlers()
	return handler
}

// setupSocketHandlers configures all Socket.IO event handlers
func (h *SocketIoHandler) setupSocketHandlers() {
	// Authorization handler
	h.io.OnAuthorization(func(params map[string]string) bool {
		log.Printf("Authorization attempt with params: %v", params)
		// For now, allow all connections
		// In production, you would validate tokens here
		return true
	})

	// Main connection handler
	h.io.OnConnection(func(socket *socketio.Socket) {
		log.Printf("✅ Socket connected: %s (namespace: %s)", socket.Id, socket.Nps)

		// Send welcome message
		welcome := h.socketService.HandleWelcome()
		socket.Emit("connect_response", welcome)

		// Send connect response with token
		connectResp := models.ConnectResponse{
			Token:     123456, // Generate random token
			Message:   "Welcome to the Game Admin Server!",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			SocketID:  socket.Id,
			Status:    "connected",
			Event:     "connect",
		}
		socket.Emit("connect", connectResp)

		// Device info handler
		socket.On("device:info", func(event *socketio.EventPayload) {

			if len(event.Data) == 0 {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "device_info",
					Message:   "No device info provided",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Parse device info
			deviceInfoData, ok := event.Data[0].(map[string]interface{})
			if !ok {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "device_info",
					Message:   "Invalid device info format",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Convert to DeviceInfo struct
			deviceInfoJSON, _ := json.Marshal(deviceInfoData)
			var deviceInfo models.DeviceInfo
			if err := json.Unmarshal(deviceInfoJSON, &deviceInfo); err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "device_info",
					Message:   "Failed to parse device info",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Process device info
			response := h.socketService.HandleDeviceInfo(deviceInfo, socket.Id)
			socket.Emit("device:info:ack", response)
		})

		// Login handler
		socket.On("login", func(event *socketio.EventPayload) {

			if len(event.Data) == 0 {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "login_data",
					Message:   "No login data provided",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Parse login request
			loginData, ok := event.Data[0].(map[string]interface{})
			if !ok {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "login_data",
					Message:   "Invalid login data format",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Convert to LoginRequest struct
			loginJSON, _ := json.Marshal(loginData)
			var loginReq models.LoginRequest
			if err := json.Unmarshal(loginJSON, &loginReq); err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "login_data",
					Message:   "Failed to parse login data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Process login
			response, err := h.socketService.HandleLogin(loginReq)
			if err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeVerificationError,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "login",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Set socket ID in response
			response.SocketID = socket.Id
			socket.Emit("otp:sent", response)
		})

		// OTP verification handler
		socket.On("verify:otp", func(event *socketio.EventPayload) {

			if len(event.Data) == 0 {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "otp_data",
					Message:   "No OTP data provided",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Parse OTP request
			otpData, ok := event.Data[0].(map[string]interface{})
			if !ok {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "otp_data",
					Message:   "Invalid OTP data format",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Convert to OTPVerificationRequest struct
			otpJSON, _ := json.Marshal(otpData)
			var otpReq models.OTPVerificationRequest
			if err := json.Unmarshal(otpJSON, &otpReq); err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "otp_data",
					Message:   "Failed to parse OTP data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Process OTP verification
			response, err := h.socketService.HandleOTPVerification(otpReq)
			if err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidOTP,
					ErrorType: models.ErrorTypeOTP,
					Field:     "otp",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Set socket ID in response
			response.SocketID = socket.Id
			socket.Emit("otp:verified", response)
		})

		// Set profile handler
		socket.On("set:profile", func(event *socketio.EventPayload) {

			if len(event.Data) == 0 {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "profile_data",
					Message:   "No profile data provided",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Parse profile request
			profileData, ok := event.Data[0].(map[string]interface{})
			if !ok {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "profile_data",
					Message:   "Invalid profile data format",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Convert to SetProfileRequest struct
			profileJSON, _ := json.Marshal(profileData)
			var profileReq models.SetProfileRequest
			if err := json.Unmarshal(profileJSON, &profileReq); err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "profile_data",
					Message:   "Failed to parse profile data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Process profile setup
			response, err := h.socketService.HandleSetProfile(profileReq)
			if err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeVerificationError,
					ErrorType: models.ErrorTypeValidation,
					Field:     "profile",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Set socket ID in response
			response.SocketID = socket.Id
			socket.Emit("profile:set", response)
		})

		// Set language handler
		socket.On("set:language", func(event *socketio.EventPayload) {

			if len(event.Data) == 0 {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "language_data",
					Message:   "No language data provided",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Parse language request
			langData, ok := event.Data[0].(map[string]interface{})
			if !ok {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "language_data",
					Message:   "Invalid language data format",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Convert to SetLanguageRequest struct
			langJSON, _ := json.Marshal(langData)
			var langReq models.SetLanguageRequest
			if err := json.Unmarshal(langJSON, &langReq); err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "language_data",
					Message:   "Failed to parse language data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Process language setup
			response, err := h.socketService.HandleSetLanguage(langReq)
			if err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeVerificationError,
					ErrorType: models.ErrorTypeValidation,
					Field:     "language",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Set socket ID in response
			response.SocketID = socket.Id
			socket.Emit("language:set", response)
		})

		// Set profile handler
		socket.On("set:profile", func(event *socketio.EventPayload) {

			if len(event.Data) == 0 {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "profile_data",
					Message:   "No profile data provided",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Parse profile request
			profileData, ok := event.Data[0].(map[string]interface{})
			if !ok {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "profile_data",
					Message:   "Invalid profile data format",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Convert to SetProfileRequest struct
			profileJSON, _ := json.Marshal(profileData)
			var profileReq models.SetProfileRequest
			if err := json.Unmarshal(profileJSON, &profileReq); err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "profile_data",
					Message:   "Failed to parse profile data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Process profile setup
			response, err := h.socketService.HandleSetProfile(profileReq)
			if err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeVerificationError,
					ErrorType: models.ErrorTypeValidation,
					Field:     "profile",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Set socket ID in response
			response.SocketID = socket.Id
			socket.Emit("profile:set", response)
		})

		// Set language handler
		socket.On("set:language", func(event *socketio.EventPayload) {

			if len(event.Data) == 0 {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "language_data",
					Message:   "No language data provided",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Parse language request
			langData, ok := event.Data[0].(map[string]interface{})
			if !ok {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "language_data",
					Message:   "Invalid language data format",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Convert to SetLanguageRequest struct
			langJSON, _ := json.Marshal(langData)
			var langReq models.SetLanguageRequest
			if err := json.Unmarshal(langJSON, &langReq); err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "language_data",
					Message:   "Failed to parse language data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Process language setup
			response, err := h.socketService.HandleSetLanguage(langReq)
			if err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeVerificationError,
					ErrorType: models.ErrorTypeValidation,
					Field:     "language",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Set socket ID in response
			response.SocketID = socket.Id
			socket.Emit("language:set", response)
		})

		// Static message handler
		socket.On("main:screen", func(event *socketio.EventPayload) {

			if len(event.Data) == 0 {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "main_screen_data",
					Message:   "No main screen data provided",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Parse main screen request
			mainData, ok := event.Data[0].(map[string]interface{})
			if !ok {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "main_screen_data",
					Message:   "Invalid main screen data format",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Convert to MainScreenRequest struct
			mainJSON, _ := json.Marshal(mainData)
			var mainReq models.MainScreenRequest
			if err := json.Unmarshal(mainJSON, &mainReq); err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "main_screen_data",
					Message:   "Failed to parse main screen data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Validate required fields
			if mainReq.MobileNo == "" {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "mobile_no",
					Message:   "Mobile number is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			if mainReq.FCMToken == "" {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "fcm_token",
					Message:   "FCM token is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			if mainReq.JWTToken == "" {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "jwt_token",
					Message:   "JWT token is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			if mainReq.DeviceID == "" {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "device_id",
					Message:   "Device ID is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Process main screen request with authentication validation
			response, err := h.socketService.HandleMainScreen(mainReq)
			if err != nil {
				errorResp := models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeVerificationError,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "main_screen",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				}
				socket.Emit("connection_error", errorResp)
				return
			}

			// Send only the gamelist data directly
			socket.Emit("main:screen:game:list", response.Data)
		})

		// Game list update trigger handler - fetches from Redis and broadcasts to all connected clients
		socket.On("trigger_game_list_update", func(event *socketio.EventPayload) {
			log.Printf("🎮 Game list update trigger received from %s", socket.Id)

			// Fetch latest game list data from Redis
			gameListData, err := h.socketService.GetGameListFromRedis()
			if err != nil {
				log.Printf("❌ Failed to fetch game list from Redis: %v", err)
				// Fallback to generating fresh data
				gameListData = h.socketService.GetGameListDataPublic()
			} else {
				log.Printf("📖 Successfully fetched game list from Redis")
			}

			// Broadcast the updated game list to all connected clients via main:screen:game:list
			// This follows the same pattern as socket.Emit("main:screen:game:list", response.Data)
			h.io.Emit("main:screen:game:list", gameListData)

			log.Printf("📡 Game list update broadcasted to all connected clients via main:screen:game:list")
		})
		// Heartbeat handler
		socket.On("heartbeat", func(event *socketio.EventPayload) {
			log.Printf("💓 Heartbeat received from %s", socket.Id)
			response := h.socketService.HandleHeartbeat()
			socket.Emit("heartbeat", response)
		})

		// Health check handler
		socket.On("health_check", func(event *socketio.EventPayload) {
			log.Printf("🏥 Health check received from %s", socket.Id)
			response := h.socketService.HandleHealthCheck()
			socket.Emit("health_check:ack", response)
		})

		// Ping handler
		socket.On("ping", func(event *socketio.EventPayload) {
			log.Printf("🏓 Ping received from %s", socket.Id)
			socket.Emit("pong", map[string]interface{}{
				"success":   true,
				"message":   "pong",
				"timestamp": "2024-01-01T00:00:00Z",
			})
		})

		// Disconnect handlers
		socket.On("disconnecting", func(event *socketio.EventPayload) {
			log.Printf("🔌 Socket disconnecting: %s (namespace: %s)", socket.Id, socket.Nps)
		})

		socket.On("disconnect", func(event *socketio.EventPayload) {
			log.Printf("🔌 Socket disconnected: %s (namespace: %s)", socket.Id, socket.Nps)
		})
	})

	// Setup gameplay namespace "/gameplay"
	h.io.Of("/gameplay").OnConnection(func(socket *socketio.Socket) {

		// Send welcome message for gameplay
		socket.Emit("main:screen", map[string]interface{}{
			"success": true,
			"message": "Welcome to gameplay!",
			"server_info": map[string]interface{}{
				"version":  "1.0.0",
				"features": []string{"real-time", "multiplayer", "gameplay"},
			},
		})

		// Player action handler
		socket.On("player_action", func(event *socketio.EventPayload) {

			if len(event.Data) == 0 {
				socket.Emit("player_action:error", map[string]interface{}{
					"success": false,
					"message": "No action data provided",
				})
				return
			}

			// Parse player action request
			actionData, ok := event.Data[0].(map[string]interface{})
			if !ok {
				socket.Emit("player_action:error", map[string]interface{}{
					"success": false,
					"message": "Invalid action data format",
				})
				return
			}

			// Convert to PlayerActionRequest struct
			actionJSON, _ := json.Marshal(actionData)
			var actionReq models.PlayerActionRequest
			if err := json.Unmarshal(actionJSON, &actionReq); err != nil {
				socket.Emit("player_action:error", map[string]interface{}{
					"success": false,
					"message": "Failed to parse action data",
				})
				return
			}

			// Process player action
			response, err := h.socketService.HandlePlayerAction(actionReq)
			if err != nil {
				socket.Emit("player_action:error", map[string]interface{}{
					"success": false,
					"message": err.Error(),
				})
				return
			}

			socket.Emit("player_action:success", response)
		})

		// Game list update handler for gameplay namespace
		socket.On("game_list:updated", func(event *socketio.EventPayload) {
			log.Printf("🎮 Game list update received from gameplay client %s", socket.Id)

			// Fetch latest game list data from Redis
			gameListData, err := h.socketService.GetGameListFromRedis()
			if err != nil {
				log.Printf("❌ Failed to fetch game list from Redis: %v", err)
				// Fallback to generating fresh data
				gameListData = h.socketService.GetGameListDataPublic()
			} else {
				log.Printf("📖 Successfully fetched game list from Redis")
			}

			// Broadcast the updated game list to all connected clients in gameplay namespace via main:screen:game:list
			h.io.Of("/gameplay").Emit("main:screen:game:list", map[string]interface{}{
				"status":    "success",
				"message":   "Game list has been updated from Redis",
				"data":      gameListData,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"event":     "main:screen:game:list",
				"source":    "redis_update",
			})

			log.Printf("📡 Game list update broadcasted to all gameplay clients via main:screen:game:list")
		})

		// Gameplay disconnect handler
		socket.On("disconnect", func(event *socketio.EventPayload) {
			log.Printf("🎮 Gameplay socket disconnected: %s", socket.Id)
		})
	})
}

// GetIo returns the Socket.IO instance
func (h *SocketIoHandler) GetIo() *socketio.Io {
	return h.io
}

// SetupSocketRoutes configures Socket.IO routes for the Fiber app
func (h *SocketIoHandler) SetupSocketRoutes(app *fiber.App) {
	// Serve static files from WEBSITE directory if it exists
	app.Static("/", "./WEBSITE")

	// Setup Socket.IO middleware and routes
	app.Use("/", h.io.Middleware)
	app.Route("/socket.io", h.io.FiberRoute)
}
