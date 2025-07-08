package models

import (
	"time"

	"github.com/gocql/gocql"
)

// DiceRoll represents a dice roll in the system
type DiceRoll struct {
	LookupDiceID  gocql.UUID `json:"lookup_dice_id" cql:"lookup_dice_id"`
	RollID        gocql.UUID `json:"roll_id" cql:"roll_id"`
	DiceNumber    int        `json:"dice_number" cql:"dice_number"`
	RollTimestamp time.Time  `json:"roll_timestamp" cql:"roll_timestamp"`
	SessionToken  string     `json:"session_token" cql:"session_token"`
	DeviceID      string     `json:"device_id" cql:"device_id"`
	ContestID     string     `json:"contest_id" cql:"contest_id"`
	CreatedAt     time.Time  `json:"created_at" cql:"created_at"`
}

// DiceRollRequest represents the request to roll a dice
type DiceRollRequest struct {
	GameID       string `json:"game_id" validate:"required"`
	ContestID    string `json:"contest_id" validate:"required"`
	SessionToken string `json:"session_token" validate:"required"`
	DeviceID     string `json:"device_id" validate:"required"`
	JWTToken     string `json:"jwt_token" validate:"required"`
}

// DiceRollResponse represents the response for dice roll
type DiceRollResponse struct {
	Status       string                 `json:"status"`
	Message      string                 `json:"message"`
	GameID       string                 `json:"game_id"`
	UserID       string                 `json:"user_id"`
	DiceID       string                 `json:"dice_id"`
	DiceNumber   int                    `json:"dice_number"`
	RollTime     string                 `json:"roll_time"`
	ContestID    string                 `json:"contest_id"`
	SessionToken string                 `json:"session_token"`
	DeviceID     string                 `json:"device_id"`
	Data         map[string]interface{} `json:"data"`
	Timestamp    string                 `json:"timestamp"`
	SocketID     string                 `json:"socket_id"`
	Event        string                 `json:"event"`
}

// DiceHistoryRequest represents the request to get dice roll history
type DiceHistoryRequest struct {
	GameID       string `json:"game_id" validate:"required"`
	UserID       string `json:"user_id" validate:"required"`
	SessionToken string `json:"session_token" validate:"required"`
	Limit        int    `json:"limit"`
}

// DiceHistoryResponse represents the response for dice roll history
type DiceHistoryResponse struct {
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	GameID     string                 `json:"game_id"`
	UserID     string                 `json:"user_id"`
	Rolls      []DiceRoll             `json:"rolls"`
	TotalRolls int                    `json:"total_rolls"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  string                 `json:"timestamp"`
	SocketID   string                 `json:"socket_id"`
	Event      string                 `json:"event"`
}

// DiceStats represents dice roll statistics
type DiceStats struct {
	TotalRolls  int         `json:"total_rolls"`
	AverageRoll float64     `json:"average_roll"`
	HighestRoll int         `json:"highest_roll"`
	LowestRoll  int         `json:"lowest_roll"`
	RollCounts  map[int]int `json:"roll_counts"`
	RecentRolls []DiceRoll  `json:"recent_rolls"`
	GameID      string      `json:"game_id"`
	UserID      string      `json:"user_id"`
	ContestID   string      `json:"contest_id"`
}
