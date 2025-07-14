package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gofiber/redis"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

// GamePiecesService handles all game pieces related operations
type GamePiecesService struct {
	cassandraSession *gocql.Session
	redisService     *redis.Service
}

// NewGamePiecesService creates a new game pieces service instance
func NewGamePiecesService(cassandraSession *gocql.Session) *GamePiecesService {
	if cassandraSession == nil {
		panic("Cassandra session cannot be nil")
	}
	return &GamePiecesService{
		cassandraSession: cassandraSession,
		redisService:     redis.NewService(),
	}
}

// CreatePiecesForMatch creates four pieces for each user when they are matched
func (s *GamePiecesService) CreatePiecesForMatch(gameID, user1ID, user2ID string) error {

	// Check if pieces already exist for this game
	existingPieces1, err1 := s.GetUserPieces(gameID, user1ID)
	existingPieces2, err2 := s.GetUserPieces(gameID, user2ID)

	// If either user already has pieces, don't create duplicates
	if (err1 == nil && len(existingPieces1) > 0) || (err2 == nil && len(existingPieces2) > 0) {
		// Pieces already exist, don't create duplicates
		return nil
	}

	// Create four pieces for user 1
	if err := s.createUserPieces(gameID, user1ID, "player1"); err != nil {
		return fmt.Errorf("failed to create pieces for user 1: %v", err)
	}

	// Create four pieces for user 2
	if err := s.createUserPieces(gameID, user2ID, "player2"); err != nil {
		return fmt.Errorf("failed to create pieces for user 2: %v", err)
	}

	return nil
}

