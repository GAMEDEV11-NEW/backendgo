package services

import (
	"fmt"
	"gofiber/app/models"
	"time"

	socketio "github.com/doquangtan/socket.io/v4"
	"github.com/gocql/gocql"
)

// SocketService handles all socket-related business logic
// Now acts as a coordinator between AuthService and GameService
type SocketService struct {
	cassandraSession *gocql.Session
	authService      *AuthService
	gameService      *GameService
	messagingService *MessagingService
	io               *socketio.Io // Add Socket.IO instance
}

// NewSocketService creates a new socket service instance using Cassandra
func NewSocketService(cassandraSession *gocql.Session) *SocketService {
	if cassandraSession == nil {
		panic("Cassandra session cannot be nil")
	}

	service := &SocketService{
		cassandraSession: cassandraSession,
		authService:      NewAuthService(cassandraSession),
		gameService:      NewGameService(cassandraSession),
	}
	return service
}

// SetMessagingService sets the messaging service for the socket service
func (s *SocketService) SetMessagingService(messagingService *MessagingService) {
	s.messagingService = messagingService
}

// SetIo sets the Socket.IO instance for the socket service
func (s *SocketService) SetIo(io *socketio.Io) {
	s.io = io
}

// GetIo returns the Socket.IO instance
func (s *SocketService) GetIo() *socketio.Io {
	return s.io
}

// Auth-related methods - delegate to AuthService
func (s *SocketService) GenerateSessionToken() (string, error) {
	return s.authService.GenerateSessionToken()
}

func (s *SocketService) GenerateOTP() int {
	return s.authService.GenerateOTP()
}

func (s *SocketService) HandleDeviceInfo(deviceInfo models.DeviceInfo, socketID string) models.DeviceInfoResponse {
	return s.authService.HandleDeviceInfo(deviceInfo, socketID)
}

func (s *SocketService) HandleLogin(loginReq models.LoginRequest) (*models.LoginResponse, error) {
	return s.authService.HandleLogin(loginReq)
}

func (s *SocketService) HandleOTPVerification(otpReq models.OTPVerificationRequest) (*models.OTPVerificationResponse, error) {
	return s.authService.HandleOTPVerification(otpReq)
}

func (s *SocketService) HandleSetProfile(profileReq models.SetProfileRequest) (*models.SetProfileResponse, error) {
	return s.authService.HandleSetProfile(profileReq)
}

func (s *SocketService) HandleSetLanguage(langReq models.SetLanguageRequest) (*models.SetLanguageResponse, error) {
	return s.authService.HandleSetLanguage(langReq)
}

func (s *SocketService) ValidateSession(sessionToken, mobileNo string) bool {
	return s.authService.ValidateSession(sessionToken, mobileNo)
}

func (s *SocketService) CleanupExpiredSessions() error {
	return s.authService.CleanupExpiredSessions()
}

func (s *SocketService) CleanupExpiredOTPs() error {
	return s.authService.CleanupExpiredOTPs()
}

func (s *SocketService) GetLatestOTP(phoneOrEmail, purpose string) (*models.OTPData, error) {
	return s.authService.GetLatestOTP(phoneOrEmail, purpose)
}

func (s *SocketService) ResendOTP(mobileNo string) (int, error) {
	return s.authService.ResendOTP(mobileNo)
}

// GetSessionService returns the session service
func (s *SocketService) GetSessionService() *SessionService {
	return s.authService.sessionService
}

// Game-related methods - delegate to GameService
func (s *SocketService) HandlePlayerAction(actionReq models.PlayerActionRequest) (*models.PlayerActionResponse, error) {
	return s.gameService.HandlePlayerAction(actionReq)
}

func (s *SocketService) HandleHeartbeat() models.HeartbeatResponse {
	return s.gameService.HandleHeartbeat()
}

func (s *SocketService) HandleWelcome() models.WelcomeResponse {
	return s.gameService.HandleWelcome()
}

func (s *SocketService) HandleHealthCheck() models.HealthCheckResponse {
	return s.gameService.HandleHealthCheck()
}

func (s *SocketService) HandleStaticMessage(staticReq models.StaticMessageRequest) (*models.StaticMessageResponse, error) {
	return s.gameService.HandleStaticMessage(staticReq)
}

