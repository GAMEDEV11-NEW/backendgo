package config

import (
	"gofiber/app/models"
	"gofiber/app/services"
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

		// Authenticate user
		user, err := authFunc(socket, "dice:roll")
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
		diceData, ok := event.Data[0].(map[string]interface{})
		if !ok {
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

		// Send successful response to the player who rolled
		response.SocketID = socket.Id
		socket.Emit("dice:roll:response", response)

		// Broadcast dice roll to opponent
		h.broadcastDiceRollToOpponent(gameID, user.ID, response)
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
	// Get opponent user ID from the game
	opponentUserID := h.getOpponentUserID(gameID, userID)
	if opponentUserID == "" {
		return
	}

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

	// Send dice roll update to opponent
	h.sendMessageToUser(opponentUserID, diceUpdate, "opponent:dice:roll:update")

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
	// Query match_pairs table to find the opponent
	var user1ID, user2ID string

	err := h.socketService.GetCassandraSession().Query(`
		SELECT user1_id, user2_id FROM match_pairs 
		WHERE id = ?
	`, gameID).Scan(&user1ID, &user2ID)

	if err != nil {
		// Try to find match pair by user ID instead
		iter := h.socketService.GetCassandraSession().Query(`
			SELECT id, user1_id, user2_id, status FROM match_pairs 
			WHERE user1_id = ? OR user2_id = ?
			ALLOW FILTERING
		`, userID, userID).Iter()

		var matchID, u1ID, u2ID, status string
		var foundMatchPair bool
		for iter.Scan(&matchID, &u1ID, &u2ID, &status) {
			// Check if this is the most recent active match for this user
			// Only consider active match pairs (exclude disconnected, cancelled, completed)
			if status == "active" {
				user1ID = u1ID
				user2ID = u2ID
				foundMatchPair = true

				break
			}
		}
		iter.Close()

		if !foundMatchPair {
			// Check if there are any match pairs in the database at all
			allMatchPairsIter := h.socketService.GetCassandraSession().Query(`
				SELECT id, user1_id, user2_id, status FROM match_pairs 
				LIMIT 10
			`).Iter()

			var allMatchID, allUser1ID, allUser2ID, allStatus string
			for allMatchPairsIter.Scan(&allMatchID, &allUser1ID, &allUser2ID, &allStatus) {
				// Check if this match pair contains our user and is active
				// Only consider active match pairs (exclude disconnected, cancelled, completed)
				if (allUser1ID == userID || allUser2ID == userID) && allStatus == "active" {
					user1ID = allUser1ID
					user2ID = allUser2ID
					foundMatchPair = true

					break
				}
			}
			allMatchPairsIter.Close()

			if !foundMatchPair {
				return ""
			}
		}
	}

	// Determine which user is the opponent
	if user1ID == userID {
		return user2ID
	} else if user2ID == userID {
		return user1ID
	}
	return ""
}

// sendMessageToUser sends a message to a specific user using available services
func (h *GameboardSocketHandler) sendMessageToUser(userID string, message map[string]interface{}, eventName string) {
	// Try messaging service first
	if h.socketService.GetMessagingService() != nil {
		messageData := services.MessageData{
			Event:     eventName,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Priority:  "high",
			Data:      message,
		}

		err := h.socketService.GetMessagingService().SendMessageToUser(userID, messageData)
		if err != nil {
			// Try direct socket emission as fallback
			h.sendDirectToSockets(userID, eventName, message)
		}
	} else {
		// Try direct socket emission as fallback
		h.sendDirectToSockets(userID, eventName, message)
	}
}

// sendDirectToSockets sends a message directly to all sockets of a user
func (h *GameboardSocketHandler) sendDirectToSockets(userID string, eventName string, message map[string]interface{}) {
	// Get all sessions for this user
	sessions, err := h.socketService.GetSessionService().GetSessionsByUserID(userID)
	if err != nil {
		return
	}

	if len(sessions) == 0 {
		return
	}

	var sentCount int
	for _, session := range sessions {
		if session.SocketID != "" {
			// Try to emit directly to the socket using Socket.IO instance
			if h.socketService.GetIo() != nil {
				// Get all connected sockets
				sockets := h.socketService.GetIo().Sockets()

				// Find the specific socket and emit the message
				for _, socket := range sockets {
					if socket.Id == session.SocketID {
						socket.Emit(eventName, message)
						sentCount++
						break
					}
				}
			} else if h.socketService.GetMessagingService() != nil {
				// Fallback to messaging service
				err := h.socketService.GetMessagingService().SendMessageToSocket(session.SocketID, services.MessageData{
					Event:     eventName,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Priority:  "high",
					Data:      message,
				})
				if err == nil {
					sentCount++
				}
			} else {
				sentCount++
			}
		}
	}
}
