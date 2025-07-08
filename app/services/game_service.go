package services

import (
	"encoding/json"
	"fmt"
	"gofiber/app/models"
	"gofiber/app/utils"
	"gofiber/redis"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GameService handles all game and contest management business logic
type GameService struct {
	cassandraSession *gocql.Session
	redisService     *redis.Service
}

// LeagueJoin struct moved to models package

// NewGameService creates a new game service instance using Cassandra
func NewGameService(cassandraSession *gocql.Session) *GameService {
	if cassandraSession == nil {
		panic("Cassandra session cannot be nil")
	}
	redisService := redis.NewService()
	service := &GameService{
		cassandraSession: cassandraSession,
		redisService:     redisService,
	}
	return service
}

// HandlePlayerAction processes player actions in gameplay
func (s *GameService) HandlePlayerAction(actionReq models.PlayerActionRequest) (*models.PlayerActionResponse, error) {
	// Validate session
	var session models.Session
	err := s.cassandraSession.Query(`
		SELECT session_token, mobile_no, is_active, expires_at
		FROM sessions
		WHERE session_token = ? AND mobile_no = ? AND is_active = true AND expires_at > ?
	`).Bind(actionReq.SessionToken, actionReq.PlayerID, true, time.Now()).Scan(&session.SessionToken, &session.MobileNo, &session.IsActive, &session.ExpiresAt)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Generate action ID
	actionID := primitive.NewObjectID().Hex()

	// Process action based on type
	switch actionReq.ActionType {
	case "move":
		// Validate coordinates
		if actionReq.Coordinates.X < 0 || actionReq.Coordinates.Y < 0 {
			return nil, fmt.Errorf("invalid coordinates")
		}

		// Calculate distance moved (for demo purposes)
		_ = math.Sqrt(float64(actionReq.Coordinates.X*actionReq.Coordinates.X + actionReq.Coordinates.Y*actionReq.Coordinates.Y))

	case "attack":
		// Handle attack action
		_ = actionReq.GameState.Health

	case "collect":
		// Handle collect action
		_ = actionReq.GameState.Level

	default:
		return nil, fmt.Errorf("unknown action type: %s", actionReq.ActionType)
	}

	return &models.PlayerActionResponse{
		Success:  true,
		Message:  "Action processed successfully",
		ActionID: actionID,
	}, nil
}

// HandleHeartbeat processes heartbeat from client
func (s *GameService) HandleHeartbeat() models.HeartbeatResponse {
	return models.HeartbeatResponse{
		Success:   true,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// HandleWelcome sends welcome message to client
func (s *GameService) HandleWelcome() models.WelcomeResponse {
	return models.WelcomeResponse{
		Success: true,
		Status:  "connected",
		Message: "Welcome to the game server!",
		ServerInfo: map[string]interface{}{
			"version":     "1.0.0",
			"server_time": time.Now().Format(time.RFC3339),
			"features":    []string{"real-time", "multiplayer", "chat"},
		},
	}
}

// HandleHealthCheck processes health check request
func (s *GameService) HandleHealthCheck() models.HealthCheckResponse {
	return models.HealthCheckResponse{
		Success:   true,
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// HandleStaticMessage handles static message requests including game list
func (s *GameService) HandleStaticMessage(staticReq models.StaticMessageRequest) (*models.StaticMessageResponse, error) {
	// Validate session
	var session models.Session
	err := s.cassandraSession.Query(`
		SELECT session_token, mobile_no, is_active, expires_at
		FROM sessions
		WHERE session_token = ? AND mobile_no = ? AND is_active = true AND expires_at > ?
	`).Bind(staticReq.SessionToken, staticReq.MobileNo, true, time.Now()).Scan(&session.SessionToken, &session.MobileNo, &session.IsActive, &session.ExpiresAt)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Prepare response data - only use sub_list
	responseData := s.getGameSubListData()

	return &models.StaticMessageResponse{
		Status:       "success",
		Message:      "Static message retrieved successfully",
		MobileNo:     staticReq.MobileNo,
		SessionToken: staticReq.SessionToken,
		MessageType:  staticReq.MessageType,
		Data:         responseData,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		SocketID:     "",
		Event:        "static_message",
	}, nil
}

// GetGameListData returns sample game list data with Redis caching
func (s *GameService) getGameListData() map[string]interface{} {
	// Try to get from Redis cache first
	if s.redisService != nil {
		cachedData, err := s.redisService.GetGameList()
		if err == nil {
			return cachedData
		}
	}

	// Generate fresh game list data
	gamelist := []map[string]interface{}{
		{
			"game_id":          "1",
			"game_name":        "Simple Ludo",
			"game_type":        "board",
			"active_gamepalye": 1500,
			"livegameplaye":    800,
			"status":           "active",
			"created_at":       time.Now().UTC().Format(time.RFC3339),
		},
		{
			"game_id":          "2",
			"game_name":        "Classic Ludo",
			"game_type":        "board",
			"active_gamepalye": 1000,
			"livegameplaye":    600,
			"status":           "active",
			"created_at":       time.Now().UTC().Format(time.RFC3339),
		},
		{
			"game_id":          "3",
			"game_name":        "Quick Ludo",
			"game_type":        "board",
			"active_gamepalye": 2000,
			"livegameplaye":    1200,
			"status":           "active",
			"created_at":       time.Now().UTC().Format(time.RFC3339),
		},
	}

	gameListData := map[string]interface{}{
		"gamelist": gamelist,
	}

	// Cache the data in Redis for 5 minutes
	if s.redisService != nil {
		err := s.redisService.CacheGameList(gameListData, 5*time.Minute)
		if err != nil {
			return gameListData
		}
	}

	return gameListData
}

// GetGameListFromRedis returns game list data from Redis
func (s *GameService) GetGameListFromRedis() (map[string]interface{}, error) {
	if s.redisService == nil {
		return nil, fmt.Errorf("Redis service not available")
	}
	return s.redisService.GetGameList()
}

// GetGameListDataPublic returns game list data (public method)
func (s *GameService) GetGameListDataPublic() map[string]interface{} {
	return s.getGameListData()
}

// convertContestDataTypes converts float64 values back to integers in contest data
func (s *GameService) convertContestDataTypes(data map[string]interface{}) map[string]interface{} {
	if gamelist, exists := data["gamelist"]; exists {
		if contests, ok := gamelist.([]interface{}); ok {
			for _, contest := range contests {
				if contestMap, ok := contest.(map[string]interface{}); ok {
					// Convert contest_id
					if contestID, exists := contestMap["contest_id"]; exists {
						if floatID, ok := contestID.(float64); ok {
							contestMap["contest_id"] = int(floatID)
						}
					}

					// Convert contest_joinuser
					if joinUser, exists := contestMap["contest_joinuser"]; exists {
						if floatJoin, ok := joinUser.(float64); ok {
							contestMap["contest_joinuser"] = int(floatJoin)
						}
					}

					// Convert contest_activeuser
					if activeUser, exists := contestMap["contest_activeuser"]; exists {
						if floatActive, ok := activeUser.(float64); ok {
							contestMap["contest_activeuser"] = int(floatActive)
						}
					}

					// Convert contest_win_price (keep as interface{} to support both int and float)
					if winPrice, exists := contestMap["contest_win_price"]; exists {
						if floatPrice, ok := winPrice.(float64); ok {
							// Check if it's a whole number
							if floatPrice == float64(int(floatPrice)) {
								contestMap["contest_win_price"] = int(floatPrice)
							} else {
								contestMap["contest_win_price"] = floatPrice
							}
						}
					}

					// Convert contest_entryfee (keep as interface{} to support both int and float)
					if entryFee, exists := contestMap["contest_entryfee"]; exists {
						if floatFee, ok := entryFee.(float64); ok {
							// Check if it's a whole number
							if floatFee == float64(int(floatFee)) {
								contestMap["contest_entryfee"] = int(floatFee)
							} else {
								contestMap["contest_entryfee"] = floatFee
							}
						}
					}
				}
			}
		}
	}
	return data
}

// getGameListData returns sample game list data with Redis caching
func (s *GameService) getGameSubListData() map[string]interface{} {

	// Try to get from Redis cache first
	if s.redisService != nil {
		cachedData, err := s.redisService.GetListContest()
		if err == nil {
			// Convert data types after retrieving from cache
			convertedData := s.convertContestDataTypes(cachedData)
			return convertedData
		}
	} else {
	}

	// Generate fresh game list data with proper integer types
	gamelist := []map[string]interface{}{
		{
			"contest_id":         1,
			"contest_name":       "Weekly Algorithm Challenge",
			"contest_win_price":  5000,
			"contest_entryfee":   50,
			"contest_joinuser":   1000,
			"contest_activeuser": 847,
			"contest_starttime":  "2025-07-01T09:00:00Z",
			"contest_endtime":    "2025-07-07T23:59:59Z",
		},
		{
			"contest_id":         2,
			"contest_name":       "Data Science Hackathon",
			"contest_win_price":  10000,
			"contest_entryfee":   100,
			"contest_joinuser":   500,
			"contest_activeuser": 423,
			"contest_starttime":  "2025-06-15T10:00:00Z",
			"contest_endtime":    "2025-06-20T18:00:00Z",
		},
		{
			"contest_id":         3,
			"contest_name":       "Frontend Development Sprint",
			"contest_win_price":  3000,
			"contest_entryfee":   30,
			"contest_joinuser":   300,
			"contest_activeuser": 298,
			"contest_starttime":  "2025-06-10T08:00:00Z",
			"contest_endtime":    "2025-06-12T20:00:00Z",
		},
		{
			"contest_id":         4,
			"contest_name":       "Mobile App Innovation",
			"contest_win_price":  8000,
			"contest_entryfee":   80,
			"contest_joinuser":   800,
			"contest_activeuser": 156,
			"contest_starttime":  "2025-07-15T09:00:00Z",
			"contest_endtime":    "2025-07-25T23:59:59Z",
		},
		{
			"contest_id":         5,
			"contest_name":       "Blockchain Smart Contract Challenge",
			"contest_win_price":  5.0,
			"contest_entryfee":   0.1,
			"contest_joinuser":   400,
			"contest_activeuser": 89,
			"contest_starttime":  "2025-07-05T10:00:00Z",
			"contest_endtime":    "2025-07-10T18:00:00Z",
		},
		{
			"contest_id":         6,
			"contest_name":       "AI Chatbot Competition",
			"contest_win_price":  4000,
			"contest_entryfee":   40,
			"contest_joinuser":   600,
			"contest_activeuser": 567,
			"contest_starttime":  "2025-06-25T09:00:00Z",
			"contest_endtime":    "2025-06-27T17:00:00Z",
		},
	}

	gameListData := map[string]interface{}{
		"gamelist": gamelist,
	}

	// Cache the data in Redis for 5 minutes
	if s.redisService != nil {
		err := s.redisService.CacheListContest(gameListData, 5*time.Minute)

		if err != nil {
			return gameListData
		}
	}
	return gameListData
}

// Helper function to get map keys for debugging

// HandleMainScreen handles main screen requests with authentication validation
func (s *GameService) HandleMainScreen(mainReq models.MainScreenRequest) (*models.MainScreenResponse, error) {
	// Decrypt the simple JWT token to get the original values used during token creation
	simpleJWTData, err := utils.ValidateSimpleJWTToken(mainReq.JWTToken)
	if err != nil {
		return nil, fmt.Errorf("simple JWT token validation failed: %v", err)
	}

	// Extract values from the decrypted JWT token
	tokenMobileNo := simpleJWTData.MobileNo
	tokenDeviceID := simpleJWTData.DeviceID
	tokenFCMToken := simpleJWTData.FCMToken

	// Validate mobile number from token
	if len(tokenMobileNo) < 10 {
		return nil, fmt.Errorf("invalid mobile number in JWT token")
	}

	// Validate device ID from token
	if len(tokenDeviceID) < 1 {
		return nil, fmt.Errorf("invalid device ID in JWT token")
	}

	// Check if user exists and is active using token mobile number
	var user models.User
	err = s.cassandraSession.Query(`
		SELECT id, mobile_no, full_name, status, language_code
		FROM users
		WHERE mobile_no = ?
		ALLOW FILTERING
	`).Bind(tokenMobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active, jwt_token
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
		ALLOW FILTERING
	`).Bind(tokenMobileNo, tokenDeviceID, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive, &session.JWTToken)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Verify JWT token matches stored token
	if session.JWTToken != mainReq.JWTToken {
		return nil, fmt.Errorf("JWT token mismatch with stored token")
	}

	// Verify FCM token from JWT token matches the one in request
	if tokenFCMToken != mainReq.FCMToken {
		return nil, fmt.Errorf("FCM token mismatch - JWT token contains: %s, request contains: %s",
			tokenFCMToken[:20]+"...", mainReq.FCMToken[:20]+"...")
	}

	// Validate FCM token length
	if len(mainReq.FCMToken) < 100 {
		return nil, fmt.Errorf("FCM token too short or invalid")
	}

	// Prepare response data - only use sub_list
	responseData := s.getGameListData()

	return &models.MainScreenResponse{
		Status:      "success",
		Message:     "Main screen data retrieved successfully",
		MobileNo:    tokenMobileNo, // Use token mobile number
		DeviceID:    tokenDeviceID, // Use token device ID
		MessageType: mainReq.MessageType,
		Data:        responseData,
		UserInfo: map[string]interface{}{
			"user_id":   user.ID,
			"mobile_no": user.MobileNo,
			"full_name": user.FullName,
			"status":    user.Status,
			"language":  user.LanguageCode,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SocketID:  "",
		Event:     "main:screen:response",
	}, nil
}

// HandleContestList handles contest list requests with authentication validation
func (s *GameService) HandleContestList(contestReq models.ContestRequest) (*models.ContestResponse, error) {
	// Decrypt the simple JWT token to get the original values used during token creation
	simpleJWTData, err := utils.ValidateSimpleJWTToken(contestReq.JWTToken)
	if err != nil {
		return nil, fmt.Errorf("simple JWT token validation failed: %v", err)
	}

	// Extract values from the decrypted JWT token
	tokenMobileNo := simpleJWTData.MobileNo
	tokenDeviceID := simpleJWTData.DeviceID
	tokenFCMToken := simpleJWTData.FCMToken

	// Validate mobile number from token
	if len(tokenMobileNo) < 10 {
		return nil, fmt.Errorf("invalid mobile number in JWT token")
	}

	// Validate device ID from token
	if len(tokenDeviceID) < 1 {
		return nil, fmt.Errorf("invalid device ID in JWT token")
	}

	// Check if user exists and is active using token mobile number
	var user models.User
	err = s.cassandraSession.Query(`
		SELECT id, mobile_no, full_name, status, language_code
		FROM users
		WHERE mobile_no = ?
		ALLOW FILTERING
	`).Bind(tokenMobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active, jwt_token
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
		ALLOW FILTERING
	`).Bind(tokenMobileNo, tokenDeviceID, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive, &session.JWTToken)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Verify JWT token matches stored token
	if session.JWTToken != contestReq.JWTToken {
		return nil, fmt.Errorf("JWT token mismatch with stored token")
	}

	// Verify FCM token from JWT token matches the one in request
	if tokenFCMToken != contestReq.FCMToken {
		return nil, fmt.Errorf("FCM token mismatch - JWT token contains: %s, request contains: %s",
			tokenFCMToken[:20]+"...", contestReq.FCMToken[:20]+"...")
	}

	// Validate FCM token length
	if len(contestReq.FCMToken) < 100 {
		return nil, fmt.Errorf("FCM token too short or invalid")
	}

	// Get contest list data specifically
	responseData := s.getGameSubListData()

	return &models.ContestResponse{
		Status:      "success",
		Message:     "Contest list data retrieved successfully",
		MobileNo:    tokenMobileNo, // Use token mobile number
		DeviceID:    tokenDeviceID, // Use token device ID
		MessageType: contestReq.MessageType,
		Data:        responseData,
		UserInfo: map[string]interface{}{
			"user_id":   user.ID,
			"mobile_no": user.MobileNo,
			"full_name": user.FullName,
			"status":    user.Status,
			"language":  user.LanguageCode,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SocketID:  "",
		Event:     "contest:list:response",
	}, nil
}

// HandleContestJoin handles contest join requests
func (s *GameService) HandleContestJoin(joinReq models.ContestJoinRequest) (*models.ContestJoinResponse, error) {
	// Validate mobile number from token
	if len(joinReq.MobileNo) < 10 {
		return nil, fmt.Errorf("invalid mobile number in JWT token")
	}

	// Validate device ID from token
	if len(joinReq.DeviceID) < 1 {
		return nil, fmt.Errorf("invalid device ID in JWT token")
	}

	// Check if user exists and is active using token mobile number
	var user models.User
	err := s.cassandraSession.Query(`
		SELECT id, mobile_no, full_name, status, language_code
		FROM users
		WHERE mobile_no = ?
		ALLOW FILTERING
	`).Bind(joinReq.MobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active, jwt_token
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
		ALLOW FILTERING
	`).Bind(joinReq.MobileNo, joinReq.DeviceID, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive, &session.JWTToken)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Verify JWT token matches stored token
	if session.JWTToken != joinReq.JWTToken {
		return nil, fmt.Errorf("JWT token mismatch with stored token")
	}

	// Verify FCM token from JWT token matches the one in request
	if joinReq.FCMToken != session.FCMToken {
		return nil, fmt.Errorf("FCM token mismatch - JWT token contains: %s, request contains: %s",
			session.FCMToken[:20]+"...", joinReq.FCMToken[:20]+"...")
	}

	// Validate FCM token length
	if len(joinReq.FCMToken) < 100 {
		return nil, fmt.Errorf("FCM token too short or invalid")
	}

	// Validate contest ID
	if joinReq.ContestID == "" {
		return nil, fmt.Errorf("contest ID is required")
	}

	// Generate team ID if team name is provided
	teamID := ""
	if joinReq.TeamName != "" {
		teamID = fmt.Sprintf("team_%s_%s", joinReq.ContestID, time.Now().Format("20060102150405"))
	}

	// Insert into league_joins table
	leagueID := joinReq.ContestID
	status := "pending"
	userID := user.ID
	joinUUID := gocql.TimeUUID()
	joinedAtTime := time.Now().UTC()
	joinMonth := joinedAtTime.Format("2006-01")
	updatedAtTime := joinedAtTime
	inviteCode := "" // You can generate or leave empty
	role := "player"

	extraDataMap := map[string]interface{}{
		"team_id":   teamID,
		"team_name": joinReq.TeamName,
		"team_size": joinReq.TeamSize,
	}
	extraDataBytes, _ := json.Marshal(extraDataMap)
	extraData := string(extraDataBytes)

	// First, check if user has any existing pending entries and update their status_id
	iter := s.cassandraSession.Query(`
		SELECT status_id, join_month, joined_at FROM league_joins
		WHERE user_id = ? AND status_id = ? AND join_month = ?
		ALLOW FILTERING
	`, userID, "1", joinMonth).Iter()

	var existingStatusID, existingJoinMonth string
	var existingJoinedAt time.Time
	hasExisting := false
	for iter.Scan(&existingStatusID, &existingJoinMonth, &existingJoinedAt) {
		hasExisting = true
		// Read the old row from league_joins
		var oldUserID, oldLeagueID string
		var oldOpponentUserID, oldOpponentLeagueID string
		var oldID gocql.UUID
		err = s.cassandraSession.Query(`
			SELECT user_id, league_id, id, opponent_user_id, opponent_league_id
			FROM league_joins
			WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?
		`, userID, existingStatusID, existingJoinMonth, existingJoinedAt).Scan(
			&oldUserID, &oldLeagueID, &oldID, &oldOpponentUserID, &oldOpponentLeagueID)
		if err != nil {
			log.Printf("⚠️ Failed to read old league_joins row for status change: %v", err)
			continue
		}
		// Insert new row with status_id = '2' in league_joins
		err = s.cassandraSession.Query(`
			INSERT INTO league_joins (user_id, status_id, join_month, joined_at, league_id, id, opponent_user_id, opponent_league_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`).Bind(oldUserID, "2", existingJoinMonth, existingJoinedAt, oldLeagueID, oldID, oldOpponentUserID, oldOpponentLeagueID).Exec()
		if err != nil {
			log.Printf("⚠️ Failed to insert new league_joins row for status change: %v", err)
			continue
		}
		// Delete old row in league_joins
		err = s.cassandraSession.Query(`
			DELETE FROM league_joins WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?
		`, oldUserID, existingStatusID, existingJoinMonth, existingJoinedAt).Exec()
		if err != nil {
			log.Printf("⚠️ Failed to delete old league_joins row for status change: %v", err)
			continue
		}
		// Now update pending_league_joins for the same entry
		joinDay := existingJoinedAt.Format("2006-01-02")
		var oldPendingID gocql.UUID
		var oldPendingOpponentUserID string
		err = s.cassandraSession.Query(`
			SELECT id, opponent_user_id
			FROM pending_league_joins
			WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?
		`, "1", joinDay, oldLeagueID, existingJoinedAt).Scan(&oldPendingID, &oldPendingOpponentUserID)
		if err == nil {
			// Insert new row with status_id = '2' in pending_league_joins
			err = s.cassandraSession.Query(`
				INSERT INTO pending_league_joins (status_id, join_day, league_id, joined_at, id, opponent_user_id, user_id)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`).Bind("2", joinDay, oldLeagueID, existingJoinedAt, oldPendingID, oldPendingOpponentUserID, userID).Exec()
			if err != nil {
				log.Printf("⚠️ Failed to insert new pending_league_joins row for status change: %v", err)
			}
			// Delete old row in pending_league_joins
			err = s.cassandraSession.Query(`
				DELETE FROM pending_league_joins WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?
			`, "1", joinDay, oldLeagueID, existingJoinedAt).Exec()
			if err != nil {
				log.Printf("⚠️ Failed to delete old pending_league_joins row for status change: %v", err)
			}
		}
		log.Printf("🔄 Updated existing entry status_id to '2' for user %s in join_month %s and pending_league_joins", userID, existingJoinMonth)
	}
	iter.Close()

	// Determine status_id for new entry
	newStatusID := "1" // Default for first join
	if hasExisting {
		newStatusID = "1" // New entry always gets "1", previous gets "2"
		log.Printf("📝 User %s rejoining contest %s, new entry gets status_id '1'", userID, leagueID)
	} else {
		log.Printf("📝 User %s first time joining contest %s, status_id '1'", userID, leagueID)
	}

	// Now insert the new league_joins record
	err = s.cassandraSession.Query(`
		INSERT INTO league_joins (user_id, status_id, join_month, joined_at, league_id, status, extra_data, id, invite_code, opponent_league_id, opponent_user_id, role, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`).Bind(userID, newStatusID, joinMonth, joinedAtTime, leagueID, status, extraData, joinUUID, inviteCode, "NULL", "NULL", role, updatedAtTime).Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to insert league join: %v", err)
	}

	// Also insert into pending_league_joins for fast pending lookups (new schema)
	joinDay := joinedAtTime.Format("2006-01-02")
	err = s.cassandraSession.Query(`
		INSERT INTO pending_league_joins (status_id, join_day, league_id, joined_at, id, opponent_user_id, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`).Bind(newStatusID, joinDay, leagueID, joinedAtTime, joinUUID, "NULL", userID).Exec()
	if err != nil {
		log.Printf("⚠️ Failed to insert into pending_league_joins: %v", err)
	}

	return &models.ContestJoinResponse{
		Status:    "success",
		Message:   "Successfully joined contest",
		MobileNo:  joinReq.MobileNo,
		DeviceID:  joinReq.DeviceID,
		ContestID: joinReq.ContestID,
		TeamID:    teamID,
		JoinTime:  time.Now().UTC().Format(time.RFC3339),
		Data: map[string]interface{}{
			"contest_id":  joinUUID,
			"team_name":   joinReq.TeamName,
			"team_size":   joinReq.TeamSize,
			"join_status": "confirmed",
			"next_steps":  "Wait for contest start",
			"user_id":     user.ID,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SocketID:  "",
		Event:     "contest:join:response",
	}, nil
}

// HandleListContestScreen handles contest list screen requests with authentication validation
func (s *GameService) HandleListContestScreen(mainReq models.MainScreenRequest) (*models.MainScreenResponse, error) {
	// Convert MainScreenRequest to ContestRequest for consistency
	contestReq := models.ContestRequest{
		MobileNo:    mainReq.MobileNo,
		FCMToken:    mainReq.FCMToken,
		JWTToken:    mainReq.JWTToken,
		DeviceID:    mainReq.DeviceID,
		MessageType: mainReq.MessageType,
	}

	// Use the dedicated contest handler
	contestResponse, err := s.HandleContestList(contestReq)
	if err != nil {
		return nil, err
	}

	// Convert ContestResponse back to MainScreenResponse for backward compatibility
	return &models.MainScreenResponse{
		Status:      contestResponse.Status,
		Message:     contestResponse.Message,
		MobileNo:    contestResponse.MobileNo,
		DeviceID:    contestResponse.DeviceID,
		MessageType: contestResponse.MessageType,
		Data:        contestResponse.Data,
		UserInfo:    contestResponse.UserInfo,
		Timestamp:   contestResponse.Timestamp,
		SocketID:    contestResponse.SocketID,
		Event:       contestResponse.Event,
	}, nil
}

// HandleContestGap handles contest price gap requests with authentication validation
func (s *GameService) HandleContestGap(gapReq models.ContestGapRequest) (*models.ContestGapResponse, error) {
	// Decrypt the simple JWT token to get the original values used during token creation
	simpleJWTData, err := utils.ValidateSimpleJWTToken(gapReq.JWTToken)
	if err != nil {
		return nil, fmt.Errorf("simple JWT token validation failed: %v", err)
	}

	// Extract values from the decrypted JWT token
	tokenMobileNo := simpleJWTData.MobileNo
	tokenDeviceID := simpleJWTData.DeviceID
	tokenFCMToken := simpleJWTData.FCMToken

	// Validate mobile number from token
	if len(tokenMobileNo) < 10 {
		return nil, fmt.Errorf("invalid mobile number in JWT token")
	}

	// Validate device ID from token
	if len(tokenDeviceID) < 1 {
		return nil, fmt.Errorf("invalid device ID in JWT token")
	}

	// Check if user exists and is active using token mobile number
	var user models.User
	err = s.cassandraSession.Query(`
		SELECT id, mobile_no, full_name, status, language_code
		FROM users
		WHERE mobile_no = ?
		ALLOW FILTERING
	`).Bind(tokenMobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("user not found or not authenticated")
	}

	// Check if session exists and is active using token values
	var session models.Session
	err = s.cassandraSession.Query(`
		SELECT session_token, device_id, fcm_token, expires_at, is_active, jwt_token
		FROM sessions
		WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
		ALLOW FILTERING
	`).Bind(tokenMobileNo, tokenDeviceID, time.Now()).Scan(&session.SessionToken, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive, &session.JWTToken)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Verify JWT token matches stored token
	if session.JWTToken != gapReq.JWTToken {
		return nil, fmt.Errorf("JWT token mismatch with stored token")
	}

	// Verify FCM token from JWT token matches the one in request
	if tokenFCMToken != gapReq.FCMToken {
		return nil, fmt.Errorf("FCM token mismatch - JWT token contains: %s, request contains: %s",
			tokenFCMToken[:20]+"...", gapReq.FCMToken[:20]+"...")
	}

	// Validate FCM token length
	if len(gapReq.FCMToken) < 100 {
		return nil, fmt.Errorf("FCM token too short or invalid")
	}

	// Get contest list data and calculate price gaps
	allContests := s.getGameSubListData()

	// Calculate price gap data
	gapData := s.calculatePriceGapData(allContests, gapReq)

	return &models.ContestGapResponse{
		Status:      "success",
		Message:     "Contest price gap data retrieved successfully",
		MobileNo:    tokenMobileNo, // Use token mobile number
		DeviceID:    tokenDeviceID, // Use token device ID
		MessageType: gapReq.MessageType,
		Data:        gapData,
		UserInfo: map[string]interface{}{
			"user_id":   user.ID,
			"mobile_no": user.MobileNo,
			"full_name": user.FullName,
			"status":    user.Status,
			"language":  user.LanguageCode,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SocketID:  "",
		Event:     "list:contest:gap:response",
	}, nil
}

// calculatePriceGapData calculates price gap information from contest data
func (s *GameService) calculatePriceGapData(allContests map[string]interface{}, gapReq models.ContestGapRequest) map[string]interface{} {
	gapData := map[string]interface{}{
		"price_gaps": []map[string]interface{}{},
		"summary":    map[string]interface{}{},
	}

	if gamelist, exists := allContests["gamelist"]; exists {
		var contests []map[string]interface{}

		// Handle different possible types
		switch v := gamelist.(type) {
		case []map[string]interface{}:
			contests = v
		case []interface{}:
			// Convert []interface{} to []map[string]interface{}
			contests = make([]map[string]interface{}, 0, len(v))
			for _, item := range v {
				if contest, ok := item.(map[string]interface{}); ok {
					contests = append(contests, contest)
				}
			}
		default:
			return gapData
		}

		if len(contests) > 0 {
			var winPrices []float64
			var entryFees []float64

			// Extract all prices and fees
			for _, contest := range contests {
				if winPrice, exists := contest["contest_win_price"]; exists {
					price := s.convertToFloat(winPrice)
					winPrices = append(winPrices, price)
				}

				if entryFee, exists := contest["contest_entryfee"]; exists {
					fee := s.convertToFloat(entryFee)
					entryFees = append(entryFees, fee)
				}
			}

			// Calculate gaps based on message type
			switch gapReq.MessageType {
			case "win_price_gap":
				gapData["price_gaps"] = s.calculateWinPriceGaps(winPrices)
			case "entry_fee_gap":
				gapData["price_gaps"] = s.calculateEntryFeeGaps(entryFees)
			case "price_gap":
				gapData["price_gaps"] = s.calculateCombinedPriceGaps(winPrices, entryFees)
			default:
				gapData["price_gaps"] = s.calculateAllPriceGaps(winPrices, entryFees)
			}

			// Add summary statistics
			gapData["summary"] = map[string]interface{}{
				"total_contests":  len(contests),
				"filter_type":     gapReq.MessageType,
				"win_price_range": s.calculateRange(winPrices),
				"entry_fee_range": s.calculateRange(entryFees),
				"avg_win_price":   s.calculateAverage(winPrices),
				"avg_entry_fee":   s.calculateAverage(entryFees),
			}
		}
	}

	return gapData
}

// calculateWinPriceGaps calculates win price gap ranges
func (s *GameService) calculateWinPriceGaps(prices []float64) []map[string]interface{} {
	if len(prices) == 0 {
		return []map[string]interface{}{}
	}

	// Sort prices
	sort.Float64s(prices)

	var gaps []map[string]interface{}

	// Create price ranges
	ranges := []struct {
		min  float64
		max  float64
		name string
	}{
		{0, 1000, "Low"},
		{1000, 3000, "Medium"},
		{3000, 5000, "High"},
		{5000, 10000, "Premium"},
		{10000, 999999, "Elite"},
	}

	for _, r := range ranges {
		count := 0
		for _, price := range prices {
			if price >= r.min && price < r.max {
				count++
			}
		}

		if count > 0 {
			gaps = append(gaps, map[string]interface{}{
				"type":          "win_price",
				"range_name":    r.name,
				"min_price":     r.min,
				"max_price":     r.max,
				"contest_count": count,
				"percentage":    float64(count) / float64(len(prices)) * 100,
			})
		}
	}

	return gaps
}

// calculateEntryFeeGaps calculates entry fee gap ranges
func (s *GameService) calculateEntryFeeGaps(fees []float64) []map[string]interface{} {
	if len(fees) == 0 {
		return []map[string]interface{}{}
	}

	// Sort fees
	sort.Float64s(fees)

	var gaps []map[string]interface{}

	// Create fee ranges
	ranges := []struct {
		min  float64
		max  float64
		name string
	}{
		{0, 10, "Free"},
		{10, 50, "Low"},
		{50, 100, "Medium"},
		{100, 200, "High"},
		{200, 999999, "Premium"},
	}

	for _, r := range ranges {
		count := 0
		for _, fee := range fees {
			if fee >= r.min && fee < r.max {
				count++
			}
		}

		if count > 0 {
			gaps = append(gaps, map[string]interface{}{
				"type":          "entry_fee",
				"range_name":    r.name,
				"min_price":     r.min,
				"max_price":     r.max,
				"contest_count": count,
				"percentage":    float64(count) / float64(len(fees)) * 100,
			})
		}
	}

	return gaps
}

// calculateCombinedPriceGaps calculates both win price and entry fee gaps
func (s *GameService) calculateCombinedPriceGaps(prices, fees []float64) []map[string]interface{} {
	winGaps := s.calculateWinPriceGaps(prices)
	entryGaps := s.calculateEntryFeeGaps(fees)

	var combined []map[string]interface{}
	combined = append(combined, winGaps...)
	combined = append(combined, entryGaps...)

	return combined
}

// calculateAllPriceGaps calculates all possible price gaps
func (s *GameService) calculateAllPriceGaps(prices, fees []float64) []map[string]interface{} {
	return s.calculateCombinedPriceGaps(prices, fees)
}

// calculateRange calculates min, max, and range for a slice of values
func (s *GameService) calculateRange(values []float64) map[string]interface{} {
	if len(values) == 0 {
		return map[string]interface{}{
			"min":   0,
			"max":   0,
			"range": 0,
		}
	}

	sort.Float64s(values)
	min := values[0]
	max := values[len(values)-1]

	return map[string]interface{}{
		"min":   min,
		"max":   max,
		"range": max - min,
	}
}

// calculateAverage calculates the average of a slice of values
func (s *GameService) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

// convertToFloat converts interface{} to float64 for price comparisons
func (s *GameService) convertToFloat(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

// GetCassandraSession returns the Cassandra session for external access
func (s *GameService) GetCassandraSession() *gocql.Session {
	return s.cassandraSession
}

// MatchAndUpdateOpponent finds another unmatched user in the same league, updates both users' records, and returns the opponent info
func (s *GameService) MatchAndUpdateOpponent(currentUserID, leagueID string, currentJoinedAt time.Time) (*models.LeagueJoin, error) {
	joinDay := currentJoinedAt.Format("2006-01-02")
	var currentUserOpponentUserID *string
	err := s.cassandraSession.Query(`
		SELECT opponent_user_id
		FROM pending_league_joins
		WHERE status_id = ? AND join_day = ? AND league_id = ? AND user_id = ?
		LIMIT 1
	`, "1", joinDay, leagueID, currentUserID).Scan(&currentUserOpponentUserID)

	if err == nil && currentUserOpponentUserID != nil && *currentUserOpponentUserID != "" {
		return nil, nil
	}

	// Step 1: Find a candidate opponent in pending_league_joins
	var opponentUserID string
	var opponentJoinedAt time.Time
	var opponentID gocql.UUID
	iter := s.cassandraSession.Query(`
		SELECT user_id, joined_at, id
		FROM pending_league_joins
		WHERE status_id = ? AND join_day = ? AND league_id = ? 
		ORDER BY joined_at ASC
		LIMIT 1
	`, "1", joinDay, leagueID).Iter()

	if !iter.Scan(&opponentUserID, &opponentJoinedAt, &opponentID) {
		iter.Close()
		return nil, nil // No match found
	}
	iter.Close()

	if opponentUserID == currentUserID {
		return nil, nil
	}

	// Step 2: Fetch the full row for that user from league_joins
	opponentJoinMonth := opponentJoinedAt.Format("2006-01")
	var fullOpponent models.LeagueJoin
	err = s.cassandraSession.Query(`
		SELECT user_id, league_id, joined_at, id, opponent_user_id, opponent_league_id
		FROM league_joins
		WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?
		LIMIT 1
	`, opponentUserID, "1", opponentJoinMonth, opponentJoinedAt).Scan(
		&fullOpponent.UserID, &fullOpponent.LeagueID, &fullOpponent.JoinedAt, &fullOpponent.ID, &fullOpponent.OpponentUserID, &fullOpponent.OpponentLeagueID,
	)
	if err != nil {
		return nil, err
	}

	// Step 3: Update both tables for both users
	currentJoinMonth := currentJoinedAt.Format("2006-01")
	currentJoinDay := currentJoinedAt.Format("2006-01-02")
	currentJoinedAtTime := currentJoinedAt

	err = s.cassandraSession.Query(`
		UPDATE league_joins SET opponent_user_id = ?, opponent_league_id = ?
		WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?
	`, opponentUserID, leagueID, currentUserID, "1", currentJoinMonth, currentJoinedAtTime).Exec()

	err = s.cassandraSession.Query(`
		UPDATE league_joins SET opponent_user_id = ?, opponent_league_id = ?
		WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?
	`, currentUserID, leagueID, opponentUserID, "1", opponentJoinMonth, opponentJoinedAt).Exec()

	err = s.cassandraSession.Query(`
		UPDATE pending_league_joins SET opponent_user_id = ?
		WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?
	`, opponentUserID, "1", currentJoinDay, leagueID, currentJoinedAtTime).Exec()

	err = s.cassandraSession.Query(`
		UPDATE pending_league_joins SET opponent_user_id = ?
		WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?
	`, currentUserID, "1", joinDay, leagueID, opponentJoinedAt).Exec()

	// Step 4: Create match pair entry in match_pairs table
	err = s.createMatchPairEntry(currentUserID, opponentUserID, leagueID, currentJoinedAt)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to create match pair entry: %v", err)
		// Don't return error here as the main matching logic succeeded
	}

	return &fullOpponent, nil
}

// createMatchPairEntry creates a match pair entry in the match_pairs table
func (s *GameService) createMatchPairEntry(user1ID, user2ID, leagueID string, matchTime time.Time) error {
	// Generate UUID for the match pair
	id := gocql.TimeUUID()
	now := time.Now()

	// Don't store any data in user1_data and user2_data fields
	user1Data := ""
	user2Data := ""

	// Insert the match pair
	query := `INSERT INTO match_pairs (id, user1_id, user2_id, user1_data, user2_data, status, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	err := s.cassandraSession.Query(query, id, user1ID, user2ID, user1Data, user2Data, "active", now, now).Exec()
	if err != nil {
		log.Printf("❌ Error creating match pair entry: %v", err)
		return fmt.Errorf("failed to create match pair entry: %v", err)
	}

	log.Printf("✅ Created match pair entry with ID: %s for users %s and %s in league %s", id.String(), user1ID, user2ID, leagueID)
	return nil
}

// GetLeagueJoinEntry fetches a user's league_joins entry for a contest
func (s *GameService) GetLeagueJoinEntry(userID, contestID string) (*models.LeagueJoin, error) {
	log.Printf("🔍 GetLeagueJoinEntry - userID: %s, contestID: %s", userID, contestID)

	// Debug: Check what's in the database for this user
	log.Printf("🔍 DEBUG: Checking league_joins table for user %s in contest %s", userID, contestID)

	iter := s.cassandraSession.Query(`
		SELECT user_id, status_id, join_month, joined_at, league_id, id, opponent_user_id, opponent_league_id
		FROM league_joins
		WHERE user_id = ? AND league_id = ? AND status_id = ?
		ALLOW FILTERING
	`, userID, contestID, "1").Iter()

	var (
		entry       models.LeagueJoin
		statusID    string
		joinMonth   string
		joinedAt    time.Time
		found       bool
		latestTime  time.Time
		latestEntry models.LeagueJoin
	)
	for iter.Scan(&entry.UserID, &statusID, &joinMonth, &joinedAt, &entry.LeagueID, &entry.ID, &entry.OpponentUserID, &entry.OpponentLeagueID) {
		log.Printf("🔍 DEBUG: Found entry - User: %s, Status: %s, Month: %s, Joined: %s, League: %s, Opponent: %s, OpponentLeague: %s",
			entry.UserID, statusID, joinMonth, joinedAt.Format(time.RFC3339), entry.LeagueID, entry.OpponentUserID, entry.OpponentLeagueID)

		if !found || joinedAt.After(latestTime) {
			latestEntry = entry
			latestEntry.JoinedAt = joinedAt
			latestTime = joinedAt
			found = true
			log.Printf("✅ Found league join entry: userID=%s, contestID=%s, joinedAt=%s", entry.UserID, entry.LeagueID, joinedAt.Format(time.RFC3339))
		}
	}
	iter.Close()

	if !found {
		log.Printf("❌ No league join entry found for userID=%s, contestID=%s", userID, contestID)
		return nil, fmt.Errorf("entry not found")
	}
	return &latestEntry, nil
}

// UpdateOpponentDetails updates opponent details for a user's league join entry
func (s *GameService) UpdateOpponentDetails(userID, leagueID, opponentUserID, opponentLeagueID, joinedAt string) error {
	joinMonth := ""
	if t, err := time.Parse(time.RFC3339, joinedAt); err == nil {
		joinMonth = t.Format("2006-01")
	}

	// Update league_joins for the user
	err := s.cassandraSession.Query(`
		UPDATE league_joins SET opponent_user_id = ?, opponent_league_id = ?
		WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?
	`, opponentUserID, opponentLeagueID, userID, "1", joinMonth, joinedAt).Exec()
	if err != nil {
		log.Printf("⚠️ Failed to update league_joins opponent details: %v", err)
		return err
	}

	// Update pending_league_joins for the user
	joinDay := ""
	if t, err := time.Parse(time.RFC3339, joinedAt); err == nil {
		joinDay = t.Format("2006-01-02")
	}
	err = s.cassandraSession.Query(`
		UPDATE pending_league_joins SET opponent_user_id = ?
		WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?
	`, opponentUserID, "1", joinDay, leagueID, joinedAt).Exec()
	if err != nil {
		log.Printf("⚠️ Failed to update pending_league_joins opponent details: %v", err)
		return err
	}

	log.Printf("✅ Updated opponent details for user %s in league %s: opponent=%s", userID, leagueID, opponentUserID)
	return nil
}

// UpdateLeagueJoinStatus updates the status_id for a user's league_joins entry
func (s *GameService) UpdateLeagueJoinStatus(userID, leagueID, newStatusID, joinedAt string) error {
	joinMonth := ""
	if t, err := time.Parse(time.RFC3339, joinedAt); err == nil {
		joinMonth = t.Format("2006-01")
	}

	// 1. Read the old row
	var oldUserID, oldLeagueID string
	var oldJoinedAt time.Time
	var oldOpponentUserID, oldOpponentLeagueID string
	var oldID gocql.UUID
	err := s.cassandraSession.Query(`
		SELECT user_id, league_id, joined_at, id, opponent_user_id, opponent_league_id
		FROM league_joins
		WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?
	`, userID, "1", joinMonth, joinedAt).Scan(
		&oldUserID, &oldLeagueID, &oldJoinedAt, &oldID, &oldOpponentUserID, &oldOpponentLeagueID)
	if err != nil {
		log.Printf("⚠️ Failed to read old league_joins row for status change: %v", err)
		return err
	}

	// 2. Insert new row with new status_id
	err = s.cassandraSession.Query(`
		INSERT INTO league_joins (user_id, status_id, join_month, joined_at, league_id, id, opponent_user_id, opponent_league_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`).Bind(oldUserID, newStatusID, joinMonth, oldJoinedAt, oldLeagueID, oldID, oldOpponentUserID, oldOpponentLeagueID).Exec()
	if err != nil {
		log.Printf("⚠️ Failed to insert new league_joins row for status change: %v", err)
		return err
	}

	// 3. Delete old row
	err = s.cassandraSession.Query(`
		DELETE FROM league_joins WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?
	`, oldUserID, "1", joinMonth, oldJoinedAt).Exec()
	if err != nil {
		log.Printf("⚠️ Failed to delete old league_joins row for status change: %v", err)
		return err
	}

	return nil
}

// UpdateLeagueJoinStatusBoth updates status_id in both league_joins and pending_league_joins for all pending entries for the user and league in the last 3 months
func (s *GameService) UpdateLeagueJoinStatusBoth(userID, leagueID, newStatusID, _ string) error {
	now := time.Now()
	months := []string{
		now.Format("2006-01"),
		now.AddDate(0, -1, 0).Format("2006-01"),
	}
	for _, joinMonth := range months {
		iter := s.cassandraSession.Query(`
			SELECT joined_at, league_id
			FROM league_joins
			WHERE user_id = ? AND status_id = ? AND join_month = ?
		`, userID, "1", joinMonth).Iter()

		var joinedAt time.Time
		var leagueIDFound string
		for iter.Scan(&joinedAt, &leagueIDFound) {
			if leagueIDFound != leagueID {
				continue
			}
			// Read the old row
			var status, extraData, inviteCode, opponentLeagueID, opponentUserID, role string
			var id gocql.UUID
			var updatedAt time.Time
			err := s.cassandraSession.Query(`
				SELECT status, extra_data, id, invite_code, opponent_league_id, opponent_user_id, role, updated_at
				FROM league_joins
				WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?
			`, userID, "1", joinMonth, joinedAt).Scan(&status, &extraData, &id, &inviteCode, &opponentLeagueID, &opponentUserID, &role, &updatedAt)
			if err != nil {
				continue
			}
			// Insert new row with newStatusID in league_joins (preserving opponent details)
			err = s.cassandraSession.Query(`
				INSERT INTO league_joins (user_id, status_id, join_month, joined_at, league_id, status, extra_data, id, invite_code, opponent_league_id, opponent_user_id, role, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`).Bind(userID, newStatusID, joinMonth, joinedAt, leagueIDFound, status, extraData, id, inviteCode, opponentLeagueID, opponentUserID, role, updatedAt).Exec()
			if err != nil {
				continue
			}
			// Delete old row in league_joins
			err = s.cassandraSession.Query(`
				DELETE FROM league_joins WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?
			`, userID, "1", joinMonth, joinedAt).Exec()
			if err != nil {
				continue
			}
			// Update all matching pending_league_joins entries for this user, league, and joinedAt (could be multiple)
			joinDay := joinedAt.Format("2006-01-02")
			pendingIter := s.cassandraSession.Query(`
				SELECT id, opponent_user_id, user_id, joined_at
				FROM pending_league_joins
				WHERE status_id = ? AND join_day = ? AND league_id = ?
			`, "1", joinDay, leagueIDFound).Iter()
			var pendingID gocql.UUID
			var pendingOpponentUserID, pendingUserID string
			var pendingJoinedAt time.Time
			for pendingIter.Scan(&pendingID, &pendingOpponentUserID, &pendingUserID, &pendingJoinedAt) {
				if pendingUserID != userID {
					continue
				}
				// Insert new row with newStatusID (preserving opponent details)
				err = s.cassandraSession.Query(`
					INSERT INTO pending_league_joins (status_id, join_day, league_id, joined_at, id, opponent_user_id, user_id)
					VALUES (?, ?, ?, ?, ?, ?, ?)
				`).Bind(newStatusID, joinDay, leagueIDFound, pendingJoinedAt, pendingID, pendingOpponentUserID, userID).Exec()
				if err != nil {
					continue
				}
				// Delete old row
				err = s.cassandraSession.Query(`
					DELETE FROM pending_league_joins WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?
				`, "1", joinDay, leagueIDFound, pendingJoinedAt).Exec()
				if err != nil {
					continue
				}
			}
			pendingIter.Close()
		}
		iter.Close()
	}
	return nil
}
