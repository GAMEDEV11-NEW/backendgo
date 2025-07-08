package config

import (
	"gofiber/app/models"
	"gofiber/app/services"
	"log"
	"time"

	socketio "github.com/doquangtan/socket.io/v4"
)

// GameboardSocketHandler handles all gameboard-related Socket.IO events
type GameboardSocketHandler struct {
	socketService *services.SocketService
}

// NewGameboardSocketHandler creates a new gameboard socket handler instance
func NewGameboardSocketHandler(socketService *services.SocketService) *GameboardSocketHandler {
	return &GameboardSocketHandler{
		socketService: socketService,
	}
}

// SetupGameboardHandlers configures all gameboard-related Socket.IO event handlers
func (h *GameboardSocketHandler) SetupGameboardHandlers(socket *socketio.Socket, authFunc func(socket *socketio.Socket, eventName string) (*models.User, error)) {
	// Gameboard initialization handler
	socket.On("gameboard:init", func(event *socketio.EventPayload) {
		log.Printf("🎯 Gameboard initialization request from %s", socket.Id)

		// Authenticate user
		_, err := authFunc(socket, "gameboard:init")
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
				Field:     "gameboard_data",
				Message:   "No gameboard data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Parse gameboard request
		gameboardData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "gameboard_data",
				Message:   "Invalid gameboard data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Validate required fields
		mobileNo, mobileOk := gameboardData["mobile_no"].(string)
		jwtToken, jwtOk := gameboardData["jwt_token"].(string)
		deviceID, deviceOk := gameboardData["device_id"].(string)
		contestID, contestOk := gameboardData["contest_id"].(string)

		if !mobileOk || mobileNo == "" {
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

		if !jwtOk || jwtToken == "" {
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

		if !deviceOk || deviceID == "" {
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

		if !contestOk || contestID == "" {
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

	})

	// Dice roll handler
	socket.On("dice:roll", func(event *socketio.EventPayload) {
		log.Printf("🎲 Dice roll request from %s", socket.Id)

		// Debug: Print event data
		if len(event.Data) > 0 {
			log.Printf("🔍 DEBUG: Event data received: %+v", event.Data[0])
		} else {
			log.Printf("⚠️ DEBUG: No event data received")
		}

		// Authenticate user
		log.Printf("🔐 DEBUG: Calling authFunc for socket %s", socket.Id)
		user, err := authFunc(socket, "dice:roll")
		if err != nil {
			log.Printf("❌ DEBUG: Authentication failed for socket %s: %v", socket.Id, err)
			if authErr, ok := err.(*AuthenticationError); ok {
				log.Printf("🔍 DEBUG: AuthenticationError type: %+v", authErr)
				socket.Emit("authentication_error", authErr.ConnectionError)
			} else {
				log.Printf("🔍 DEBUG: Generic error type: %T, error: %v", err, err)
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

		log.Printf("✅ DEBUG: Authentication successful for user: %+v", user)

		if len(event.Data) == 0 {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "dice_data",
				Message:   "No dice data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Parse dice request
		log.Printf("🔍 DEBUG: Parsing dice data from event")
		diceData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			log.Printf("❌ DEBUG: Failed to parse dice data as map[string]interface{}")
			log.Printf("🔍 DEBUG: Event data type: %T", event.Data[0])
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "dice_data",
				Message:   "Invalid dice data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		log.Printf("✅ DEBUG: Dice data parsed successfully: %+v", diceData)

		// Validate required fields
		log.Printf("🔍 DEBUG: Extracting fields from dice data")
		gameID, gameOk := diceData["game_id"].(string)
		contestID, contestOk := diceData["contest_id"].(string)
		sessionToken, sessionOk := diceData["session_token"].(string)
		deviceID, deviceOk := diceData["device_id"].(string)
		jwtToken, jwtOk := diceData["jwt_token"].(string)

		log.Printf("🔍 DEBUG: Field extraction results:")
		log.Printf("  - game_id: %s (ok: %t)", gameID, gameOk)
		log.Printf("  - contest_id: %s (ok: %t)", contestID, contestOk)
		log.Printf("  - session_token: %s (ok: %t)", sessionToken, sessionOk)
		log.Printf("  - device_id: %s (ok: %t)", deviceID, deviceOk)
		log.Printf("  - jwt_token: %s (ok: %t)", jwtToken, jwtOk)

		if !gameOk || gameID == "" {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "game_id",
				Message:   "Game ID is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		if !contestOk || contestID == "" {
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

		if !sessionOk || sessionToken == "" {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "session_token",
				Message:   "Session token is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		if !deviceOk || deviceID == "" {
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

		if !jwtOk || jwtToken == "" {
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

		// Create dice service and roll dice
		diceService := services.NewDiceService(h.socketService.GetCassandraSession())
		rollReq := models.DiceRollRequest{
			GameID:       gameID,
			ContestID:    contestID,
			SessionToken: sessionToken,
			DeviceID:     deviceID,
			JWTToken:     jwtToken,
		}

		response, err := diceService.RollDice(rollReq, user.ID)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeSystem,
				Field:     "dice_roll",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Send successful response
		response.SocketID = socket.Id
		socket.Emit("dice:roll:response", response)
		log.Printf("🎲 Dice roll completed - User: %s, Game: %s, Number: %d", user.ID, gameID, response.DiceNumber)
	})

	// Dice history handler
	socket.On("dice:history", func(event *socketio.EventPayload) {
		log.Printf("📊 Dice history request from %s", socket.Id)

		// Authenticate user
		user, err := authFunc(socket, "dice:history")
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
				Field:     "history_data",
				Message:   "No history data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Parse history request
		historyData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "history_data",
				Message:   "Invalid history data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Validate required fields
		gameID, gameOk := historyData["game_id"].(string)
		sessionToken, sessionOk := historyData["session_token"].(string)
		limit, _ := historyData["limit"].(float64)

		if !gameOk || gameID == "" {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "game_id",
				Message:   "Game ID is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		if !sessionOk || sessionToken == "" {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "session_token",
				Message:   "Session token is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Create dice service and get history
		diceService := services.NewDiceService(h.socketService.GetCassandraSession())
		historyReq := models.DiceHistoryRequest{
			GameID:       gameID,
			UserID:       user.ID,
			SessionToken: sessionToken,
			Limit:        int(limit),
		}

		response, err := diceService.GetDiceHistory(historyReq)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeSystem,
				Field:     "dice_history",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Send successful response
		response.SocketID = socket.Id
		socket.Emit("dice:history:response", response)
		log.Printf("📊 Dice history retrieved - User: %s, Game: %s, Total Rolls: %d", user.ID, gameID, response.TotalRolls)
	})

}
