package config

import (
	"encoding/json"
	"gofiber/app/models"
	"gofiber/app/services"
	"gofiber/app/utils"
	"strings"
	"time"

	socketio "github.com/doquangtan/socket.io/v4"
	"github.com/gocql/gocql"
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
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "main_screen_data",
				Message:   "No main screen data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		var reqData map[string]interface{}
		var jwtToken string // Declare before the if block
		if raw, ok := event.Data[0].(map[string]interface{}); ok {
			encStr, hasUserData := raw["user_data"].(string)
			if !hasUserData || encStr == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "user_data",
					Message:   "user_data is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			var hasJWT bool
			jwtToken, hasJWT = raw["jwt_token"].(string)
			if !hasJWT || jwtToken == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "jwt_token",
					Message:   "JWT token is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			decrypted, err := utils.DecryptUserData(encStr, jwtToken)
			if err != nil {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "user_data",
					Message:   "Failed to decrypt user_data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			// Ensure jwt_token is set in the decrypted data
			decrypted["jwt_token"] = jwtToken
			reqData = decrypted
		} else {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "main_screen_data",
				Message:   "Invalid main screen data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		mainJSON, _ := json.Marshal(reqData)
		var mainReq models.MainScreenRequest
		if err := json.Unmarshal(mainJSON, &mainReq); err != nil {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "main_screen_data",
				Message:   "Failed to parse main screen data",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}
		// Fallback: if JWTToken is missing in mainReq but present in jwtToken, set it
		if mainReq.JWTToken == "" && jwtToken != "" {
			mainReq.JWTToken = jwtToken
		}
		// Set JWTToken explicitly so HandleMainScreen receives it
		mainReq.JWTToken = jwtToken
		// Validate required fields
		if mainReq.MobileNo == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "mobile_no",
				Message:   "Mobile number is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		if mainReq.FCMToken == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "fcm_token",
				Message:   "FCM token is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		if mainReq.DeviceID == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "device_id",
				Message:   "Device ID is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}
		// Process main screen request with authentication validation
		response, err := h.socketService.HandleMainScreen(mainReq)
		if err != nil {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "main_screen",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}
		// Send only the gamelist data directly
		socket.Emit("main:screen:game:list", response.Data)
	})

	// Game list update trigger handler - fetches from Redis and broadcasts to all connected clients
	socket.On("trigger_game_list_update", func(event *socketio.EventPayload) {
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
			// Fallback to generating fresh data
			gameListData = h.socketService.GetGameListDataPublic()
		}
		// Broadcast the updated game list to all connected clients via main:screen:game:list
		socket.Emit("main:screen:game:list", gameListData)
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
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "contest_data",
				Message:   "No contest data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		var reqData map[string]interface{}
		var jwtToken string // Declare before the if block
		if raw, ok := event.Data[0].(map[string]interface{}); ok {
			encStr, hasUserData := raw["user_data"].(string)
			if !hasUserData || encStr == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "user_data",
					Message:   "user_data is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			var hasJWT bool
			jwtToken, hasJWT = raw["jwt_token"].(string)
			if !hasJWT || jwtToken == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "jwt_token",
					Message:   "JWT token is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			decrypted, err := utils.DecryptUserData(encStr, jwtToken)
			if err != nil {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "user_data",
					Message:   "Failed to decrypt user_data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			// Ensure jwt_token is set in the decrypted data
			decrypted["jwt_token"] = jwtToken
			reqData = decrypted
		} else {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "contest_data",
				Message:   "Invalid contest data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		contestJSON, _ := json.Marshal(reqData)
		var contestReq models.ContestRequest
		if err := json.Unmarshal(contestJSON, &contestReq); err != nil {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "contest_data",
				Message:   "Failed to parse contest data",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}
		// Fallback: if JWTToken is missing in contestReq but present in jwtToken, set it
		if contestReq.JWTToken == "" && jwtToken != "" {
			contestReq.JWTToken = jwtToken
		}
		// Validate required fields
		if contestReq.MobileNo == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "mobile_no",
				Message:   "Mobile number is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		if contestReq.FCMToken == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "fcm_token",
				Message:   "FCM token is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		if contestReq.JWTToken == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "jwt_token",
				Message:   "JWT token is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		if contestReq.DeviceID == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "device_id",
				Message:   "Device ID is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		// Process contest request with authentication validation
		response, err := h.socketService.HandleContestList(contestReq)
		if err != nil {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "contest_list",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}
		// Send contest list data
		socket.Emit("contest:list:response", response)
	})

	// Contest price gap handler
	socket.On("list:contest:gap", func(event *socketio.EventPayload) {
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
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "gap_data",
				Message:   "No gap data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		var reqData map[string]interface{}
		var jwtToken string // Declare before the if block
		if raw, ok := event.Data[0].(map[string]interface{}); ok {
			encStr, hasUserData := raw["user_data"].(string)
			if !hasUserData || encStr == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "user_data",
					Message:   "user_data is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			var hasJWT bool
			jwtToken, hasJWT = raw["jwt_token"].(string)
			if !hasJWT || jwtToken == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "jwt_token",
					Message:   "JWT token is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			decrypted, err := utils.DecryptUserData(encStr, jwtToken)
			if err != nil {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "user_data",
					Message:   "Failed to decrypt user_data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			// Ensure jwt_token is set in the decrypted data
			decrypted["jwt_token"] = jwtToken
			reqData = decrypted
		} else {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "gap_data",
				Message:   "Invalid gap data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		gapJSON, _ := json.Marshal(reqData)
		var gapReq models.ContestGapRequest
		if err := json.Unmarshal(gapJSON, &gapReq); err != nil {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "gap_data",
				Message:   "Failed to parse gap data",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}
		// Fallback: if JWTToken is missing in gapReq but present in jwtToken, set it
		if gapReq.JWTToken == "" && jwtToken != "" {
			gapReq.JWTToken = jwtToken
		}
		// Validate required fields
		if gapReq.MobileNo == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "mobile_no",
				Message:   "Mobile number is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		if gapReq.FCMToken == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "fcm_token",
				Message:   "FCM token is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		if gapReq.JWTToken == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "jwt_token",
				Message:   "JWT token is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		if gapReq.DeviceID == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "device_id",
				Message:   "Device ID is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		// Process contest gap request
		response, err := h.socketService.HandleContestGap(gapReq)
		if err != nil {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "contest_gap",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}
		// Send contest gap response
		socket.Emit("list:contest:gap:response", response)
	})

	// Contest join handler - Simplified version (only join, no matchmaking)
	socket.On("contest:join", func(event *socketio.EventPayload) {
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
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "join_data",
				Message:   "No join data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		var reqData map[string]interface{}
		var jwtToken string // Declare before the if block
		if raw, ok := event.Data[0].(map[string]interface{}); ok {
			encStr, hasUserData := raw["user_data"].(string)
			if !hasUserData || encStr == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "user_data",
					Message:   "user_data is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			var hasJWT bool
			jwtToken, hasJWT = raw["jwt_token"].(string)
			if !hasJWT || jwtToken == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "jwt_token",
					Message:   "JWT token is required",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			decrypted, err := utils.DecryptUserData(encStr, jwtToken)
			if err != nil {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeInvalidFormat,
					ErrorType: models.ErrorTypeFormat,
					Field:     "user_data",
					Message:   "Failed to decrypt user_data",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					SocketID:  socket.Id,
					Event:     "connection_error",
				})
				return
			}
			reqData = decrypted
		} else {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "join_data",
				Message:   "Invalid join data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		joinJSON, _ := json.Marshal(reqData)
		var joinReq models.ContestJoinRequest
		if err := json.Unmarshal(joinJSON, &joinReq); err != nil {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "join_data",
				Message:   "Failed to parse join data",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}
		// Fallback: if JWTToken is missing in joinReq but present in jwtToken, set it
		if joinReq.JWTToken == "" && jwtToken != "" {
			joinReq.JWTToken = jwtToken
		}
		// Validate required fields
		if joinReq.MobileNo == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "mobile_no",
				Message:   "Mobile number is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		if joinReq.FCMToken == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "fcm_token",
				Message:   "FCM token is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		if joinReq.JWTToken == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "jwt_token",
				Message:   "JWT token is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		if joinReq.DeviceID == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "device_id",
				Message:   "Device ID is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		if joinReq.ContestID == "" {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "contest_id",
				Message:   "Contest ID is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}

		// Process contest join request (only join, no matchmaking)
		response, err := h.socketService.HandleContestJoin(joinReq)
		if err != nil {
			socket.Emit("connection_error", models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeAuthentication,
				Field:     "contest_join",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			})
			return
		}
		// Send simple join response without matchmaking
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
		var reqData map[string]interface{}
		var jwtToken string // Declare jwtToken before the if block
		if raw, ok := event.Data[0].(map[string]interface{}); ok {
			encStr, hasUserData := raw["user_data"].(string)
			if !hasUserData || encStr == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "user_data",
					Message:   "user_data is required for authentication",
				})
				return
			}
			var hasJWT bool
			jwtToken, hasJWT = raw["jwt_token"].(string)
			if !hasJWT || jwtToken == "" {
				socket.Emit("opponent:response", map[string]interface{}{
					"status":     "error",
					"error_code": "auth_failed",
					"error_type": "authentication",
					"field":      "jwt_token",
					"message":    "Invalid or expired token",
				})
				return
			}
			decrypted, err := utils.DecryptUserData(encStr, jwtToken)
			if err != nil {
				socket.Emit("opponent:response", map[string]interface{}{
					"status":     "error",
					"error_code": "auth_failed",
					"error_type": "authentication",
					"field":      "user_data",
					"message":    "Failed to decrypt user_data",
				})
				return
			}
			// Fallback: if jwt_token is missing in decrypted but present in jwtToken, set it
			if _, ok := decrypted["jwt_token"]; !ok && jwtToken != "" {
				decrypted["jwt_token"] = jwtToken
			}
			reqData = decrypted
		} else {
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
		// Validate mobile number from token
		if len(simpleJWTData.MobileNo) < 10 {
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
		if len(simpleJWTData.DeviceID) < 1 {
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
		`).Bind(simpleJWTData.MobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)

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
		// Compute join_month for fast lookup
		joinMonth := time.Now().Format("2006-01")
		if joinedAtStr, ok := reqData["joined_at"].(string); ok && joinedAtStr != "" {
			if t, err := time.Parse(time.RFC3339, joinedAtStr); err == nil {
				joinMonth = t.Format("2006-01")
			} else {
				socket.Emit("opponent:response", map[string]interface{}{
					"status":     "error",
					"error_code": "invalid_format",
					"error_type": "format",
					"field":      "joined_at",
					"message":    "Failed to parse joined_at",
				})
				return
			}
		}
		entry, err := h.socketService.GetLeagueJoinEntry(userID, contestID, joinMonth)
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
			// Get user pieces for this game using match_pair_id from league_joins
			var userPieces []map[string]interface{}
			var opponentPieces []map[string]interface{}
			var gameID string
			// Use match_pair_id as game_id from game_pieces table
			gameID = entry.MatchPairID.String()
			gamePiecesService := services.NewGamePiecesService(h.socketService.GetCassandraSession())
			userPieces, err = gamePiecesService.GetUserPiecesCurrentState(gameID, userID)
			if err != nil {
				// Continue without pieces if there's an error
				userPieces = []map[string]interface{}{}
			} else {
				// Enhance pieces with comprehensive data structure
				userPieces = h.enhancePiecesWithComprehensiveData(userPieces, gameID, userID)
			}

			// Get opponent's pieces and enhance them with comprehensive data
			opponentPieces, err = gamePiecesService.GetUserPiecesCurrentState(gameID, entry.OpponentUserID)
			if err != nil {
				// Continue without pieces if there's an error
				opponentPieces = []map[string]interface{}{}
			} else {
				// Enhance pieces with comprehensive data structure
				opponentPieces = h.enhancePiecesWithComprehensiveData(opponentPieces, gameID, entry.OpponentUserID)
			}

			// Get user's dice ID
			var userDiceID gocql.UUID
			err = h.socketService.GetCassandraSession().Query(`SELECT dice_id FROM dice_rolls_lookup WHERE game_id = ? AND user_id = ? LIMIT 1`, gameID, userID).Scan(&userDiceID)
			if err != nil {
				userDiceID = gocql.UUID{}
			}

			// Get opponent's dice ID
			var opponentDiceID gocql.UUID
			err = h.socketService.GetCassandraSession().Query(`SELECT dice_id FROM dice_rolls_lookup WHERE game_id = ? AND user_id = ? LIMIT 1`, gameID, entry.OpponentUserID).Scan(&opponentDiceID)
			if err != nil {
				opponentDiceID = gocql.UUID{}
			}

			response := map[string]interface{}{
				"status":             "success",
				"user_id":            userID,
				"opponent_user_id":   entry.OpponentUserID,
				"opponent_league_id": entry.OpponentLeagueID,
				"joined_at":          entry.JoinedAt.Format(time.RFC3339),
				"game_id":            gameID,
				"user_pieces":        userPieces,
				"opponent_pieces":    opponentPieces,
				"user_dice":          userDiceID.String(),
				"opponent_dice":      opponentDiceID.String(),
				"pieces_status":      "active", // Indicates pieces are available
				"turn_id":            entry.TurnID,
			}
			socket.Emit("opponent:response", response)
		} else {
			// Optionally, emit a minimal/pending response, but do NOT include zero dice IDs
			response := map[string]interface{}{
				"status":             "pending",
				"user_id":            userID,
				"opponent_user_id":   entry.OpponentUserID,
				"opponent_league_id": entry.OpponentLeagueID,
				"joined_at":          entry.JoinedAt.Format(time.RFC3339),
				"game_id":            "",
				"user_pieces":        []map[string]interface{}{},
				"opponent_pieces":    []map[string]interface{}{},
				"pieces_status":      "pending",
				"turn_id":            entry.TurnID,
			}
			socket.Emit("opponent:response", response)
		}
	})

	// Get opponent info during active gameplay
	socket.On("get:opponent:info", func(event *socketio.EventPayload) {
		// Authenticate user
		_, err := authFunc(socket, "get:opponent:info")
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
			socket.Emit("opponent:info:response", map[string]interface{}{
				"status":     "error",
				"error_code": "missing_field",
				"error_type": "field",
				"field":      "request_data",
				"message":    "No data provided",
			})
			return
		}

		var reqData map[string]interface{}
		var jwtToken string // Declare jwtToken before the if block
		if raw, ok := event.Data[0].(map[string]interface{}); ok {
			encStr, hasUserData := raw["user_data"].(string)
			if !hasUserData || encStr == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "user_data",
					Message:   "user_data is required",
				})
				return
			}
			var hasJWT bool
			jwtToken, hasJWT = raw["jwt_token"].(string)
			if !hasJWT || jwtToken == "" {
				socket.Emit("opponent:info:response", map[string]interface{}{
					"status":     "error",
					"error_code": "auth_failed",
					"error_type": "authentication",
					"field":      "jwt_token",
					"message":    "Invalid or expired token",
				})
				return
			}
			decrypted, err := utils.DecryptUserData(encStr, jwtToken)
			if err != nil {
				socket.Emit("opponent:info:response", map[string]interface{}{
					"status":     "error",
					"error_code": "auth_failed",
					"error_type": "authentication",
					"field":      "user_data",
					"message":    "Failed to decrypt user_data",
				})
				return
			}
			// Fallback: if jwt_token is missing in decrypted but present in jwtToken, set it
			if _, ok := decrypted["jwt_token"]; !ok && jwtToken != "" {
				decrypted["jwt_token"] = jwtToken
			}
			reqData = decrypted
		} else {
			socket.Emit("opponent:info:response", map[string]interface{}{
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
		gameID, gameOk := reqData["game_id"].(string)

		if !userOk || userID == "" {
			socket.Emit("opponent:info:response", map[string]interface{}{
				"status":     "error",
				"error_code": "missing_field",
				"error_type": "field",
				"field":      "user_id",
				"message":    "user_id is required",
			})
			return
		}

		if !contestOk || contestID == "" {
			socket.Emit("opponent:info:response", map[string]interface{}{
				"status":     "error",
				"error_code": "missing_field",
				"error_type": "field",
				"field":      "contest_id",
				"message":    "contest_id is required",
			})
			return
		}

		if !gameOk || gameID == "" {
			socket.Emit("opponent:info:response", map[string]interface{}{
				"status":     "error",
				"error_code": "missing_field",
				"error_type": "field",
				"field":      "game_id",
				"message":    "game_id is required",
			})
			return
		}

		// Compute join_month for fast lookup
		joinMonth := time.Now().Format("2006-01")
		if joinedAtStr, ok := reqData["joined_at"].(string); ok && joinedAtStr != "" {
			if t, err := time.Parse(time.RFC3339, joinedAtStr); err == nil {
				joinMonth = t.Format("2006-01")
			} else {
				socket.Emit("opponent:info:response", map[string]interface{}{
					"status":     "error",
					"error_code": "invalid_format",
					"error_type": "format",
					"field":      "joined_at",
					"message":    "Failed to parse joined_at",
				})
				return
			}
		}
		// Get league join entry to find opponent
		entry, err := h.socketService.GetLeagueJoinEntry(userID, contestID, joinMonth)
		if err != nil {
			socket.Emit("opponent:info:response", map[string]interface{}{
				"status":     "error",
				"error_code": "not_found",
				"error_type": "database",
				"field":      "league_joins",
				"message":    "Could not fetch entry",
			})
			return
		}
		if entry.OpponentUserID == "" || entry.OpponentUserID == "NULL" {
			socket.Emit("opponent:info:response", map[string]interface{}{
				"status":  "error",
				"message": "No opponent found for this game",
			})
			return
		}
		// Get opponent user details
		var opponentUser models.User
		err = h.socketService.GetCassandraSession().Query(`
			SELECT id, mobile_no, full_name, status, language_code
			FROM users
			WHERE id = ?
		`, entry.OpponentUserID).Scan(&opponentUser.ID, &opponentUser.MobileNo, &opponentUser.FullName, &opponentUser.Status, &opponentUser.LanguageCode)

		if err != nil {
			socket.Emit("opponent:info:response", map[string]interface{}{
				"status":     "error",
				"error_code": "not_found",
				"error_type": "database",
				"field":      "opponent_user",
				"message":    "Opponent user not found",
			})
			return
		}
		// Get opponent's pieces for this game
		gamePiecesService := services.NewGamePiecesService(h.socketService.GetCassandraSession())
		opponentPieces, err := gamePiecesService.GetUserPiecesCurrentState(gameID, entry.OpponentUserID)
		if err != nil {
			opponentPieces = []map[string]interface{}{}
		}
		response := map[string]interface{}{
			"status":             "success",
			"opponent_user_id":   entry.OpponentUserID,
			"opponent_league_id": entry.OpponentLeagueID,
			"opponent_name":      opponentUser.FullName,
			"opponent_mobile":    opponentUser.MobileNo,
			"game_id":            gameID,
			"contest_id":         contestID,
			"opponent_pieces":    opponentPieces,
			"joined_at":          entry.JoinedAt.Format(time.RFC3339),
		}
		socket.Emit("opponent:info:response", response)
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
		var reqData map[string]interface{}
		var jwtToken string // Declare jwtToken before the if block
		if raw, ok := event.Data[0].(map[string]interface{}); ok {
			encStr, hasUserData := raw["user_data"].(string)
			if !hasUserData || encStr == "" {
				socket.Emit("connection_error", models.ConnectionError{
					Status:    "error",
					ErrorCode: models.ErrorCodeMissingField,
					ErrorType: models.ErrorTypeField,
					Field:     "user_data",
					Message:   "user_data is required",
				})
				return
			}
			var hasJWT bool
			jwtToken, hasJWT = raw["jwt_token"].(string)
			if !hasJWT || jwtToken == "" {
				socket.Emit("cancel:find:response", map[string]interface{}{
					"status":  "error",
					"message": "Invalid or expired token",
				})
				return
			}
			decrypted, err := utils.DecryptUserData(encStr, jwtToken)
			if err != nil {
				socket.Emit("cancel:find:response", map[string]interface{}{
					"status":  "error",
					"message": "Failed to decrypt user_data",
				})
				return
			}
			// Fallback: if jwt_token is missing in decrypted but present in jwtToken, set it
			if _, ok := decrypted["jwt_token"]; !ok && jwtToken != "" {
				decrypted["jwt_token"] = jwtToken
			}
			reqData = decrypted
		} else {
			socket.Emit("cancel:find:response", map[string]interface{}{
				"status":  "error",
				"message": "Invalid data format",
			})
			return
		}

		userID, userOk := reqData["user_id"].(string)
		contestID, contestOk := reqData["contest_id"].(string)
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
		joinMonth := time.Now().Format("2006-01")
		if joinedAtStr, ok := reqData["joined_at"].(string); ok && joinedAtStr != "" {
			if t, err := time.Parse(time.RFC3339, joinedAtStr); err == nil {
				joinMonth = t.Format("2006-01")
			} else {
				socket.Emit("cancel:find:response", map[string]interface{}{
					"status":     "error",
					"error_code": "invalid_format",
					"error_type": "format",
					"field":      "joined_at",
					"message":    "Failed to parse joined_at",
				})
				return
			}
		}
		// Get league join entry to find joined_at
		entry, err := h.socketService.GetLeagueJoinEntry(userID, contestID, joinMonth)
		if err != nil {
			socket.Emit("cancel:find:response", map[string]interface{}{
				"status":  "error",
				"message": "Could not fetch entry",
			})
			return
		}
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

// enhancePiecesWithComprehensiveData enhances basic piece data with comprehensive structure
func (h *GameSocketHandler) enhancePiecesWithComprehensiveData(pieces []map[string]interface{}, gameID, userID string) []map[string]interface{} {
	enhancedPieces := make([]map[string]interface{}, 0, len(pieces))

	for _, piece := range pieces {
		pieceID, _ := piece["piece_id"].(string)
		pieceType, _ := piece["piece_type"].(string)
		toPosLast, _ := piece["to_pos_last"].(string)

		// Create comprehensive current state
		currentState := map[string]interface{}{
			"position":                toPosLast,
			"position_number":         h.parsePositionNumber(toPosLast),
			"status":                  "active",
			"moves":                   0, // Initial state
			"total_positions_visited": []string{h.parsePositionNumber(toPosLast)},
		}

		// Create initial move history
		moveHistory := []map[string]interface{}{
			{
				"move_number":     0,
				"from_pos":        "initial",
				"to_pos":          toPosLast,
				"position_number": h.parsePositionNumber(toPosLast),
				"timestamp":       time.Now().UTC().Format(time.RFC3339),
			},
		}

		// Create piece metadata
		pieceMetadata := map[string]interface{}{
			"created_by":              userID,
			"game_type":               "standard",
			"piece_value":             1,
			"current_position_number": h.parsePositionNumber(toPosLast),
			"total_positions":         57,
		}

		// Create enhanced piece with comprehensive data
		enhancedPiece := map[string]interface{}{
			"game_id":        piece["game_id"],
			"user_id":        piece["user_id"],
			"move_number":    piece["move_number"],
			"piece_id":       pieceID,
			"player_id":      piece["player_id"],
			"from_pos_last":  piece["from_pos_last"],
			"to_pos_last":    toPosLast,
			"piece_type":     pieceType,
			"captured_piece": piece["captured_piece"],
			"created_at":     piece["created_at"],
			"updated_at":     piece["updated_at"],
			"current_state":  currentState,
			"move_history":   moveHistory,
			"piece_metadata": pieceMetadata,
			"total_moves":    0,
			"last_position":  toPosLast,
			"last_move_time": time.Now().UTC().Format(time.RFC3339),
		}

		enhancedPieces = append(enhancedPieces, enhancedPiece)
	}

	return enhancedPieces
}

// parsePositionNumber extracts position number from position string
func (h *GameSocketHandler) parsePositionNumber(position string) string {
	if position == "initial" || position == "" {
		return "0"
	}

	// If position contains "total", extract the number
	if strings.Contains(position, "total") {
		parts := strings.Fields(position)
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	// Otherwise, return the position as is
	return position
}