func (s *SocketService) GetGameListFromRedis() (map[string]interface{}, error) {
	return s.gameService.GetGameListFromRedis()
}

func (s *SocketService) GetGameListDataPublic() map[string]interface{} {
	return s.gameService.GetGameListDataPublic()
}

func (s *SocketService) HandleMainScreen(mainReq models.MainScreenRequest) (*models.MainScreenResponse, error) {
	return s.gameService.HandleMainScreen(mainReq)
}

func (s *SocketService) HandleContestList(contestReq models.ContestRequest) (*models.ContestResponse, error) {
	return s.gameService.HandleContestList(contestReq)
}

func (s *SocketService) HandleContestJoin(joinReq models.ContestJoinRequest) (*models.ContestJoinResponse, error) {
	return s.gameService.HandleContestJoin(joinReq)
}

func (s *SocketService) HandleListContestScreen(mainReq models.MainScreenRequest) (*models.MainScreenResponse, error) {
	return s.gameService.HandleListContestScreen(mainReq)
}

func (s *SocketService) HandleContestGap(gapReq models.ContestGapRequest) (*models.ContestGapResponse, error) {
	return s.gameService.HandleContestGap(gapReq)
}

// Expose MatchAndUpdateOpponent for matchmaking
func (s *SocketService) MatchAndUpdateOpponent(currentUserID, leagueID string, currentJoinedAt time.Time) (*models.LeagueJoin, gocql.UUID, error) {
	return s.gameService.MatchAndUpdateOpponent(currentUserID, leagueID, currentJoinedAt)
}

// Expose GetLeagueJoinEntry for checking opponent info
func (s *SocketService) GetLeagueJoinEntry(userID, contestID string) (*models.LeagueJoin, error) {
	return s.gameService.GetLeagueJoinEntry(userID, contestID)
}

// Expose UpdateLeagueJoinStatus for updating status_id on disconnect or rejoin
func (s *SocketService) UpdateLeagueJoinStatus(userID, leagueID, newStatusID, joinedAt string) error {
	return s.gameService.UpdateLeagueJoinStatus(userID, leagueID, newStatusID, joinedAt)
}

// Expose UpdateOpponentDetails for updating opponent information
func (s *SocketService) UpdateOpponentDetails(userID, leagueID, opponentUserID, opponentLeagueID, joinedAt string) error {
	return s.gameService.UpdateOpponentDetails(userID, leagueID, opponentUserID, opponentLeagueID, joinedAt)
}

// UpdateLeagueJoinStatusBoth updates status_id in both league_joins and pending_league_joins
func (s *SocketService) UpdateLeagueJoinStatusBoth(userID, leagueID, newStatusID, joinedAt string) error {
	return s.gameService.UpdateLeagueJoinStatusBoth(userID, leagueID, newStatusID, joinedAt)
}

// CreateGamePieces creates pieces for matched users
func (s *SocketService) CreateGamePieces(gameID, user1ID, user2ID string) error {
	gamePiecesService := NewGamePiecesService(s.cassandraSession)
	return gamePiecesService.CreatePiecesForMatch(gameID, user1ID, user2ID)
}

