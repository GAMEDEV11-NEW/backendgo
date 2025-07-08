# 🎲 Dice Functionality Implementation

## Overview

This implementation provides a complete dice rolling system with a two-table database structure for optimal performance and data organization.

## Database Structure

### Table 1: `dice_rolls_lookup` (Fast Lookup Table)
- **Primary Key**: `(game_id, user_id)` (partition key) + `dice_id` (clustering key)
- **Purpose**: Fast lookups by game_id and user_id combination
- **Columns**:
  - `game_id` (TEXT)
  - `user_id` (TEXT) 
  - `dice_id` (UUID) - Unique identifier for each dice roll
  - `created_at` (TIMESTAMP)

### Table 2: `dice_rolls_data` (Full Data Table)
- **Primary Key**: `dice_id` (UUID) - Primary key only
- **Purpose**: Store complete dice roll information
- **Columns**:
  - `dice_id` (UUID) - Primary key
  - `game_id` (TEXT)
  - `user_id` (TEXT)
  - `dice_number` (INT) - Random number 1-6
  - `roll_timestamp` (TIMESTAMP)
  - `session_token` (TEXT)
  - `device_id` (TEXT)
  - `contest_id` (TEXT)
  - `created_at` (TIMESTAMP)

## Key Features

### 🎯 Unique Dice IDs
- Each dice roll gets a unique UUID (`dice_id`)
- No duplicate dice IDs across the system
- Enables fast lookups and data integrity

### ⚡ Fast Lookups
- **Table 1** optimized for queries by `game_id` + `user_id`
- **Table 2** optimized for queries by `dice_id`
- Efficient two-step query process for complex operations

### 🔒 Data Integrity
- Both tables maintain referential integrity
- Atomic operations ensure data consistency
- Session validation for security

## API Endpoints

### 1. Dice Roll Event: `dice:roll`
**Request:**
```json
{
  "game_id": "dice_game_001",
  "contest_id": "contest_456", 
  "session_token": "session_token_789",
  "device_id": "device_abc",
  "jwt_token": "jwt_token_xyz"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Dice rolled successfully",
  "game_id": "dice_game_001",
  "user_id": "user_123",
  "dice_id": "uuid-here",
  "dice_number": 6,
  "roll_time": "2024-01-01T12:00:00Z",
  "contest_id": "contest_456",
  "session_token": "session_token_789",
  "device_id": "device_abc",
  "data": {
    "roll_id": "uuid-here",
    "roll_timestamp": "2024-01-01T12:00:00Z",
    "game_name": "Dice Game",
    "contest_name": "contest_456",
    "is_winner": true,
    "bonus_points": 60
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "socket_id": "socket_123",
  "event": "dice:roll:response"
}
```

### 2. Dice History Event: `dice:history`
**Request:**
```json
{
  "game_id": "dice_game_001",
  "session_token": "session_token_789",
  "limit": 50
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Dice history retrieved successfully",
  "game_id": "dice_game_001",
  "user_id": "user_123",
  "rolls": [
    {
      "game_id": "dice_game_001",
      "user_id": "user_123",
      "dice_id": "uuid-1",
      "dice_number": 6,
      "roll_timestamp": "2024-01-01T12:00:00Z",
      "session_token": "session_token_789",
      "device_id": "device_abc",
      "contest_id": "contest_456",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "total_rolls": 1,
  "data": {
    "total_rolls": 1,
    "limit": 50,
    "game_id": "dice_game_001",
    "user_id": "user_123"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "socket_id": "socket_123",
  "event": "dice:history:response"
}
```

## Implementation Details

### Database Operations

#### 1. Dice Roll Storage
```go
// Step 1: Insert into lookup table for fast queries
INSERT INTO dice_rolls_lookup (game_id, user_id, dice_id, created_at)
VALUES (?, ?, ?, ?)

// Step 2: Insert into data table for full information
INSERT INTO dice_rolls_data (dice_id, game_id, user_id, dice_number, roll_timestamp, session_token, device_id, contest_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
```

#### 2. Dice History Retrieval
```go
// Step 1: Get dice IDs from lookup table
SELECT dice_id FROM dice_rolls_lookup 
WHERE game_id = ? AND user_id = ?
LIMIT ?

// Step 2: Get full data for each dice ID
SELECT * FROM dice_rolls_data WHERE dice_id = ?
```

#### 3. Dice Statistics
```go
// Get all dice IDs for a user/game combination
SELECT dice_id FROM dice_rolls_lookup 
WHERE game_id = ? AND user_id = ?

// Get statistics from data table
SELECT dice_number, contest_id FROM dice_rolls_data WHERE dice_id = ?
```

### Random Number Generation
- Uses Go's `math/rand` package
- Generates numbers 1-6 (standard dice)
- Seeds with current timestamp for randomness

### Security Features
- Session token validation
- Device ID verification
- JWT token authentication
- User ID verification

## Performance Benefits

### 🚀 Fast Queries
- **Lookup Table**: O(1) queries by game_id + user_id
- **Data Table**: O(1) queries by dice_id
- **Composite Queries**: Two-step process for complex operations

### 💾 Efficient Storage
- **Lookup Table**: Minimal data, fast scans
- **Data Table**: Complete information, indexed by unique ID
- **No Duplication**: Each dice_id appears only once in each table

### 🔍 Scalable Design
- **Horizontal Scaling**: Partitioned by game_id + user_id
- **Vertical Scaling**: Separate tables for different query patterns
- **Future-Proof**: Easy to add new query patterns

## Usage Examples

### Frontend Integration
```javascript
// Roll a dice
socket.emit('dice:roll', {
  game_id: 'dice_game_001',
  contest_id: 'contest_456',
  session_token: 'session_token_789',
  device_id: 'device_abc',
  jwt_token: 'jwt_token_xyz'
});

// Listen for response
socket.on('dice:roll:response', (response) => {
  console.log('Dice rolled:', response.dice_number);
  console.log('Is winner:', response.data.is_winner);
});

// Get dice history
socket.emit('dice:history', {
  game_id: 'dice_game_001',
  session_token: 'session_token_789',
  limit: 50
});

// Listen for history response
socket.on('dice:history:response', (response) => {
  console.log('Total rolls:', response.total_rolls);
  console.log('Recent rolls:', response.rolls);
});
```

## Testing

Run the test script to verify functionality:
```bash
python3 test_dice.py
```

This will test:
- Table structure validation
- Data insertion and retrieval
- Query performance
- Data integrity

## Error Handling

### Common Error Codes
- `MISSING_FIELD`: Required field not provided
- `INVALID_FORMAT`: Data format is incorrect
- `INVALID_SESSION`: Session token is invalid or expired
- `VERIFICATION_ERROR`: Processing error during dice operations

### Error Response Format
```json
{
  "status": "error",
  "error_code": "MISSING_FIELD",
  "error_type": "FIELD_ERROR",
  "field": "game_id",
  "message": "Game ID is required",
  "timestamp": "2024-01-01T12:00:00Z",
  "socket_id": "socket_123",
  "event": "connection_error"
}
```

## Future Enhancements

### Potential Features
- **Dice Statistics**: Average, highest, lowest rolls
- **Leaderboards**: Compare dice rolls across users
- **Achievements**: Special rewards for specific dice combinations
- **Real-time Updates**: Live dice roll broadcasts
- **Multi-player**: Simultaneous dice rolling

### Performance Optimizations
- **Caching**: Redis cache for frequent queries
- **Batch Operations**: Bulk dice roll processing
- **Indexing**: Additional indexes for complex queries
- **Partitioning**: Time-based partitioning for large datasets 