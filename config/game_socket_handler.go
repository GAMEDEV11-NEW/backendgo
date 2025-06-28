package config

import (
	"encoding/json"
	"gofiber/app/models"
	"gofiber/app/services"
	"log"
	"time"

	socketio "github.com/doquangtan/socket.io/v4"
)

// GameSocketHandler handles all game and contest-related Socket.IO events
type GameSocketHandler struct {
	socketService *services.SocketService
}

// NewGameSocketHandler creates a new game socket handler instance
func NewGameSocketHandler(socketService *services.SocketService) *GameSocketHandler {
	return &GameSocketHandler{
		socketService: socketService,
	}
}

// SetupGameHandlers configures all game and contest-related Socket.IO event handlers
func (h *GameSocketHandler) SetupGameHandlers(socket *socketio.Socket) {
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
		socket.Emit("main:screen:game:list", gameListData)

		log.Printf("📡 Game list update broadcasted to all connected clients via main:screen:game:list")
	})

	// Contest list handler
	socket.On("list:contest", func(event *socketio.EventPayload) {
		if len(event.Data) == 0 {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "contest_data",
				Message:   "No contest data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Parse contest request
		contestData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "contest_data",
				Message:   "Invalid contest data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Convert to ContestRequest struct
		contestJSON, _ := json.Marshal(contestData)
		var contestReq models.ContestRequest
		if err := json.Unmarshal(contestJSON, &contestReq); err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "contest_data",
				Message:   "Failed to parse contest data",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Validate required fields
		if contestReq.MobileNo == "" {
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

		if contestReq.FCMToken == "" {
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

		if contestReq.JWTToken == "" {
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

		if contestReq.DeviceID == "" {
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

		// Process contest request with authentication validation
		response, err := h.socketService.HandleContestList(contestReq)
		log.Printf("🏆 Contest list request received - Data: %+v", response)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "contest_list",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Send contest list data
		socket.Emit("contest:list:response", response)
	})

	// Contest join handler
	socket.On("contest:join", func(event *socketio.EventPayload) {
		log.Printf("🏆 Contest join request received from %s", socket.Id)

		if len(event.Data) == 0 {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "join_data",
				Message:   "No join data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Parse contest join request
		joinData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "join_data",
				Message:   "Invalid join data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Convert to ContestJoinRequest struct
		joinJSON, _ := json.Marshal(joinData)
		var joinReq models.ContestJoinRequest
		if err := json.Unmarshal(joinJSON, &joinReq); err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "join_data",
				Message:   "Failed to parse join data",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Validate required fields
		if joinReq.MobileNo == "" {
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

		if joinReq.FCMToken == "" {
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

		if joinReq.JWTToken == "" {
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

		if joinReq.DeviceID == "" {
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

		if joinReq.ContestID == "" {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "contest_id",
				Message:   "Contest ID is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Process contest join request
		response, err := h.socketService.HandleContestJoin(joinReq)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "contest_join",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Send contest join response
		socket.Emit("contest:join:response", response)
	})

	// Contest price gap handler
	socket.On("list:contest:gap", func(event *socketio.EventPayload) {
		log.Printf("💰 Contest price gap request received from %s", socket.Id)

		if len(event.Data) == 0 {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "gap_data",
				Message:   "No gap data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Parse contest gap request
		gapData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "gap_data",
				Message:   "Invalid gap data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Convert to ContestGapRequest struct
		gapJSON, _ := json.Marshal(gapData)
		var gapReq models.ContestGapRequest
		if err := json.Unmarshal(gapJSON, &gapReq); err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "gap_data",
				Message:   "Failed to parse gap data",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Validate required fields
		if gapReq.MobileNo == "" {
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

		if gapReq.FCMToken == "" {
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

		if gapReq.JWTToken == "" {
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

		if gapReq.DeviceID == "" {
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

		// Process contest gap request
		response, err := h.socketService.HandleContestGap(gapReq)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "contest_gap",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Send contest gap response
		socket.Emit("list:contest:gap:response", response)
	})
}

// SetupGameplayHandlers configures all gameplay-related Socket.IO event handlers
func (h *GameSocketHandler) SetupGameplayHandlers(socket *socketio.Socket) {
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
		socket.Emit("main:screen:game:list", map[string]interface{}{
			"status":    "success",
			"message":   "Game list has been updated from Redis",
			"data":      gameListData,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"event":     "main:screen:game:list",
			"source":    "redis_update",
		})

		log.Printf("📡 Game list update broadcasted to all gameplay clients via main:screen:game:list")
	})
} 