// GetMatchPairID retrieves the match_pairs ID for a given user pair and league
func (s *SocketService) GetMatchPairID(user1ID, user2ID, leagueID string) (gocql.UUID, error) {
	// If we don't have both user IDs, we can't find the match pair
	if user1ID == "" || user2ID == "" {
		return gocql.UUID{}, fmt.Errorf("both user IDs are required")
	}

	// First, let's check what match pairs exist in the database for debugging
	debugIter := s.cassandraSession.Query(`
		SELECT id, user1_id, user2_id, status FROM match_pairs 
		LIMIT 20
	`).Iter()

	var debugID gocql.UUID
	var debugUser1ID, debugUser2ID, debugStatus string
	debugCount := 0
	for debugIter.Scan(&debugID, &debugUser1ID, &debugUser2ID, &debugStatus) {
		debugCount++
	}
	debugIter.Close()

	// Get ALL match pairs for this user pair and find the most recent one
	iter := s.cassandraSession.Query(`
		SELECT id, user1_id, user2_id, status FROM match_pairs 
		WHERE user1_id = ? AND user2_id = ?
		ALLOW FILTERING
	`, user1ID, user2ID).Iter()

	var foundMatchPair bool
	var latestMatchPairID gocql.UUID
	var latestTime time.Time

	var id gocql.UUID
	var u1ID, u2ID, status string
	for iter.Scan(&id, &u1ID, &u2ID, &status) {
		// Only consider active match pairs (exclude disconnected, cancelled, completed)
		if status == "active" {
			// Parse the UUID to get the timestamp (UUIDs are time-based)
			if id.Time().After(latestTime) {
				latestMatchPairID = id
				latestTime = id.Time()
				foundMatchPair = true
			}
		}
	}
	iter.Close()

	// If not found with first order, try reverse order
	if !foundMatchPair {
		iter = s.cassandraSession.Query(`
			SELECT id, user1_id, user2_id, status FROM match_pairs 
			WHERE user1_id = ? AND user2_id = ?
			ALLOW FILTERING
		`, user2ID, user1ID).Iter()

		for iter.Scan(&id, &u1ID, &u2ID, &status) {
			// Only consider active match pairs (exclude disconnected, cancelled, completed)
			if status == "active" {
				// Parse the UUID to get the timestamp (UUIDs are time-based)
				if id.Time().After(latestTime) {
					latestMatchPairID = id
					latestTime = id.Time()
					foundMatchPair = true
				}
			}
		}
		iter.Close()
	}

	if foundMatchPair {
		return latestMatchPairID, nil
	}

	// Let's also check if there are any match pairs that contain either of our users
	userMatchIter := s.cassandraSession.Query(`
		SELECT id, user1_id, user2_id, status FROM match_pairs 
		WHERE user1_id = ? OR user2_id = ? OR user1_id = ? OR user2_id = ?
		ALLOW FILTERING
	`, user1ID, user1ID, user2ID, user2ID).Iter()

	var userMatchID gocql.UUID
	var userMatchUser1ID, userMatchUser2ID, userMatchStatus string
	userMatchCount := 0
	for userMatchIter.Scan(&userMatchID, &userMatchUser1ID, &userMatchUser2ID, &userMatchStatus) {
		userMatchCount++
	}
	userMatchIter.Close()

	return gocql.UUID{}, fmt.Errorf("failed to get match pair ID: no active match pairs found")
}

// GetCassandraSession returns the Cassandra session for external access
func (s *SocketService) GetCassandraSession() *gocql.Session {
	return s.cassandraSession
}

// GetAuthService returns the auth service instance
func (s *SocketService) GetAuthService() *AuthService {
	return s.authService
}

// GetGameService returns the game service instance
func (s *SocketService) GetGameService() *GameService {
	return s.gameService
}

// GetMessagingService returns the messaging service instance
func (s *SocketService) GetMessagingService() *MessagingService {
	return s.messagingService
}

// StoreConnectionData stores connection data when a user connects
func (s *SocketService) StoreConnectionData(socketID, userID, mobileNo, sessionToken, deviceID, fcmToken, userAgent, ipAddress, namespace string) error {
	if s.messagingService == nil {
		return fmt.Errorf("messaging service not initialized")
	}

	// Connection data is now stored as part of session data
	// This method is kept for compatibility but the actual storage is handled by session service
	return nil
}

// RemoveConnectionData removes connection data when a user disconnects
func (s *SocketService) RemoveConnectionData(socketID string) error {
	if s.messagingService == nil {
		return fmt.Errorf("messaging service not initialized")
	}

	return s.messagingService.RemoveConnectionData(socketID)
}

// UpdateConnectionData updates connection data (e.g., last seen)
func (s *SocketService) UpdateConnectionData(socketID string) error {
	if s.messagingService == nil {
		return fmt.Errorf("messaging service not initialized")
	}

	return s.messagingService.UpdateConnectionData(socketID)
}

// UpdateMatchPairStatus updates the status of a match pair
func (s *SocketService) UpdateMatchPairStatus(userID, newStatus string) error {
	// Find all match pairs for this user and update their status
	iter := s.cassandraSession.Query(`
		SELECT id, user1_id, user2_id, status FROM match_pairs 
		WHERE user1_id = ? OR user2_id = ?
		ALLOW FILTERING
	`, userID, userID).Iter()

	var matchID gocql.UUID
	var u1ID, u2ID, currentStatus string
	updatedCount := 0

	for iter.Scan(&matchID, &u1ID, &u2ID, &currentStatus) {
		// Only update if current status is "active"
		if currentStatus == "active" {
			err := s.cassandraSession.Query(`
				UPDATE match_pairs SET status = ?, updated_at = ?
				WHERE id = ?
			`, newStatus, time.Now(), matchID).Exec()

			if err != nil {
				// Failed to update match pair
			} else {
				updatedCount++
			}
		}
	}
	iter.Close()

	return nil
}

