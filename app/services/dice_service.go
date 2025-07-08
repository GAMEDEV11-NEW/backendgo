package services

import (
	"fmt"
	"gofiber/app/models"
	"gofiber/redis"
	"log"
	"math/rand"
	"time"

	"github.com/gocql/gocql"
)

// DiceService handles all dice-related business logic
type DiceService struct {
	cassandraSession *gocql.Session
	redisService     *redis.Service
}

// NewDiceService creates a new dice service instance
func NewDiceService(cassandraSession *gocql.Session) *DiceService {
	if cassandraSession == nil {
		panic("Cassandra session cannot be nil")
	}
	redisService := redis.NewService()
	return &DiceService{
		cassandraSession: cassandraSession,
		redisService:     redisService,
	}
}

// RollDice generates a random dice number and stores it in the database
func (s *DiceService) RollDice(rollReq models.DiceRollRequest, userID string) (*models.DiceRollResponse, error) {
	// Validate session from Redis
	log.Printf("🔍 DEBUG: Checking session in Redis for token=%s, userID=%s", rollReq.SessionToken, userID)

	// Get session from Redis
	sessionData, err := s.redisService.GetSession(rollReq.SessionToken)
	if err != nil {
		log.Printf("❌ DEBUG: Session not found in Redis: %v", err)
		return nil, fmt.Errorf("invalid or expired session")
	}

	log.Printf("✅ DEBUG: Session found in Redis: %+v", sessionData)

	// Extract session information
	deviceID, ok := sessionData["device_id"].(string)
	if !ok {
		log.Printf("❌ DEBUG: Device ID not found in Redis data")
		return nil, fmt.Errorf("invalid session data")
	}

	// Check if session is active
	isActive, ok := sessionData["is_active"].(bool)
	if !ok || !isActive {
		log.Printf("❌ DEBUG: Session is not active")
		return nil, fmt.Errorf("session is not active")
	}

	// Check if session is expired
	expiresAtStr, ok := sessionData["expires_at"].(string)
	if !ok {
		log.Printf("❌ DEBUG: Expires at not found in Redis data")
		return nil, fmt.Errorf("invalid session data")
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		log.Printf("❌ DEBUG: Failed to parse expires_at: %v", err)
		return nil, fmt.Errorf("invalid session data")
	}

	if time.Now().After(expiresAt) {
		log.Printf("❌ DEBUG: Session is expired. Expires at: %s, Current time: %s",
			expiresAt.Format(time.RFC3339), time.Now().Format(time.RFC3339))
		return nil, fmt.Errorf("session is expired")
	}

	log.Printf("✅ DEBUG: Session validation successful")
	log.Printf("🔍 DEBUG: Device ID comparison - Request: %s, Session: %s", rollReq.DeviceID, deviceID)

	// Validate device ID
	if deviceID != rollReq.DeviceID {
		log.Printf("❌ DEBUG: Device ID mismatch. Expected: %s, Got: %s", deviceID, rollReq.DeviceID)
		return nil, fmt.Errorf("device ID mismatch")
	}

	// Generate random dice number (1-6 for standard dice)
	rand.Seed(time.Now().UnixNano())
	diceNumber := rand.Intn(6) + 1

	// 1. Get or create lookup_dice_id for (game_id, user_id) - ensure only ONE entry per user per game
	var lookupDiceID gocql.UUID
	err = s.cassandraSession.Query(`SELECT dice_id FROM dice_rolls_lookup WHERE game_id = ? AND user_id = ? LIMIT 1`, rollReq.GameID, userID).Scan(&lookupDiceID)
	if err != nil {
		// Not found, create new - this ensures only ONE dice_id per user per game
		lookupDiceID = gocql.TimeUUID()
		err = s.cassandraSession.Query(`INSERT INTO dice_rolls_lookup (game_id, user_id, dice_id, created_at) VALUES (?, ?, ?, ?)`, rollReq.GameID, userID, lookupDiceID, time.Now().UTC()).Exec()
		if err != nil {
			log.Printf("❌ Error creating dice_rolls_lookup: %v", err)
			return nil, fmt.Errorf("failed to create dice_rolls_lookup: %v", err)
		}
		log.Printf("✅ DEBUG: Created new lookup_dice_id: %s for game_id: %s, user_id: %s", lookupDiceID.String(), rollReq.GameID, userID)
	} else {
		log.Printf("✅ DEBUG: Reusing existing lookup_dice_id: %s for game_id: %s, user_id: %s", lookupDiceID.String(), rollReq.GameID, userID)
	}

	// 2. Generate a new roll_id for this roll
	rollID := gocql.TimeUUID()
	rollTime := time.Now().UTC()
	createdAt := rollTime

	// 3. Insert into dice_rolls_data
	err = s.cassandraSession.Query(`
		INSERT INTO dice_rolls_data (lookup_dice_id, roll_id, dice_number, roll_timestamp, session_token, device_id, contest_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`).Bind(lookupDiceID, rollID, diceNumber, rollTime, rollReq.SessionToken, rollReq.DeviceID, rollReq.ContestID, createdAt).Exec()
	if err != nil {
		log.Printf("❌ Error storing dice roll: %v", err)
		return nil, fmt.Errorf("failed to store dice roll: %v", err)
	}

	log.Printf("🎲 Dice roll stored - LookupDiceID: %s, RollID: %s, Number: %d", lookupDiceID.String(), rollID.String(), diceNumber)
	log.Printf("📊 DEBUG: This ensures single dice_id per user/game, multiple rolls stored under same lookup_dice_id")

	// Prepare response data
	responseData := map[string]interface{}{
		"roll_id":        rollID.String(),
		"roll_timestamp": rollTime.Format(time.RFC3339),
		"game_name":      "Dice Game",
		"contest_name":   rollReq.ContestID,
		"is_winner":      diceNumber == 6,
		"bonus_points":   diceNumber * 10,
	}

	return &models.DiceRollResponse{
		Status:       "success",
		Message:      "Dice rolled successfully",
		GameID:       rollReq.GameID,
		UserID:       userID,
		DiceID:       rollID.String(),
		DiceNumber:   diceNumber,
		RollTime:     rollTime.Format(time.RFC3339),
		ContestID:    rollReq.ContestID,
		SessionToken: rollReq.SessionToken,
		DeviceID:     rollReq.DeviceID,
		Data:         responseData,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		SocketID:     "",
		Event:        "dice:roll:response",
	}, nil
}

