package config

import (
	"encoding/json"
	"gofiber/app/models"
	"gofiber/app/services"
	"gofiber/app/utils"
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
func (h *GameSocketHandler) SetupGameHandlers(socket *socketio.Socket, authFunc func(socket *socketio.Socket, eventName string) (*models.User, error)) {
	// Static message handler
	socket.On("main:screen", func(event *socketio.EventPayload) {
		// Authenticate user
		_, err := authFunc(socket, "main:screen")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

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

		// Authenticate user
		_, err := authFunc(socket, "trigger_game_list_update")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

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
		// Authenticate user
		_, err := authFunc(socket, "list:contest")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

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

	// Contest price gap handler
	socket.On("list:contest:gap", func(event *socketio.EventPayload) {
		log.Printf("💰 Contest price gap request received from %s", socket.Id)

		// Authenticate user
		_, err := authFunc(socket, "list:contest:gap")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

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

	// Contest join handler
	socket.On("contest:join", func(event *socketio.EventPayload) {
		log.Printf("🏆 Contest join request received from %s", socket.Id)

		// Authenticate user
		_, err := authFunc(socket, "contest:join")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

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

		// Get user_id from response for matchmaking
		currentUserID := ""
		if response != nil && response.Data != nil {
			if id, ok := response.Data["user_id"].(string); ok {
				currentUserID = id
			}
		}
		leagueID := joinReq.ContestID
		entry, err := h.socketService.GetLeagueJoinEntry(currentUserID, leagueID)
		if err == nil && entry != nil {
			currentJoinedAt := entry.JoinedAt
			opponent, err := h.socketService.MatchAndUpdateOpponent(currentUserID, leagueID, currentJoinedAt)
			if err != nil {
				log.Printf("❌ MatchAndUpdateOpponent error: %v", err)
			} else if opponent != nil {
				if response != nil && response.Data != nil {
					response.Data["opponent"] = map[string]interface{}{
						"opponent_user_id":   opponent.UserID,
						"opponent_league_id": opponent.LeagueID,
					}
				} else {
					log.Printf("⚠️ Response or response.Data is nil, cannot add opponent data")
				}
			} else {
				log.Printf("⏳ No opponent found for matching")
			}
		} else {
			log.Printf("⚠️ Missing required fields for opponent matching - currentUserID: %s, leagueID: %s", currentUserID, leagueID)
		}
		log.Printf("🏆 Contest join response: %+v", response)
		socket.Emit("contest:join:response", response)
	})

	// Add a handler for checking opponent info
	socket.On("check:opponent", func(event *socketio.EventPayload) {
		// Authenticate user
		_, err := authFunc(socket, "check:opponent")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

		if len(event.Data) == 0 {
			socket.Emit("opponent:response", map[string]interface{}{
				"status":     "error",
				"error_code": "missing_field",
				"error_type": "field",
				"field":      "request_data",
				"message":    "No data provided",
			})
			return
		}
		reqData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			socket.Emit("opponent:response", map[string]interface{}{
				"status":     "error",
				"error_code": "invalid_format",
				"error_type": "format",
				"field":      "request_data",
				"message":    "Invalid data format",
			})
			return
		}
		userID, userOk := reqData["user_id"].(string)
		contestID, contestOk := reqData["contest_id"].(string)
		jwtToken, jwtOk := reqData["jwt_token"].(string)

		if !userOk || userID == "" {
			socket.Emit("opponent:response", map[string]interface{}{
				"status":     "error",
				"error_code": "missing_field",
				"error_type": "field",
				"field":      "user_id",
				"message":    "user_id is required and must be a string",
			})
			return
		}
		if !contestOk || contestID == "" {
			socket.Emit("opponent:response", map[string]interface{}{
				"status":     "error",
				"error_code": "missing_field",
				"error_type": "field",
				"field":      "contest_id",
				"message":    "contest_id is required and must be a string",
			})
			return
		}
		if !jwtOk || jwtToken == "" {
			socket.Emit("opponent:response", map[string]interface{}{
				"status":     "error",
				"error_code": "missing_field",
				"error_type": "field",
				"field":      "jwt_token",
				"message":    "jwt_token is required for authentication",
			})
			return
		}
		// Validate the JWT token and extract info (like in service layer)
		simpleJWTData, err := utils.ValidateSimpleJWTToken(jwtToken)
		if err != nil {
			socket.Emit("opponent:response", map[string]interface{}{
				"status":     "error",
				"error_code": "auth_failed",
				"error_type": "authentication",
				"field":      "jwt_token",
				"message":    "Invalid or expired token",
			})
			return
		}
		tokenMobileNo := simpleJWTData.MobileNo
		tokenDeviceID := simpleJWTData.DeviceID
		// Validate mobile number from token
		if len(tokenMobileNo) < 10 {
			socket.Emit("opponent:response", map[string]interface{}{
				"status":     "error",
				"error_code": "auth_failed",
				"error_type": "authentication",
				"field":      "jwt_token",
				"message":    "invalid mobile number in JWT token",
			})
			return
		}
		// Validate device ID from token
		if len(tokenDeviceID) < 1 {
			socket.Emit("opponent:response", map[string]interface{}{
				"status":     "error",
				"error_code": "auth_failed",
				"error_type": "authentication",
				"field":      "jwt_token",
				"message":    "invalid device ID in JWT token",
			})
			return
		}
		var user models.User
		err = h.socketService.GetCassandraSession().Query(`
			SELECT id, mobile_no, full_name, status, language_code
			FROM users
			WHERE mobile_no = ?
			ALLOW FILTERING
		`).Bind(tokenMobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)

		if err != nil {
			socket.Emit("opponent:response", map[string]interface{}{
				"status":     "error",
				"error_code": "auth_failed",
				"error_type": "authentication",
				"field":      "jwt_token",
				"message":    "User not found for token mobile number",
			})
			return
		}

		// Now compare the user ID from database with the request user_id
		if user.ID != userID {
			socket.Emit("opponent:response", map[string]interface{}{
				"status":     "error",
				"error_code": "auth_failed",
				"error_type": "authentication",
				"field":      "user_id",
				"message":    "Token user does not match request user_id",
			})
			return
		}
		log.Printf("🗄️ Fetching league join entry for userID: %s, contestID: %s", userID, contestID)
		log.Printf("🔍 About to call GetLeagueJoinEntry with userID=%s, contestID=%s", userID, contestID)
		entry, err := h.socketService.GetLeagueJoinEntry(userID, contestID)
		if err != nil {
			socket.Emit("opponent:response", map[string]interface{}{
				"status":     "error",
				"error_code": "not_found",
				"error_type": "database",
				"field":      "league_joins",
				"message":    "Could not fetch entry",
			})
			return
		}

		if entry.OpponentUserID != "" && entry.OpponentUserID != "NULL" {
			log.Printf("🔍 OpponentUserID: %s", entry.OpponentUserID)
			log.Printf("✅ Opponent found - sending success response")
			response := map[string]interface{}{
				"status":             "success",
				"opponent_user_id":   entry.OpponentUserID,
				"opponent_league_id": entry.OpponentLeagueID,
				"joined_at":          entry.JoinedAt.Format(time.RFC3339),
			}
			socket.Emit("opponent:response", response)
		} else {
			// Do NOT attempt to match here, just return pending
			response := map[string]interface{}{
				"status":    "pending",
				"message":   "No opponent found yet",
				"joined_at": entry.JoinedAt.Format(time.RFC3339),
			}
			socket.Emit("opponent:response", response)
		}
	})

	socket.On("cancel:find", func(event *socketio.EventPayload) {
		// Authenticate user
		_, err := authFunc(socket, "cancel:find")
		if err != nil {
			if authErr, ok := err.(*AuthenticationError); ok {
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidSession,
					ErrorType: models.ErrorTypeAuthentication,
					Field:     "authentication",
					Message:   err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
			}
			return
		}

		if len(event.Data) == 0 {
			socket.Emit("cancel:find:response", map[string]interface{}{
				"status":  "error",
				"message": "No data provided",
			})
			return
		}
		reqData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			socket.Emit("cancel:find:response", map[string]interface{}{
				"status":  "error",
				"message": "Invalid data format",
			})
			return
		}
		userID, userOk := reqData["user_id"].(string)
		contestID, contestOk := reqData["contest_id"].(string)
		jwtToken, jwtOk := reqData["jwt_token"].(string)
		log.Printf("🔍 userID: %s, contestID: %s, jwtToken: %s", userID, contestID, jwtToken)
		if !userOk || userID == "" {
			socket.Emit("cancel:find:response", map[string]interface{}{
				"status":  "error",
				"message": "user_id is required",
			})
			return
		}
		if !contestOk || contestID == "" {
			socket.Emit("cancel:find:response", map[string]interface{}{
				"status":  "error",
				"message": "contest_id is required",
			})
			return
		}
		if !jwtOk || jwtToken == "" {
			socket.Emit("cancel:find:response", map[string]interface{}{
				"status":  "error",
				"message": "jwt_token is required",
			})
			return
		}
		// Authenticate user
		simpleJWTData, err := utils.ValidateSimpleJWTToken(jwtToken)
		if err != nil {
			socket.Emit("cancel:find:response", map[string]interface{}{
				"status":  "error",
				"message": "Invalid or expired token",
			})
			return
		}
		var user models.User
		err = h.socketService.GetCassandraSession().Query(`
			SELECT id, mobile_no, full_name, status, language_code
			FROM users
			WHERE mobile_no = ?
			ALLOW FILTERING
		`).Bind(simpleJWTData.MobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
		if err != nil || user.ID != userID {
			socket.Emit("cancel:find:response", map[string]interface{}{
				"status":  "error",
				"message": "Authentication failed",
			})
			return
		}
		// Get entry to find joined_at
		entry, err := h.socketService.GetLeagueJoinEntry(userID, contestID)
		if err != nil {
			socket.Emit("cancel:find:response", map[string]interface{}{
				"status":  "error",
				"message": "Could not fetch entry",
			})
			return
		}
		log.Printf("🔍 entry: %+v", entry)
		// Update status_id to '4' in both tables
		err = h.socketService.UpdateLeagueJoinStatusBoth(userID, contestID, "4", entry.JoinedAt.Format(time.RFC3339))
		if err != nil {
			socket.Emit("cancel:find:response", map[string]interface{}{
				"status":  "error",
				"message": "Failed to update status",
			})
			return
		}
		socket.Emit("cancel:find:response", map[string]interface{}{
			"status":  "success",
			"message": "Matchmaking cancelled",
		})
	})
}