// HandleSocketDisconnect handles user disconnection by:
// 1. Removing socket mapping but keeping session active
// 2. Updating league_joins table to set status_id = 3 and opponent_user_id = NULL for the user
// 3. Updating match_pairs table to set status = "disconnected" for the user
func (s *SocketService) HandleSocketDisconnect(socketID string) error {

	// Step 1: Get user info from sessions_by_socket table
	var mobileNo, userID, sessionToken string
	var createdAt time.Time

	err := s.cassandraSession.Query(`
		SELECT mobile_no, user_id, session_token, created_at
		FROM sessions_by_socket
		WHERE socket_id = ?
	`, socketID).Scan(&mobileNo, &userID, &sessionToken, &createdAt)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil // Not an error - just no session to clean up
		}
		return err
	}

	// Step 2: Delete socket mapping but keep session active
	err = s.cassandraSession.Query(`
		DELETE FROM sessions_by_socket
		WHERE socket_id = ?
	`, socketID).Exec()

	if err != nil {
		return err
	}

	// Step 2.5: Keep Redis session active - just remove socket mapping
	// Session remains valid until user explicitly logs out

	// Step 3: Update match_pairs table status for this user
	err = s.UpdateMatchPairStatus(userID, "disconnected")

	// Step 4: Find and update ALL league_joins entries for this user with status_id = 1 for the current and previous two months
	months := []string{}
	now := time.Now()
	for i := 0; i < 3; i++ {
		months = append(months, now.AddDate(0, -i, 0).Format("2006-01"))
	}
	for _, joinMonth := range months {
		iter := s.cassandraSession.Query(`
			SELECT joined_at, league_id
			FROM league_joins
			WHERE user_id = ? AND status_id = ? AND join_month = ?
		`, userID, "1", joinMonth).Iter()

		var joinedAt time.Time
		var leagueID string
		for iter.Scan(&joinedAt, &leagueID) {
			// Read the old row
			var status, extraData, inviteCode, opponentLeagueID, opponentUserID, role string
			var id gocql.UUID
			var updatedAt time.Time
			err = s.cassandraSession.Query(`
				SELECT status, extra_data, id, invite_code, opponent_league_id, opponent_user_id, role, updated_at
				FROM league_joins
				WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?
			`, userID, "1", joinMonth, joinedAt).Scan(&status, &extraData, &id, &inviteCode, &opponentLeagueID, &opponentUserID, &role, &updatedAt)
			if err != nil {
				continue
			}
			// Insert new row with status_id = '3' in league_joins (preserving opponent details)
			err = s.cassandraSession.Query(`
				INSERT INTO league_joins (user_id, status_id, join_month, joined_at, league_id, status, extra_data, id, invite_code, opponent_league_id, opponent_user_id, role, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`).Bind(userID, "3", joinMonth, joinedAt, leagueID, status, extraData, id, inviteCode, opponentLeagueID, opponentUserID, role, updatedAt).Exec()
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
			`, "1", joinDay, leagueID).Iter()
			var pendingID gocql.UUID
			var pendingOpponentUserID, pendingUserID string
			var pendingJoinedAt time.Time
			for pendingIter.Scan(&pendingID, &pendingOpponentUserID, &pendingUserID, &pendingJoinedAt) {
				if pendingUserID != userID {
					continue
				}
				// Insert new row with status_id = '3' (preserving opponent details)
				err = s.cassandraSession.Query(`
					INSERT INTO pending_league_joins (status_id, join_day, league_id, joined_at, id, opponent_user_id, user_id)
					VALUES (?, ?, ?, ?, ?, ?, ?)
				`).Bind("3", joinDay, leagueID, pendingJoinedAt, pendingID, pendingOpponentUserID, userID).Exec()
				if err != nil {
					continue
				}
				// Delete old row
				err = s.cassandraSession.Query(`
					DELETE FROM pending_league_joins WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?
				`, "1", joinDay, leagueID, pendingJoinedAt).Exec()
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