// GetDiceHistory retrieves dice roll history for a user in a specific game
func (s *DiceService) GetDiceHistory(historyReq models.DiceHistoryRequest) (*models.DiceHistoryResponse, error) {
	// Validate session from Redis
	sessionData, err := s.redisService.GetSession(historyReq.SessionToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Check if session is active
	isActive, ok := sessionData["is_active"].(bool)
	if !ok || !isActive {
		return nil, fmt.Errorf("session is not active")
	}

	// Check if session is expired
	expiresAtStr, ok := sessionData["expires_at"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid session data")
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, fmt.Errorf("invalid session data")
	}

	if time.Now().After(expiresAt) {
		return nil, fmt.Errorf("session is expired")
	}

	// Set default limit if not provided
	limit := historyReq.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}

	// 1. Get lookup_dice_id for (game_id, user_id)
	var lookupDiceID gocql.UUID
	err = s.cassandraSession.Query(`SELECT dice_id FROM dice_rolls_lookup WHERE game_id = ? AND user_id = ? LIMIT 1`, historyReq.GameID, historyReq.UserID).Scan(&lookupDiceID)
	if err != nil {
		return &models.DiceHistoryResponse{
			Status:     "success",
			Message:    "No dice rolls found",
			GameID:     historyReq.GameID,
			UserID:     historyReq.UserID,
			Rolls:      []models.DiceRoll{},
			TotalRolls: 0,
			Data:       map[string]interface{}{"total_rolls": 0, "limit": historyReq.Limit, "game_id": historyReq.GameID, "user_id": historyReq.UserID},
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			SocketID:   "",
			Event:      "dice:history:response",
		}, nil
	}

	// 2. Query all rolls for this lookup_dice_id
	iter := s.cassandraSession.Query(`
		SELECT lookup_dice_id, roll_id, dice_number, roll_timestamp, session_token, device_id, contest_id, created_at
		FROM dice_rolls_data
		WHERE lookup_dice_id = ?
		LIMIT ?
	`, lookupDiceID, limit).Iter()

	var rolls []models.DiceRoll
	var totalRolls int
	for {
		var roll models.DiceRoll
		if !iter.Scan(&roll.LookupDiceID, &roll.RollID, &roll.DiceNumber, &roll.RollTimestamp, &roll.SessionToken, &roll.DeviceID, &roll.ContestID, &roll.CreatedAt) {
			break
		}
		rolls = append(rolls, roll)
		totalRolls++
	}
	iter.Close()

	responseData := map[string]interface{}{
		"total_rolls": totalRolls,
		"limit":       limit,
		"game_id":     historyReq.GameID,
		"user_id":     historyReq.UserID,
	}

	return &models.DiceHistoryResponse{
		Status:     "success",
		Message:    "Dice history retrieved successfully",
		GameID:     historyReq.GameID,
		UserID:     historyReq.UserID,
		Rolls:      rolls,
		TotalRolls: totalRolls,
		Data:       responseData,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		SocketID:   "",
		Event:      "dice:history:response",
	}, nil
}