// createUserPieces creates four pieces for a single user
func (s *GamePiecesService) createUserPieces(gameID, userID, playerID string) error {
	now := time.Now()

	// Create four pieces for the user
	for i := 1; i <= 4; i++ {
		pieceID := uuid.New().String()
		pieceType := fmt.Sprintf("piece_%d", i)

		// Create initial game piece entry (move_number = 0 for initial setup)
		err := s.cassandraSession.Query(`
			INSERT INTO game_pieces (
				game_id, user_id, move_number, piece_id, player_id,
				from_pos_last, to_pos_last, piece_type, captured_piece,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, gameID, userID, 0, pieceID, playerID,
			"", "initial", pieceType, "", now, now).Exec()

		if err != nil {
			return fmt.Errorf("failed to create piece %d: %v", i, err)
		}

	}

	return nil
}

// RecordPieceMove records a new move for a game piece
func (s *GamePiecesService) RecordPieceMove(gameID, userID, pieceID, fromPos, toPos, pieceType, capturedPiece string) error {
	now := time.Now()

	// Check if this piece already exists
	var existingMoveNumber int
	var existingFromPos, existingToPos, existingPieceType string

	err := s.cassandraSession.Query(`
		SELECT move_number, from_pos_last, to_pos_last, piece_type 
		FROM game_pieces 
		WHERE game_id = ? AND user_id = ? AND move_number = 0 AND piece_id = ?
	`, gameID, userID, pieceID).Scan(&existingMoveNumber, &existingFromPos, &existingToPos, &existingPieceType)

	if err == gocql.ErrNotFound {
		// This piece doesn't exist yet, create initial entry
		err = s.cassandraSession.Query(`
			INSERT INTO game_pieces (
				game_id, user_id, move_number, piece_id, player_id,
				from_pos_last, to_pos_last, piece_type, captured_piece,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, gameID, userID, 0, pieceID, userID,
			"", "initial", pieceType, "", now, now).Exec()

		if err != nil {
			return fmt.Errorf("failed to insert initial piece: %v", err)
		}

	} else if err != nil {
		return fmt.Errorf("failed to check existing piece: %v", err)
	}

	// Update the existing piece instead of creating a new move entry
	err = s.cassandraSession.Query(`
		UPDATE game_pieces SET 
			from_pos_last = ?, to_pos_last = ?, piece_type = ?, captured_piece = ?, updated_at = ?
		WHERE game_id = ? AND user_id = ? AND move_number = 0 AND piece_id = ?
	`, fromPos, toPos, pieceType, capturedPiece, now, gameID, userID, pieceID).Exec()

	if err != nil {
		return fmt.Errorf("failed to update piece move: %v", err)
	}

	// Store move data in Redis instead of updating piece_moves table immediately
	err = s.storePieceMoveInRedis(gameID, userID, pieceID, fromPos, toPos, pieceType, capturedPiece, now)
	if err != nil {
		return fmt.Errorf("failed to store piece move in Redis: %v", err)
	}

	return nil
}

// storePieceMoveInRedis stores piece move data in Redis for fast access during game
func (s *GamePiecesService) storePieceMoveInRedis(gameID, userID, pieceID, fromPos, toPos, pieceType, capturedPiece string, moveTime time.Time) error {
	// Create Redis key for this piece
	redisKey := fmt.Sprintf("piece_moves:%s:%s:%s", gameID, userID, pieceID)

	// Get existing piece data from Redis or create new
	pieceData, err := s.getPieceDataFromRedis(redisKey)
	if err != nil {

		// Create new piece data
		pieceData = map[string]interface{}{
			"game_id":        gameID,
			"user_id":        userID,
			"piece_id":       pieceID,
			"total_moves":    0,
			"last_position":  "initial",
			"last_move_time": moveTime.Format(time.RFC3339),
			"player_id":      userID,
			"piece_type":     pieceType,
			"current_state": map[string]interface{}{
				"position":                toPos,
				"position_number":         s.parsePositionNumber(toPos),
				"status":                  "active",
				"moves":                   1,
				"total_positions_visited": []string{s.parsePositionNumber(toPos)},
			},
			"move_history": []map[string]interface{}{
				{
					"move_number":     1,
					"from_pos":        fromPos,
					"to_pos":          toPos,
					"position_number": s.parsePositionNumber(toPos),
					"timestamp":       moveTime.Format(time.RFC3339),
				},
			},
			"piece_metadata": map[string]interface{}{
				"created_by":              userID,
				"game_type":               "standard",
				"piece_value":             1,
				"current_position_number": s.parsePositionNumber(toPos),
				"total_positions":         57,
			},
			"created_at": moveTime.Format(time.RFC3339),
			"updated_at": moveTime.Format(time.RFC3339),
		}
	} else {

		// Update existing piece data
		totalMovesFloat, ok := pieceData["total_moves"].(float64)
		if !ok {
			// Try int type as fallback
			if totalMovesInt, ok := pieceData["total_moves"].(int); ok {
				totalMovesFloat = float64(totalMovesInt)
			} else {

				return fmt.Errorf("invalid total_moves type in piece data")
			}
		}
		totalMoves := int(totalMovesFloat) + 1
		positionNumber := s.parsePositionNumber(toPos)

		// Update current state
		currentState, ok := pieceData["current_state"].(map[string]interface{})
		if !ok {

			return fmt.Errorf("invalid current_state type in piece data")
		}
		currentState["position"] = toPos
		currentState["position_number"] = positionNumber
		currentState["last_move_time"] = moveTime.Format(time.RFC3339)
		currentState["moves"] = totalMoves

		// Update move history (keep last 10 moves)
		var moveHistory []map[string]interface{}

		// Handle different possible types for move_history
		switch v := pieceData["move_history"].(type) {
		case []map[string]interface{}:
			moveHistory = v
		case []interface{}:
			// Convert []interface{} to []map[string]interface{}
			for _, item := range v {
				if moveMap, ok := item.(map[string]interface{}); ok {
					moveHistory = append(moveHistory, moveMap)
				}
			}
		default:
			log.Printf("[REDIS_STORE] Warning: unexpected move_history type: %T, creating new history", v)
			moveHistory = []map[string]interface{}{}
		}

		newMove := map[string]interface{}{
			"move_number":     totalMoves,
			"from_pos":        fromPos,
			"to_pos":          toPos,
			"position_number": positionNumber,
			"timestamp":       moveTime.Format(time.RFC3339),
		}
		moveHistory = append(moveHistory, newMove)

		// Keep only last 10 moves
		if len(moveHistory) > 10 {
			moveHistory = moveHistory[len(moveHistory)-10:]
		}

		// Update piece data
		pieceData["total_moves"] = totalMoves
		pieceData["last_position"] = toPos
		pieceData["last_move_time"] = moveTime.Format(time.RFC3339)
		pieceData["current_state"] = currentState
		pieceData["move_history"] = moveHistory
		pieceData["updated_at"] = moveTime.Format(time.RFC3339)
	}

	// Store updated piece data in Redis
	err = s.savePieceDataToRedis(redisKey, pieceData)
	if err != nil {
		return fmt.Errorf("failed to save piece data to Redis: %v", err)
	}

	return nil
}

// getPieceDataFromRedis retrieves piece data from Redis
func (s *GamePiecesService) getPieceDataFromRedis(redisKey string) (map[string]interface{}, error) {
	var pieceData map[string]interface{}
	err := s.redisService.Get(redisKey, &pieceData)
	if err != nil {

		return nil, err
	}

	// Validate piece data structure
	err = s.validateAndCleanPieceData(pieceData)
	if err != nil {

		return nil, err
	}

	return pieceData, nil
}

// savePieceDataToRedis saves piece data to Redis
func (s *GamePiecesService) savePieceDataToRedis(redisKey string, pieceData map[string]interface{}) error {
	// Store with 24 hour expiration (game sessions typically don't last longer)
	expiration := 24 * time.Hour
	err := s.redisService.Set(redisKey, pieceData, expiration)
	if err != nil {
		return fmt.Errorf("failed to save piece data to Redis: %v", err)
	}

	_, _ = json.Marshal(pieceData)

	return nil
}

// parsePositionNumber extracts position number from position string
func (s *GamePiecesService) parsePositionNumber(position string) string {
	if strings.Contains(position, "total") {
		parts := strings.Fields(position)
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	return position
}

// SaveGamePiecesToDatabase saves all piece moves from Redis to database at game end
func (s *GamePiecesService) SaveGamePiecesToDatabase(gameID string) error {

	// Get all piece keys for this game from Redis
	pieceKeys, err := s.getPieceKeysForGame(gameID)
	if err != nil {
		return fmt.Errorf("failed to get piece keys: %v", err)
	}

	// Save each piece's data to piece_moves table
	for _, pieceKey := range pieceKeys {
		pieceData, err := s.getPieceDataFromRedis(pieceKey)
		if err != nil {
			continue
		}

		err = s.savePieceDataToDatabase(pieceData)
		if err != nil {
			continue
		}

		// Remove from Redis after successful save
		s.removePieceDataFromRedis(pieceKey)
	}

	return nil
}

// getPieceKeysForGame gets all Redis keys for pieces in a game
func (s *GamePiecesService) getPieceKeysForGame(gameID string) ([]string, error) {
	// Use Redis SCAN to find all keys matching the pattern
	pattern := fmt.Sprintf("piece_moves:%s:*", gameID)

	// For now, we'll use a simple approach - scan for keys
	// In a production environment, you might want to maintain a set of keys
	var keys []string

	// Get all keys matching the pattern
	iter := s.redisService.GetClient().Scan(s.redisService.GetContext(), 0, pattern, 100).Iterator()
	for iter.Next(s.redisService.GetContext()) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan Redis keys: %v", err)
	}

	return keys, nil
}

// savePieceDataToDatabase saves a single piece's data to piece_moves table
func (s *GamePiecesService) savePieceDataToDatabase(pieceData map[string]interface{}) error {
	gameID := pieceData["game_id"].(string)
	userID := pieceData["user_id"].(string)
	pieceID := pieceData["piece_id"].(string)

	// Handle total_moves type conversion
	var totalMoves int
	if totalMovesFloat, ok := pieceData["total_moves"].(float64); ok {
		totalMoves = int(totalMovesFloat)
	} else if totalMovesInt, ok := pieceData["total_moves"].(int); ok {
		totalMoves = totalMovesInt
	} else {
		return fmt.Errorf("invalid total_moves type in piece data")
	}

	lastPosition := pieceData["last_position"].(string)
	lastMoveTimeStr := pieceData["last_move_time"].(string)
	playerID := pieceData["player_id"].(string)
	pieceType := pieceData["piece_type"].(string)

	// Parse last move time
	lastMoveTime, err := time.Parse(time.RFC3339, lastMoveTimeStr)
	if err != nil {
		return fmt.Errorf("failed to parse last move time: %v", err)
	}

	// Marshal JSON fields
	currentStateJSON, _ := json.Marshal(pieceData["current_state"])
	moveHistoryJSON, _ := json.Marshal(pieceData["move_history"])
	pieceMetadataJSON, _ := json.Marshal(pieceData["piece_metadata"])

	// Insert into piece_moves table
	err = s.cassandraSession.Query(`
		INSERT INTO piece_moves (
			game_id, user_id, piece_id, total_moves, last_position,
			last_move_time, player_id, piece_type, current_state,
			move_history, piece_metadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, gameID, userID, pieceID, totalMoves, lastPosition, lastMoveTime, playerID, pieceType,
		string(currentStateJSON), string(moveHistoryJSON), string(pieceMetadataJSON),
		time.Now(), time.Now()).Exec()

	if err != nil {
		return fmt.Errorf("failed to insert piece data to database: %v", err)
	}

	return nil
}

// removePieceDataFromRedis removes piece data from Redis after saving to database
func (s *GamePiecesService) removePieceDataFromRedis(redisKey string) error {
	err := s.redisService.Delete(redisKey)
	if err != nil {
		return fmt.Errorf("failed to remove Redis key %s: %v", redisKey, err)
	}

	return nil
}

// GetUserPieces retrieves all pieces for a specific game and user
func (s *GamePiecesService) GetUserPieces(gameID, userID string) ([]map[string]interface{}, error) {

	var pieces []map[string]interface{}

	query := `SELECT game_id, user_id, move_number, piece_id, player_id,
			   from_pos_last, to_pos_last, piece_type, captured_piece,
			   created_at, updated_at
		FROM game_pieces 
		WHERE game_id = ? AND user_id = ?
		ORDER BY move_number ASC`

	iter := s.cassandraSession.Query(query, gameID, userID).Iter()

	var gameIDVal, userIDVal, moveNumber, pieceID, playerID, fromPosLast, toPosLast, pieceType, capturedPiece string
	var createdAt, updatedAt time.Time

	scanCount := 0
	for iter.Scan(&gameIDVal, &userIDVal, &moveNumber, &pieceID, &playerID,
		&fromPosLast, &toPosLast, &pieceType, &capturedPiece,
		&createdAt, &updatedAt) {

		scanCount++

		piece := map[string]interface{}{
			"game_id":        gameIDVal,
			"user_id":        userIDVal,
			"move_number":    moveNumber,
			"piece_id":       pieceID,
			"player_id":      playerID,
			"from_pos_last":  fromPosLast,
			"to_pos_last":    toPosLast,
			"piece_type":     pieceType,
			"captured_piece": capturedPiece,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
		}
		pieces = append(pieces, piece)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to retrieve user pieces: %v", err)
	}

	return pieces, nil
}

// GetUserPiecesFromPieceMoves retrieves current piece state from piece_moves table using gameID and userID
func (s *GamePiecesService) GetUserPiecesFromPieceMoves(gameID, userID string) ([]map[string]interface{}, error) {

	var pieces []map[string]interface{}

	query := `SELECT game_id, user_id, piece_id, total_moves, last_position,
			   last_move_time, player_id, current_state,
			   move_history, piece_metadata, created_at, updated_at
		FROM piece_moves 
		WHERE game_id = ? AND user_id = ?`

	iter := s.cassandraSession.Query(query, gameID, userID).Iter()

	var gameIDVal, userIDVal, pieceID, lastPosition, playerID, currentState, moveHistory, pieceMetadata string
	var totalMoves int
	var lastMoveTime, createdAt, updatedAt time.Time

	scanCount := 0
	for iter.Scan(&gameIDVal, &userIDVal, &pieceID, &totalMoves, &lastPosition,
		&lastMoveTime, &playerID, &currentState, &moveHistory, &pieceMetadata,
		&createdAt, &updatedAt) {

		scanCount++

		piece := map[string]interface{}{
			"game_id":        gameIDVal,
			"user_id":        userIDVal,
			"piece_id":       pieceID,
			"total_moves":    totalMoves,
			"last_position":  lastPosition,
			"last_move_time": lastMoveTime,
			"player_id":      playerID,
			"current_state":  currentState,
			"move_history":   moveHistory,
			"piece_metadata": pieceMetadata,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
		}
		pieces = append(pieces, piece)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to retrieve user pieces from piece_moves: %v", err)
	}

	return pieces, nil
}

// createNewPieceMoveEntry creates a new piece_moves entry for a piece that hasn't moved before
func (s *GamePiecesService) createNewPieceMoveEntry(gameID, userID, pieceID, position string, moveTime time.Time, pieceType string) error {
	now := time.Now()

	// Parse position number (e.g., "57" from "total 57")
	positionNumber := "0"
	if strings.Contains(position, "total") {
		parts := strings.Fields(position)
		if len(parts) >= 2 {
			positionNumber = parts[1]
		}
	} else {
		positionNumber = position
	}

	// Create initial state
	currentState := map[string]interface{}{
		"position":                position,
		"position_number":         positionNumber,
		"status":                  "active",
		"moves":                   1,
		"total_positions_visited": []string{positionNumber},
	}
	currentStateJSON, _ := json.Marshal(currentState)

	// Create move history
	moveHistory := []map[string]interface{}{
		{
			"move_number":     1,
			"from_pos":        "initial",
			"to_pos":          position,
			"position_number": positionNumber,
			"timestamp":       moveTime.Format(time.RFC3339),
		},
	}
	moveHistoryJSON, _ := json.Marshal(moveHistory)

	// Create piece metadata
	pieceMetadata := map[string]interface{}{
		"created_by":              userID,
		"game_type":               "standard",
		"piece_value":             1,
		"current_position_number": positionNumber,
		"total_positions":         57, // Total positions available
	}
	pieceMetadataJSON, _ := json.Marshal(pieceMetadata)

	// Insert new piece_moves entry
	err := s.cassandraSession.Query(`
		INSERT INTO piece_moves (
			game_id, user_id, piece_id, total_moves, last_position,
			last_move_time, player_id, current_state,
			move_history, piece_metadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, gameID, userID, pieceID, 1, position, moveTime, userID,
		string(currentStateJSON), string(moveHistoryJSON), string(pieceMetadataJSON), now, now).Exec()

	if err != nil {
		return fmt.Errorf("failed to create new piece_moves entry: %v", err)
	}

	return nil
}

// GetUserPiecesCurrentState retrieves only the current state of pieces (latest move for each piece)
func (s *GamePiecesService) GetUserPiecesCurrentState(gameID, userID string) ([]map[string]interface{}, error) {

	// First try to get pieces from Redis (fast access during game)
	redisPieces, err := s.getUserPiecesFromRedis(gameID, userID)
	if err == nil && len(redisPieces) > 0 {
		return redisPieces, nil
	}

	// Fallback to database if Redis doesn't have data

	// Get all pieces for this user in this game (only move_number = 0 since we update existing pieces)
	iter := s.cassandraSession.Query(`
		SELECT game_id, user_id, move_number, piece_id, player_id,
			   from_pos_last, to_pos_last, piece_type, captured_piece,
			   created_at, updated_at
		FROM game_pieces 
		WHERE game_id = ? AND user_id = ? AND move_number = 0
	`, gameID, userID).Iter()

	var pieces []map[string]interface{}

	var gameIDVal, userIDVal, moveNumber, pieceID, playerID, fromPosLast, toPosLast, pieceType, capturedPiece string
	var createdAt, updatedAt time.Time

	for iter.Scan(&gameIDVal, &userIDVal, &moveNumber, &pieceID, &playerID,
		&fromPosLast, &toPosLast, &pieceType, &capturedPiece,
		&createdAt, &updatedAt) {

		piece := map[string]interface{}{
			"game_id":        gameIDVal,
			"user_id":        userIDVal,
			"move_number":    moveNumber,
			"piece_id":       pieceID,
			"player_id":      playerID,
			"from_pos_last":  fromPosLast,
			"to_pos_last":    toPosLast,
			"piece_type":     pieceType,
			"captured_piece": capturedPiece,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
		}
		pieces = append(pieces, piece)
	}
	iter.Close()

	return pieces, nil
}

// getUserPiecesFromRedis gets current piece state from Redis
func (s *GamePiecesService) getUserPiecesFromRedis(gameID, userID string) ([]map[string]interface{}, error) {
	// Get all piece keys for this user in this game
	pattern := fmt.Sprintf("piece_moves:%s:%s:*", gameID, userID)

	var pieces []map[string]interface{}

	// Scan for keys matching the pattern
	iter := s.redisService.GetClient().Scan(s.redisService.GetContext(), 0, pattern, 100).Iterator()
	for iter.Next(s.redisService.GetContext()) {
		redisKey := iter.Val()

		// Get piece data from Redis
		pieceData, err := s.getPieceDataFromRedis(redisKey)
		if err != nil {
			continue
		}

		// Convert to the expected format
		piece := map[string]interface{}{
			"game_id":        pieceData["game_id"],
			"user_id":        pieceData["user_id"],
			"move_number":    "0", // Always 0 since we update existing pieces
			"piece_id":       pieceData["piece_id"],
			"player_id":      pieceData["player_id"],
			"from_pos_last":  "previous_position",
			"to_pos_last":    pieceData["last_position"],
			"piece_type":     pieceData["piece_type"],
			"captured_piece": "",
			"created_at":     pieceData["created_at"],
			"updated_at":     pieceData["updated_at"],
		}
		pieces = append(pieces, piece)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan Redis keys: %v", err)
	}

	return pieces, nil
}

// validateAndCleanPieceData validates and cleans up piece data from Redis
func (s *GamePiecesService) validateAndCleanPieceData(pieceData map[string]interface{}) error {
	if pieceData == nil {
		return fmt.Errorf("piece data is nil")
	}

	// Ensure all required fields exist
	requiredFields := []string{"game_id", "user_id", "piece_id", "total_moves", "last_position", "last_move_time", "player_id", "piece_type"}
	for _, field := range requiredFields {
		if _, exists := pieceData[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Ensure total_moves is a number
	if _, ok := pieceData["total_moves"].(float64); !ok {
		if _, ok := pieceData["total_moves"].(int); !ok {
			return fmt.Errorf("total_moves must be a number")
		}
	}

	// Ensure current_state is a map
	if _, ok := pieceData["current_state"].(map[string]interface{}); !ok {
		return fmt.Errorf("current_state must be a map")
	}

	// Ensure move_history is a slice
	if _, ok := pieceData["move_history"].([]map[string]interface{}); !ok {
		if _, ok := pieceData["move_history"].([]interface{}); !ok {
			return fmt.Errorf("move_history must be a slice")
		}
	}

	return nil
}
