package services

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

// MatchmakingService handles contest matchmaking logic
type MatchmakingService struct {
	cassandraSession *gocql.Session
	gameService      *GameService
	socketService    *SocketService
}

// NewMatchmakingService creates a new matchmaking service instance
func NewMatchmakingService(cassandraSession *gocql.Session) *MatchmakingService {
	return &MatchmakingService{
		cassandraSession: cassandraSession,
		gameService:      NewGameService(cassandraSession),
		socketService:    NewSocketService(cassandraSession),
	}
}

// PendingMatch represents a user waiting for a match
type PendingMatch struct {
	UserID   string    `json:"user_id"`
	LeagueID string    `json:"league_id"`
	JoinedAt time.Time `json:"joined_at"`
	ID       string    `json:"id"`
	StatusID string    `json:"status_id"`
	JoinDay  string    `json:"join_day"`
}

// MatchResult represents a successful match
type MatchResult struct {
	User1ID   string    `json:"user1_id"`
	User2ID   string    `json:"user2_id"`
	LeagueID  string    `json:"league_id"`
	MatchID   string    `json:"match_id"`
	CreatedAt time.Time `json:"created_at"`
	GameID    string    `json:"game_id"`
}

// ProcessMatchmaking runs the matchmaking algorithm for all pending users
func (m *MatchmakingService) ProcessMatchmaking() error {
	// DEBUG: Find pending users for status_id=1, today's join_day, and league_id 1-10
	debugJoinDay := time.Now().Format("2006-01-02")
	statusID := "1"
	for i := 1; i <= 10; i++ {
		leagueID := fmt.Sprintf("%d", i)
		iter := m.cassandraSession.Query(`
			SELECT user_id, join_day, league_id FROM pending_league_joins
			WHERE status_id = ? AND join_day = ? AND league_id = ?
			LIMIT 100
		`, statusID, debugJoinDay, leagueID).Iter()
		var userID, joinDayResult, leagueIDResult string
		found := false
		for iter.Scan(&userID, &joinDayResult, &leagueIDResult) {
			found = true
		}
		iter.Close()
		if !found {
		}
	}

	// Main matchmaking loop: for each league_id 1-10 and today's join_day
	matchCount := 0
	mainJoinDay := time.Now().Format("2006-01-02")
	mainStatusID := "1"
	for i := 1; i <= 10; i++ {
		leagueID := fmt.Sprintf("%d", i)
		iter := m.cassandraSession.Query(`
			SELECT user_id, league_id, joined_at, id, status_id, join_day
			FROM pending_league_joins
			WHERE status_id = ? AND join_day = ? AND league_id = ?
			ORDER BY joined_at ASC LIMIT 2
		`, mainStatusID, mainJoinDay, leagueID).Iter()

		var users []PendingMatch
		var userID, leagueIDVal, statusIDVal, joinDayVal string
		var joinedAt time.Time
		var id gocql.UUID
		for iter.Scan(&userID, &leagueIDVal, &joinedAt, &id, &statusIDVal, &joinDayVal) {
			users = append(users, PendingMatch{
				UserID:   userID,
				LeagueID: leagueIDVal,
				JoinedAt: joinedAt,
				ID:       id.String(),
				StatusID: statusIDVal,
				JoinDay:  joinDayVal,
			})
		}
		iter.Close()

		if len(users) == 2 {
			_, err := m.createMatch(users[0], users[1])
			if err != nil {
				continue
			}
			matchCount++
		}
	}
	return nil
}

// createDiceRolls creates dice roll entries for both users in a match
func (m *MatchmakingService) createDiceRolls(gameID, user1ID, user2ID string) error {
	now := time.Now()

	// Create dice lookup for user1
	user1DiceID := gocql.TimeUUID()
	err := m.cassandraSession.Query(`
		INSERT INTO dice_rolls_lookup (game_id, user_id, dice_id, created_at)
		VALUES (?, ?, ?, ?)
	`, gameID, user1ID, user1DiceID, now).Exec()
	if err != nil {
		return fmt.Errorf("failed to create dice lookup for user1: %v", err)
	}

	// Create dice lookup for user2
	user2DiceID := gocql.TimeUUID()
	err = m.cassandraSession.Query(`
		INSERT INTO dice_rolls_lookup (game_id, user_id, dice_id, created_at)
		VALUES (?, ?, ?, ?)
	`, gameID, user2ID, user2DiceID, now).Exec()
	if err != nil {
		return fmt.Errorf("failed to create dice lookup for user2: %v", err)
	}

	return nil
}

