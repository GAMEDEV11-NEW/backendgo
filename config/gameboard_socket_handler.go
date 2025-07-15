package config

import (
	"fmt"
	"gofiber/app/models"
	"gofiber/app/services"
	"time"

	"gofiber/app/utils"

	socketio "github.com/doquangtan/socket.io/v4"
	"github.com/gocql/gocql"
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

		// Authenticate user
		user, err := authFunc(socket, "gameboard:init")
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

		// Store session data for this socket connection
		h.storeSessionData(socket.Id, user.ID, mobileNo, deviceID)

		// Send successful gameboard initialization response
		response := map[string]interface{}{
			"status":     "success",
			"message":    "Gameboard initialized successfully",
			"game_id":    contestID, // Using contest_id as game_id for now
			"user_id":    user.ID,
			"mobile_no":  mobileNo,
			"device_id":  deviceID,
			"contest_id": contestID,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"socket_id":  socket.Id,
			"event":      "gameboard:init:response",
		}
		socket.Emit("gameboard:init:response", response)
	})

	// Dice roll handler
	socket.On("dice:roll", func(event *socketio.EventPayload) {
		fmt.Printf("DEBUG: dice:roll event received for socket %s\n", socket.Id)

		// Authenticate user
		user, err := authFunc(socket, "dice:roll")
		if err != nil {
			fmt.Printf("DEBUG: Authentication failed for socket %s: %v\n", socket.Id, err)
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
		fmt.Printf("DEBUG: Authentication successful for user %s\n", user.ID)

		if len(event.Data) == 0 {
			fmt.Printf("DEBUG: No dice data provided for socket %s\n", socket.Id)
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

		// Support both legacy and new encrypted payloads
		var diceData map[string]interface{}
		var jwtToken string
		if raw, ok := event.Data[0].(map[string]interface{}); ok {
			// Check for new encrypted format
			if encrypted, hasEncrypted := raw["user_data"]; hasEncrypted {
				jwtToken, _ = raw["jwt_token"].(string)
				encStr, _ := encrypted.(string)
				if jwtToken != "" && encStr != "" {
					decrypted, err := utils.DecryptUserData(encStr, jwtToken)
					if err != nil {
						fmt.Printf("DEBUG: Failed to decrypt user_data: %v\n", err)
						errorResp := models.ConnectionError{
							Status:    "error",
							ErrorCode: models.ErrorCodeInvalidFormat,
							ErrorType: models.ErrorTypeFormat,
							Field:     "user_data",
							Message:   "Failed to decrypt user_data",
							Timestamp: time.Now().UTC().Format(time.RFC3339),
							SocketID:  socket.Id,
							Event:     "connection_error",
						}
						socket.Emit("connection_error", errorResp)
						return
					}
					diceData = decrypted
				} else {
					errorResp := models.ConnectionError{
						Status:    "error",
						ErrorCode: models.ErrorCodeMissingField,
						ErrorType: models.ErrorTypeField,
						Field:     "jwt_token/user_data",
						Message:   "jwt_token and user_data are required",
						Timestamp: time.Now().UTC().Format(time.RFC3339),
						SocketID:  socket.Id,
						Event:     "connection_error",
					}
					socket.Emit("connection_error", errorResp)
					return
				}
			} else {
				// Legacy format
				diceData = raw
			}
		} else {
			fmt.Printf("DEBUG: Invalid dice data format for socket %s\n", socket.Id)
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

		// Validate required fields
		gameID, gameOk := diceData["game_id"].(string)
		contestID, contestOk := diceData["contest_id"].(string)
		sessionToken, sessionOk := diceData["session_token"].(string)
		deviceID, deviceOk := diceData["device_id"].(string)
		jwtToken, jwtOk := diceData["jwt_token"].(string)

		fmt.Printf("DEBUG: Field validation - gameID: %s, contestID: %s, sessionToken: %s, deviceID: %s, jwtToken: %s\n",
			gameID, contestID, sessionToken, deviceID, jwtToken)

		if !gameOk || gameID == "" {
			fmt.Printf("DEBUG: Missing game_id for socket %s\n", socket.Id)
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
			fmt.Printf("DEBUG: Missing contest_id for socket %s\n", socket.Id)
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
			fmt.Printf("DEBUG: Missing session_token for socket %s\n", socket.Id)
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
			fmt.Printf("DEBUG: Missing device_id for socket %s\n", socket.Id)
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
			fmt.Printf("DEBUG: Missing jwt_token for socket %s\n", socket.Id)
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

		fmt.Printf("DEBUG: All fields validated successfully for user %s, game %s\n", user.ID, gameID)

		// Create dice service and roll dice
		fmt.Printf("DEBUG: Creating dice service for user %s\n", user.ID)
		diceService := services.NewDiceService(h.socketService.GetCassandraSession())
		rollReq := models.DiceRollRequest{
			GameID:       gameID,
			ContestID:    contestID,
			SessionToken: sessionToken,
			DeviceID:     deviceID,
			JWTToken:     jwtToken,
		}

		fmt.Printf("DEBUG: Calling RollDice for user %s, game %s\n", user.ID, gameID)
		response, err := diceService.RollDice(rollReq, user.ID)
		if err != nil {
			fmt.Printf("DEBUG: RollDice failed for user %s: %v\n", user.ID, err)
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

		fmt.Printf("DEBUG: Dice roll successful for user %s, dice number: %d\n", user.ID, response.DiceNumber)

		// Send successful response to the player who rolled
		response.SocketID = socket.Id
		fmt.Printf("DEBUG: Sending dice:roll:response to user %s\n", user.ID)
		socket.Emit("dice:roll:response", response)

		// Broadcast dice roll to opponent
		fmt.Printf("DEBUG: Broadcasting dice roll to opponent for game %s, user %s\n", gameID, user.ID)
		h.broadcastDiceRollToOpponent(gameID, user.ID, response)
		fmt.Printf("DEBUG: dice:roll event completed for user %s\n", user.ID)
	})

	// Dice history handler
	socket.On("dice:history", func(event *socketio.EventPayload) {

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
	})

	// Piece move handler
	socket.On("piece:move", func(event *socketio.EventPayload) {

		// Authenticate user
		user, err := authFunc(socket, "piece:move")
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
				Field:     "piece_move_data",
				Message:   "No piece move data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Parse piece move request
		moveData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "piece_move_data",
				Message:   "Invalid piece move data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Validate required fields
		gameID, gameOk := moveData["game_id"].(string)
		pieceID, pieceOk := moveData["piece_id"].(string)
		fromPos, fromOk := moveData["from_pos"].(string)
		toPos, toOk := moveData["to_pos"].(string)
		pieceType, typeOk := moveData["piece_type"].(string)
		capturedPiece, _ := moveData["captured_piece"].(string)
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

		// Validate game_id format (should be a UUID, not just "1")
		if gameID == "1" || len(gameID) < 10 {
			// Continue processing but log the warning
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "game_id",
				Message:   "Invalid game_id format",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		if !pieceOk || pieceID == "" {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "piece_id",
				Message:   "Piece ID is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		if !fromOk || fromPos == "" {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "from_pos",
				Message:   "From position is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		if !toOk || toPos == "" {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "to_pos",
				Message:   "To position is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		if !typeOk || pieceType == "" {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeMissingField,
				ErrorType: models.ErrorTypeField,
				Field:     "piece_type",
				Message:   "Piece type is required",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Create game pieces service and record move
		gamePiecesService := services.NewGamePiecesService(h.socketService.GetCassandraSession())
		err = gamePiecesService.RecordPieceMove(gameID, user.ID, pieceID, fromPos, toPos, pieceType, capturedPiece)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeSystem,
				Field:     "piece_move",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Send confirmation to the player who made the move
		moveConfirmation := map[string]interface{}{
			"status":     "success",
			"message":    "Piece move recorded successfully",
			"game_id":    gameID,
			"user_id":    user.ID,
			"piece_id":   pieceID,
			"from_pos":   fromPos,
			"to_pos":     toPos,
			"piece_type": pieceType,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"socket_id":  socket.Id,
			"event":      "piece:move:response",
		}
		if capturedPiece != "" {
			moveConfirmation["captured_piece"] = capturedPiece
		}
		socket.Emit("piece:move:response", moveConfirmation)

		// Broadcast move to opponent
		h.broadcastMoveToOpponent(gameID, user.ID, pieceID, fromPos, toPos, pieceType, capturedPiece)
	})

	// Add game end handler to save all piece moves to database
	socket.On("game:end", func(event *socketio.EventPayload) {
		// Authenticate user
		user, err := authFunc(socket, "game:end")
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
				Field:     "game_end_data",
				Message:   "No game end data provided",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Parse game end request
		endData, ok := event.Data[0].(map[string]interface{})
		if !ok {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeInvalidFormat,
				ErrorType: models.ErrorTypeFormat,
				Field:     "game_end_data",
				Message:   "Invalid game end data format",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Validate required fields
		gameID, gameOk := endData["game_id"].(string)
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

		// Create game pieces service and save all piece moves to database
		gamePiecesService := services.NewGamePiecesService(h.socketService.GetCassandraSession())
		err = gamePiecesService.SaveGamePiecesToDatabase(gameID)
		if err != nil {
			errorResp := models.ConnectionError{
				Status:    "error",
				ErrorCode: models.ErrorCodeVerificationError,
				ErrorType: models.ErrorTypeSystem,
				Field:     "game_end",
				Message:   err.Error(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				SocketID:  socket.Id,
				Event:     "connection_error",
			}
			socket.Emit("connection_error", errorResp)
			return
		}

		// Send confirmation to the user
		endConfirmation := map[string]interface{}{
			"status":    "success",
			"message":   "Game ended successfully and all piece moves saved to database",
			"game_id":   gameID,
			"user_id":   user.ID,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"socket_id": socket.Id,
			"event":     "game:end:response",
		}
		socket.Emit("game:end:response", endConfirmation)

		// Broadcast comprehensive game end to opponent with original data
		h.broadcastGameEndToOpponent(gameID, user.ID)
	})
}

// storeSessionData stores session data for a socket connection
func (h *GameboardSocketHandler) storeSessionData(socketID, userID, mobileNo, deviceID string) {
	// Update session data with socket information
	sessions, err := h.socketService.GetSessionService().GetSessionsByUserID(userID)
	if err != nil {
		return
	}

	// Update the most recent active session with socket ID
	for _, session := range sessions {
		if session.IsActive && session.MobileNo == mobileNo && session.DeviceID == deviceID {
			updates := map[string]interface{}{
				"socket_id": socketID,
				"last_seen": time.Now(),
			}
			_ = h.socketService.GetSessionService().UpdateSession(session.SessionToken, updates)
			break
		}
	}
}

// broadcastDiceRollToOpponent sends the dice roll update to the opponent
func (h *GameboardSocketHandler) broadcastDiceRollToOpponent(gameID, userID string, diceResponse *models.DiceRollResponse) {
	fmt.Printf("DEBUG: broadcastDiceRollToOpponent called for game %s, user %s\n", gameID, userID)

	// Get opponent user ID from the game
	opponentUserID := h.getOpponentUserID(gameID, userID)
	if opponentUserID == "" {
		fmt.Printf("DEBUG: No opponent found for game %s, user %s\n", gameID, userID)
		return
	}
	fmt.Printf("DEBUG: Found opponent %s for game %s, user %s\n", opponentUserID, gameID, userID)

	// Create dice roll update message
	diceUpdate := map[string]interface{}{
		"status":      "opponent_dice_roll",
		"game_id":     gameID,
		"user_id":     userID,
		"dice_id":     diceResponse.DiceID,
		"dice_number": diceResponse.DiceNumber,
		"roll_time":   diceResponse.RollTime,
		"contest_id":  diceResponse.ContestID,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"event":       "opponent:dice:roll:update",
		"data":        diceResponse.Data,
	}

	fmt.Printf("DEBUG: Created dice update message for opponent %s: %+v\n", opponentUserID, diceUpdate)

	// Send dice roll update to opponent
	fmt.Printf("DEBUG: Sending dice roll update to opponent %s\n", opponentUserID)
	h.sendMessageToUser(opponentUserID, diceUpdate, "opponent:dice:roll:update")
	fmt.Printf("DEBUG: broadcastDiceRollToOpponent completed for opponent %s\n", opponentUserID)
}

// broadcastMoveToOpponent sends the move update to the opponent
func (h *GameboardSocketHandler) broadcastMoveToOpponent(gameID, userID, pieceID, fromPos, toPos, pieceType, capturedPiece string) {
	// Get opponent user ID from the game
	opponentUserID := h.getOpponentUserID(gameID, userID)
	if opponentUserID == "" {
		return
	}

	// Create simple move update message (basic format)
	moveUpdate := map[string]interface{}{
		"event":      "opponent:move:update",
		"from_pos":   fromPos,
		"game_id":    gameID,
		"piece_id":   pieceID,
		"piece_type": pieceType,
		"status":     "opponent_move",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"to_pos":     toPos,
		"user_id":    userID,
	}

	// Add captured piece if exists
	if capturedPiece != "" {
		moveUpdate["captured_piece"] = capturedPiece
	}

	// Send simple move update to opponent
	h.sendMessageToUser(opponentUserID, moveUpdate, "opponent:move:update")

}

// broadcastGameEndToOpponent sends the game end update to the opponent
func (h *GameboardSocketHandler) broadcastGameEndToOpponent(gameID, userID string) {
	// Get opponent user ID from the game
	opponentUserID := h.getOpponentUserID(gameID, userID)
	if opponentUserID == "" {
		return
	}

	// Get final game state and statistics
	gamePiecesService := services.NewGamePiecesService(h.socketService.GetCassandraSession())

	// Get final piece states for both users
	userPieces, err := gamePiecesService.GetUserPiecesCurrentState(gameID, userID)
	if err != nil {
		userPieces = []map[string]interface{}{}
	}

	opponentPieces, err := gamePiecesService.GetUserPiecesCurrentState(gameID, opponentUserID)
	if err != nil {
		opponentPieces = []map[string]interface{}{}
	}

	// Calculate game statistics
	totalUserMoves := 0
	totalOpponentMoves := 0
	userCompletedPieces := 0
	opponentCompletedPieces := 0

	// Count user moves and completed pieces
	for _, piece := range userPieces {
		if pieceData, ok := piece["current_state"].(map[string]interface{}); ok {
			if moves, ok := pieceData["moves"].(int); ok {
				totalUserMoves += moves
			}
			if position, ok := pieceData["position"].(string); ok {
				if position == "57" { // Final position
					userCompletedPieces++
				}
			}
		}
	}

	// Count opponent moves and completed pieces
	for _, piece := range opponentPieces {
		if pieceData, ok := piece["current_state"].(map[string]interface{}); ok {
			if moves, ok := pieceData["moves"].(int); ok {
				totalOpponentMoves += moves
			}
			if position, ok := pieceData["position"].(string); ok {
				if position == "57" { // Final position
					opponentCompletedPieces++
				}
			}
		}
	}

	// Determine winner (more completed pieces, or more moves if tied)
	winner := "tie"
	if userCompletedPieces > opponentCompletedPieces {
		winner = userID
	} else if opponentCompletedPieces > userCompletedPieces {
		winner = opponentUserID
	} else if totalUserMoves > totalOpponentMoves {
		winner = userID
	} else if totalOpponentMoves > totalUserMoves {
		winner = opponentUserID
	}

	// Create comprehensive game end update message
	gameEndUpdate := map[string]interface{}{
		"status":           "opponent_game_end",
		"game_id":          gameID,
		"user_id":          userID,
		"opponent_user_id": opponentUserID,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"event":            "opponent:game:end:update",
		"game_end_data": map[string]interface{}{
			"winner":                    winner,
			"winner_user_id":            winner,
			"game_status":               "completed",
			"total_user_moves":          totalUserMoves,
			"total_opponent_moves":      totalOpponentMoves,
			"user_completed_pieces":     userCompletedPieces,
			"opponent_completed_pieces": opponentCompletedPieces,
			"user_pieces":               userPieces,
			"opponent_pieces":           opponentPieces,
			"database_saved":            true,
			"redis_cleared":             true,
			"game_duration":             "completed", // Could calculate actual duration if needed
		},
		"message":           "Game ended successfully - all piece moves saved to database",
		"notification_type": "game_completion",
		"priority":          "high",
	}

	// Send comprehensive game end update to opponent
	h.sendMessageToUser(opponentUserID, gameEndUpdate, "opponent:game:end:update")

}

// getOpponentUserID finds the opponent user ID for a given game and user
func (h *GameboardSocketHandler) getOpponentUserID(gameID, userID string) string {
	fmt.Printf("DEBUG: getOpponentUserID called for game %s, user %s\n", gameID, userID)

	// Query match_pairs table to find the opponent using user1_data and user2_data columns
	var user1ID, user2ID, user1Data, user2Data string

	err := h.socketService.GetCassandraSession().Query(`
		SELECT user1_id, user2_id, user1_data, user2_data FROM match_pairs 
		WHERE id = ?
	`, gameID).Scan(&user1ID, &user2ID, &user1Data, &user2Data)

	if err != nil {
		fmt.Printf("DEBUG: Failed to find match pair by game ID %s: %v\n", gameID, err)
		// Try to find match pair by user ID instead
		iter := h.socketService.GetCassandraSession().Query(`
			SELECT id, user1_id, user2_id, user1_data, user2_data, status FROM match_pairs 
			WHERE user1_id = ? OR user2_id = ?
			ALLOW FILTERING
		`, userID, userID).Iter()

		var matchID gocql.UUID
		var u1ID, u2ID, u1Data, u2Data, status string
		for iter.Scan(&matchID, &u1ID, &u2ID, &u1Data, &u2Data, &status) {
			if status == "active" {
				fmt.Printf("DEBUG: Found match pair by user ID: user1=%s, user2=%s, user1_data=%s, user2_data=%s\n", u1ID, u2ID, u1Data, u2Data)

				// Check if current user matches user1_data or user2_data
				if u1Data == userID {
					fmt.Printf("DEBUG: User %s found in user1_data, opponent is %s\n", userID, u2Data)
					iter.Close()
					return u2Data
				} else if u2Data == userID {
					fmt.Printf("DEBUG: User %s found in user2_data, opponent is %s\n", userID, u1Data)
					iter.Close()
					return u1Data
				}
			}
		}
		iter.Close()
		fmt.Printf("DEBUG: User %s not found in any match pair data\n", userID)
		return ""
	}

	fmt.Printf("DEBUG: Found match pair by game ID: user1=%s, user2=%s, user1_data=%s, user2_data=%s\n", user1ID, user2ID, user1Data, user2Data)

	// Check if current user matches user1_data or user2_data
	if user1Data == userID {
		fmt.Printf("DEBUG: User %s found in user1_data, opponent is %s\n", userID, user2Data)
		return user2Data
	} else if user2Data == userID {
		fmt.Printf("DEBUG: User %s found in user2_data, opponent is %s\n", userID, user1Data)
		return user1Data
	}

	fmt.Printf("DEBUG: User %s not found in match pair data (user1_data=%s, user2_data=%s)\n", userID, user1Data, user2Data)
	return ""
}

// sendMessageToUser sends a message to a specific user using available services
func (h *GameboardSocketHandler) sendMessageToUser(userID string, message map[string]interface{}, eventName string) {
	fmt.Printf("DEBUG: sendMessageToUser called for user %s, event %s\n", userID, eventName)

	// Try messaging service first
	if h.socketService.GetMessagingService() != nil {
		fmt.Printf("DEBUG: Using messaging service for user %s\n", userID)
		messageData := services.MessageData{
			Event:     eventName,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Priority:  "high",
			Data:      message,
		}

		err := h.socketService.GetMessagingService().SendMessageToUser(userID, messageData)
		if err != nil {
			fmt.Printf("DEBUG: Messaging service failed for user %s: %v, trying direct socket emission\n", userID, err)
			// Try direct socket emission as fallback
			h.sendDirectToSockets(userID, eventName, message)
		} else {
			fmt.Printf("DEBUG: Message sent successfully via messaging service to user %s\n", userID)
		}
	} else {
		fmt.Printf("DEBUG: No messaging service available, using direct socket emission for user %s\n", userID)
		// Try direct socket emission as fallback
		h.sendDirectToSockets(userID, eventName, message)
	}
}

// sendDirectToSockets sends a message directly to all sockets of a user
func (h *GameboardSocketHandler) sendDirectToSockets(userID string, eventName string, message map[string]interface{}) {
	fmt.Printf("DEBUG: sendDirectToSockets called for user %s, event %s\n", userID, eventName)

	// Get all sessions for this user
	sessions, err := h.socketService.GetSessionService().GetSessionsByUserID(userID)
	if err != nil {
		fmt.Printf("DEBUG: Failed to get sessions for user %s: %v\n", userID, err)
		return
	}

	fmt.Printf("DEBUG: Found %d sessions for user %s\n", len(sessions), userID)
	if len(sessions) == 0 {
		fmt.Printf("DEBUG: No sessions found for user %s\n", userID)
		return
	}

	var sentCount int
	for _, session := range sessions {
		fmt.Printf("DEBUG: Processing session for user %s: socketID=%s, isActive=%t\n", userID, session.SocketID, session.IsActive)
		if session.SocketID != "" {
			// Try to emit directly to the socket using Socket.IO instance
			if h.socketService.GetIo() != nil {
				fmt.Printf("DEBUG: Using Socket.IO instance for user %s, socket %s\n", userID, session.SocketID)
				// Get all connected sockets
				sockets := h.socketService.GetIo().Sockets()
				fmt.Printf("DEBUG: Total connected sockets: %d\n", len(sockets))

				// Find the specific socket and emit the message
				for _, socket := range sockets {
					if socket.Id == session.SocketID {
						fmt.Printf("DEBUG: Found matching socket %s for user %s, emitting message\n", session.SocketID, userID)
						socket.Emit(eventName, message)
						sentCount++
						break
					}
				}
			} else if h.socketService.GetMessagingService() != nil {
				fmt.Printf("DEBUG: Using messaging service fallback for user %s, socket %s\n", userID, session.SocketID)
				// Fallback to messaging service
				err := h.socketService.GetMessagingService().SendMessageToSocket(session.SocketID, services.MessageData{
					Event:     eventName,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Priority:  "high",
					Data:      message,
				})
				if err == nil {
					sentCount++
					fmt.Printf("DEBUG: Message sent via messaging service to socket %s\n", session.SocketID)
				} else {
					fmt.Printf("DEBUG: Failed to send message via messaging service to socket %s: %v\n", session.SocketID, err)
				}
			} else {
				fmt.Printf("DEBUG: No Socket.IO or messaging service available for user %s\n", userID)
				sentCount++
			}
		} else {
			fmt.Printf("DEBUG: Session for user %s has no socket ID\n", userID)
		}
	}

	fmt.Printf("DEBUG: sendDirectToSockets completed for user %s, sent to %d sockets\n", userID, sentCount)
}