// GetDiceStats retrieves statistics for dice rolls
func (s *DiceService) GetDiceStats(gameID, userID, sessionToken string) (*models.DiceStats, error) {
	// Validate session from Redis
	sessionData, err := s.redisService.GetSession(sessionToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Check if session is active
	isActive, ok := sessionData["is_active"].(bool)
	if !ok || !isActive {
		return nil, fmt.Errorf("session is not active")
	}

	// Check if session is expired
	expiresAtStr, ok := sessionData["expires_at"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid session data")
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, fmt.Errorf("invalid session data")
	}

	if time.Now().After(expiresAt) {
		return nil, fmt.Errorf("session is expired")
	}

	// Get lookup_dice_id for (game_id, user_id)
	var lookupDiceID gocql.UUID
	err = s.cassandraSession.Query(`SELECT dice_id FROM dice_rolls_lookup WHERE game_id = ? AND user_id = ? LIMIT 1`, gameID, userID).Scan(&lookupDiceID)
	if err != nil {
		return &models.DiceStats{
			TotalRolls:  0,
			AverageRoll: 0,
			HighestRoll: 0,
			LowestRoll:  0,
			RollCounts:  map[int]int{},
			RecentRolls: []models.DiceRoll{},
			GameID:      gameID,
			UserID:      userID,
			ContestID:   "",
		}, nil
	}

	// Query dice roll data for statistics
	var diceNumber int
	var contestID string
	var totalRolls int
	var sumRolls int
	var highestRoll, lowestRoll int
	rollCounts := make(map[int]int)
	var recentRolls []models.DiceRoll

	firstRoll := true
	for {
		err := s.cassandraSession.Query(`
			SELECT dice_number, contest_id
			FROM dice_rolls_data
			WHERE lookup_dice_id = ?
		`, lookupDiceID).Scan(&diceNumber, &contestID)

		if err == nil {
			totalRolls++
			sumRolls += diceNumber
			rollCounts[diceNumber]++

			if firstRoll {
				highestRoll = diceNumber
				lowestRoll = diceNumber
				firstRoll = false
			} else {
				if diceNumber > highestRoll {
					highestRoll = diceNumber
				}
				if diceNumber < lowestRoll {
					lowestRoll = diceNumber
				}
			}
		} else {
			break // No more rolls for this lookup_dice_id
		}
	}

	// Calculate average
	var averageRoll float64
	if totalRolls > 0 {
		averageRoll = float64(sumRolls) / float64(totalRolls)
	}

	// Get recent rolls (last 10) for this lookup_dice_id
	iter := s.cassandraSession.Query(`
		SELECT lookup_dice_id, roll_id, dice_number, roll_timestamp, session_token, device_id, contest_id, created_at
		FROM dice_rolls_data
		WHERE lookup_dice_id = ?
		LIMIT 10
	`, lookupDiceID).Iter()

	for {
		var recentRoll models.DiceRoll
		if !iter.Scan(&recentRoll.LookupDiceID, &recentRoll.RollID, &recentRoll.DiceNumber, &recentRoll.RollTimestamp, &recentRoll.SessionToken, &recentRoll.DeviceID, &recentRoll.ContestID, &recentRoll.CreatedAt) {
			break
		}
		recentRolls = append(recentRolls, recentRoll)
	}
	iter.Close()

	return &models.DiceStats{
		TotalRolls:  totalRolls,
		AverageRoll: averageRoll,
		HighestRoll: highestRoll,
		LowestRoll:  lowestRoll,
		RollCounts:  rollCounts,
		RecentRolls: recentRolls,
		GameID:      gameID,
		UserID:      userID,
		ContestID:   contestID,
	}, nil
}