// createMatch creates a match between two users
func (m *MatchmakingService) createMatch(user1, user2 PendingMatch) (*MatchResult, error) {
	// Store only user IDs in user1_data and user2_data columns
	user1Data := user1.UserID
	user2Data := user2.UserID

	// Create match pair entry with user IDs
	matchPairID := gocql.TimeUUID()
	now := time.Now()

	// Store league_joins.id for user1_id and user2_id, plus user IDs in data columns
	err := m.cassandraSession.Query(`
		INSERT INTO match_pairs (id, user1_id, user2_id, user1_data, user2_data, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, matchPairID, user1.ID, user2.ID, user1Data, user2Data, "active", now, now).Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to create match pair: %v", err)
	}

	// Determine turnID based on join time
	var turnID1, turnID2 int
	if user1.JoinedAt.Before(user2.JoinedAt) {
		turnID1 = 1
		turnID2 = 2
	} else {
		turnID1 = 2
		turnID2 = 1
	}
	// Update both users' league_joins entries with opponent info
	_ = m.updateUserWithOpponent(user1, user2.UserID, user1.LeagueID, matchPairID, turnID1)
	_ = m.updateUserWithOpponent(user2, user1.UserID, user2.LeagueID, matchPairID, turnID2)

	// Update pending_league_joins entries with opponent info
	_ = m.updatePendingWithOpponent(user1, user2.UserID)
	_ = m.updatePendingWithOpponent(user2, user1.UserID)

	// Delete both users from pending_league_joins so they are not matched again
	deleteQuery := `DELETE FROM pending_league_joins WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?`
	for _, user := range []PendingMatch{user1, user2} {
		err := m.cassandraSession.Query(deleteQuery, user.StatusID, user.JoinDay, user.LeagueID, user.JoinedAt).Exec()
		if err != nil {
		}
	}

	// Create game pieces for both users
	err = m.socketService.CreateGamePieces(matchPairID.String(), user1.UserID, user2.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to create game pieces: %v", err)
	}

	// Create dice rolls for both users
	err = m.createDiceRolls(matchPairID.String(), user1.UserID, user2.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to create dice rolls: %v", err)
	}

	return &MatchResult{
		User1ID:   user1.UserID,
		User2ID:   user2.UserID,
		LeagueID:  user1.LeagueID,
		MatchID:   matchPairID.String(),
		CreatedAt: now,
		GameID:    matchPairID.String(),
	}, nil
}

// updateUserWithOpponent updates a user's league_joins entry with opponent info
func (m *MatchmakingService) updateUserWithOpponent(user PendingMatch, opponentUserID, leagueID string, matchPairID gocql.UUID, turnID int) error {
	joinMonth := user.JoinedAt.Format("2006-01")
	err := m.cassandraSession.Query(
		`UPDATE league_joins SET opponent_user_id = ?, opponent_league_id = ?, match_pair_id = ?, turn_id = ? WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?`,
		opponentUserID, leagueID, matchPairID, turnID, user.UserID, user.StatusID, joinMonth, user.JoinedAt,
	).Exec()
	if err != nil {
		return fmt.Errorf("failed to update league_joins: %v", err)
	}
	return nil
}

// updatePendingWithOpponent updates a user's pending_league_joins entry with opponent info
func (m *MatchmakingService) updatePendingWithOpponent(user PendingMatch, opponentUserID string) error {
	return m.cassandraSession.Query(`
		UPDATE pending_league_joins SET opponent_user_id = ?
		WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?
	`, opponentUserID, user.StatusID, user.JoinDay, user.LeagueID, user.JoinedAt).Exec()
}

// GetMatchmakingStats returns statistics about the matchmaking process
func (m *MatchmakingService) GetMatchmakingStats() (map[string]interface{}, error) {
	// Count pending users (approximate, for monitoring only)
	var pendingCount int
	iter := m.cassandraSession.Query(`
		SELECT user_id FROM pending_league_joins WHERE status_id = ? LIMIT 10000
	`, "1").Iter()
	var tmp string
	for iter.Scan(&tmp) {
		pendingCount++
	}
	iter.Close()

	// Count active matches (approximate, for monitoring only)
	var activeMatches int
	iter2 := m.cassandraSession.Query(`
		SELECT id FROM match_pairs LIMIT 10000
	`).Iter()
	var tmpID gocql.UUID
	for iter2.Scan(&tmpID) {
		activeMatches++
	}
	iter2.Close()

	return map[string]interface{}{
		"pending_users":  pendingCount,
		"active_matches": activeMatches,
		"last_run":       time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// CleanupExpiredMatches removes matches that are older than the specified duration
func (m *MatchmakingService) CleanupExpiredMatches(maxAge time.Duration) error {
	cutoffTime := time.Now().Add(-maxAge)
	iter := m.cassandraSession.Query(`
		SELECT id, created_at FROM match_pairs WHERE created_at < ? ALLOW FILTERING
	`, cutoffTime).Iter()
	var matchID gocql.UUID
	var createdAt time.Time
	var expiredCount int
	for iter.Scan(&matchID, &createdAt) {
		err := m.cassandraSession.Query(`
			UPDATE match_pairs SET status = ? WHERE id = ?
		`, "expired", matchID).Exec()
		if err == nil {
			expiredCount++
		}
	}
	iter.Close()
	if expiredCount > 0 {
	}
	return nil
}
