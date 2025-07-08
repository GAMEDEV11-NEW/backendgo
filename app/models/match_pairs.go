package models

import (
	"time"

	"github.com/gocql/gocql"
)

// MatchPair represents a match pair between two users
type MatchPair struct {
	ID        gocql.UUID `json:"id" cql:"id"`
	User1ID   string     `json:"user1_id" cql:"user1_id"`
	User2ID   string     `json:"user2_id" cql:"user2_id"`
	User1Data string     `json:"user1_data" cql:"user1_data"` // JSON string containing user1 data
	User2Data string     `json:"user2_data" cql:"user2_data"` // JSON string containing user2 data
	Status    string     `json:"status" cql:"status"`         // e.g., "pending", "active", "completed", "cancelled"
	CreatedAt time.Time  `json:"created_at" cql:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" cql:"updated_at"`
}

// MatchPairStatus constants
const (
	MatchStatusPending   = "pending"
	MatchStatusActive    = "active"
	MatchStatusCompleted = "completed"
	MatchStatusCancelled = "cancelled"
)

// CreateMatchPairRequest represents the request to create a new match pair
type CreateMatchPairRequest struct {
	User1ID   string `json:"user1_id" validate:"required"`
	User2ID   string `json:"user2_id" validate:"required"`
	User1Data string `json:"user1_data" validate:"required"`
	User2Data string `json:"user2_data" validate:"required"`
	Status    string `json:"status"`
}

// UpdateMatchPairRequest represents the request to update a match pair
type UpdateMatchPairRequest struct {
	Status    string `json:"status"`
	User1Data string `json:"user1_data"`
	User2Data string `json:"user2_data"`
}

// MatchPairResponse represents the response for match pair operations
type MatchPairResponse struct {
	ID        string    `json:"id"`
	User1ID   string    `json:"user1_id"`
	User2ID   string    `json:"user2_id"`
	User1Data string    `json:"user1_data"`
	User2Data string    `json:"user2_data"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MatchUsersRequest represents the request to match two users
type MatchUsersRequest struct {
	User1ID   string `json:"user1_id" validate:"required"`
	User2ID   string `json:"user2_id" validate:"required"`
	User1Data string `json:"user1_data" validate:"required"`
	User2Data string `json:"user2_data" validate:"required"`
}
