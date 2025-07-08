# 🎲 Dice Game API Documentation

## Overview

This document explains how to use the dice rolling functionality in your gaming application. The system provides real-time dice rolling with unique IDs, fast database lookups, and comprehensive history tracking.

## 🚀 Quick Start

### 1. Database Setup

First, run the database setup to create the required tables:

```bash
python3 DATABASESETUP.py
```

This creates two tables:
- `dice_rolls_lookup` - Fast lookups by game_id + user_id
- `dice_rolls_data` - Complete dice roll data indexed by dice_id

### 2. Socket Connection

Connect to the Socket.IO server:

```javascript
const socket = io('ws://your-server:8080');
```

## 📡 API Endpoints

### 1. Roll a Dice

**Event:** `dice:roll`

**Request:**
```javascript
socket.emit('dice:roll', {
  game_id: 'dice_game_001',
  contest_id: 'contest_456',
  session_token: 'your_session_token',
  device_id: 'your_device_id',
  jwt_token: 'your_jwt_token'
});
```

**Response:**
```javascript
socket.on('dice:roll:response', (response) => {
  console.log('Dice rolled:', response.dice_number);
  console.log('Is winner:', response.data.is_winner);
  console.log('Bonus points:', response.data.bonus_points);
});
```

**Response Structure:**
```json
{
  "status": "success",
  "message": "Dice rolled successfully",
  "game_id": "dice_game_001",
  "user_id": "user_123",
  "dice_id": "550e8400-e29b-41d4-a716-446655440000",
  "dice_number": 6,
  "roll_time": "2024-01-01T12:00:00Z",
  "contest_id": "contest_456",
  "session_token": "your_session_token",
  "device_id": "your_device_id",
  "data": {
    "roll_id": "550e8400-e29b-41d4-a716-446655440000",
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

### 2. Get Dice History

**Event:** `dice:history`

**Request:**
```javascript
socket.emit('dice:history', {
  game_id: 'dice_game_001',
  session_token: 'your_session_token',
  limit: 50  // Optional, default is 50
});
```

**Response:**
```javascript
socket.on('dice:history:response', (response) => {
  console.log('Total rolls:', response.total_rolls);
  console.log('Recent rolls:', response.rolls);
});
```

**Response Structure:**
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
      "dice_id": "550e8400-e29b-41d4-a716-446655440000",
      "dice_number": 6,
      "roll_timestamp": "2024-01-01T12:00:00Z",
      "session_token": "your_session_token",
      "device_id": "your_device_id",
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

## 🔧 Complete Implementation Example

### Frontend (JavaScript)

```javascript
// Connect to Socket.IO server
const socket = io('ws://your-server:8080');

// Authentication (required before dice operations)
socket.emit('gameboard:init', {
  mobile_no: '1234567890',
  jwt_token: 'your_jwt_token',
  device_id: 'your_device_id',
  contest_id: 'contest_456'
});

// Listen for authentication response
socket.on('gameboard:init:response', (response) => {
  console.log('Authenticated:', response.status);
  
  // Now you can roll dice
  rollDice();
});

// Roll a dice
function rollDice() {
  socket.emit('dice:roll', {
    game_id: 'dice_game_001',
    contest_id: 'contest_456',
    session_token: 'your_session_token',
    device_id: 'your_device_id',
    jwt_token: 'your_jwt_token'
  });
}

// Listen for dice roll response
socket.on('dice:roll:response', (response) => {
  if (response.status === 'success') {
    console.log('🎲 Dice rolled:', response.dice_number);
    console.log('🏆 Is winner:', response.data.is_winner);
    console.log('💰 Bonus points:', response.data.bonus_points);
    
    // Update UI with dice result
    updateDiceUI(response);
  } else {
    console.error('Dice roll failed:', response.message);
  }
});

// Get dice history
function getDiceHistory() {
  socket.emit('dice:history', {
    game_id: 'dice_game_001',
    session_token: 'your_session_token',
    limit: 10
  });
}

// Listen for history response
socket.on('dice:history:response', (response) => {
  if (response.status === 'success') {
    console.log('📊 Total rolls:', response.total_rolls);
    console.log('📈 Recent rolls:', response.rolls);
    
    // Update UI with history
    updateHistoryUI(response);
  } else {
    console.error('History retrieval failed:', response.message);
  }
});

// Update UI functions
function updateDiceUI(response) {
  const diceElement = document.getElementById('dice-result');
  diceElement.textContent = response.dice_number;
  
  if (response.data.is_winner) {
    diceElement.classList.add('winner');
  }
  
  // Show bonus points
  const bonusElement = document.getElementById('bonus-points');
  bonusElement.textContent = `+${response.data.bonus_points} points`;
}

function updateHistoryUI(response) {
  const historyElement = document.getElementById('dice-history');
  historyElement.innerHTML = '';
  
  response.rolls.forEach(roll => {
    const rollElement = document.createElement('div');
    rollElement.className = 'roll-item';
    rollElement.innerHTML = `
      <span class="dice-number">${roll.dice_number}</span>
      <span class="roll-time">${new Date(roll.roll_timestamp).toLocaleTimeString()}</span>
    `;
    historyElement.appendChild(rollElement);
  });
}

// Error handling
socket.on('connection_error', (error) => {
  console.error('Connection error:', error);
});

socket.on('authentication_error', (error) => {
  console.error('Authentication error:', error);
});
```

### Backend (Go)

The dice functionality is already implemented in your Go backend. Here's how it works:

#### 1. Dice Service

```go
// Create dice service
diceService := services.NewDiceService(cassandraSession)

// Roll a dice
rollReq := models.DiceRollRequest{
    GameID:       "dice_game_001",
    ContestID:    "contest_456",
    SessionToken: "session_token",
    DeviceID:     "device_id",
    JWTToken:     "jwt_token",
}

response, err := diceService.RollDice(rollReq, userID)
```

#### 2. Socket Handlers

The socket handlers are already implemented in `config/gameboard_socket_handler.go`:

```go
// Dice roll handler
socket.On("dice:roll", func(event *socketio.EventPayload) {
    // Authentication and validation
    // Dice roll processing
    // Response emission
})

// Dice history handler  
socket.On("dice:history", func(event *socketio.EventPayload) {
    // Authentication and validation
    // History retrieval
    // Response emission
})
```

## 🎯 Key Features

### Unique Dice IDs
- Each dice roll gets a unique UUID
- No duplicate dice IDs across the system
- Enables fast lookups and data integrity

### Fast Database Queries
- **Lookup Table**: Fast queries by `game_id` + `user_id`
- **Data Table**: Fast queries by `dice_id`
- Two-step process for complex operations

### Random Number Generation
- Generates numbers 1-6 (standard dice)
- Uses Go's `math/rand` with timestamp seeding
- Fair and unpredictable results

### Security Features
- Session token validation
- Device ID verification
- JWT token authentication
- User ID verification

## 📊 Data Structure

### Database Tables

#### 1. `dice_rolls_lookup`
```sql
CREATE TABLE dice_rolls_lookup (
    game_id TEXT,
    user_id TEXT,
    dice_id UUID,
    created_at TIMESTAMP,
    PRIMARY KEY ((game_id, user_id), dice_id)
) WITH CLUSTERING ORDER BY (dice_id DESC);
```

#### 2. `dice_rolls_data`
```sql
CREATE TABLE dice_rolls_data (
    dice_id UUID PRIMARY KEY,
    game_id TEXT,
    user_id TEXT,
    dice_number INT,
    roll_timestamp TIMESTAMP,
    session_token TEXT,
    device_id TEXT,
    contest_id TEXT,
    created_at TIMESTAMP
);
```

### Response Data Fields

| Field | Type | Description |
|-------|------|-------------|
| `dice_id` | UUID | Unique identifier for the dice roll |
| `dice_number` | INT | Random number 1-6 |
| `game_id` | TEXT | Game identifier |
| `user_id` | TEXT | User identifier |
| `contest_id` | TEXT | Contest identifier |
| `roll_time` | TIMESTAMP | When the dice was rolled |
| `is_winner` | BOOLEAN | Whether roll is a winning number (6) |
| `bonus_points` | INT | Points earned (dice_number * 10) |

## ⚠️ Error Handling

### Common Error Responses

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

### Error Codes

| Error Code | Description |
|------------|-------------|
| `MISSING_FIELD` | Required field not provided |
| `INVALID_FORMAT` | Data format is incorrect |
| `INVALID_SESSION` | Session token is invalid or expired |
| `VERIFICATION_ERROR` | Processing error during dice operations |

## 🧪 Testing

### Run Tests

```bash
python3 test_dice.py
```

### Test Coverage

- ✅ Table structure validation
- ✅ Data insertion and retrieval
- ✅ Query performance
- ✅ Data integrity
- ✅ Unique ID generation

## 🚀 Performance Tips

### 1. Connection Management
```javascript
// Reconnect on disconnect
socket.on('disconnect', () => {
  console.log('Disconnected, attempting to reconnect...');
  setTimeout(() => {
    socket.connect();
  }, 1000);
});
```

### 2. Error Handling
```javascript
// Handle all error types
socket.on('connection_error', handleError);
socket.on('authentication_error', handleError);
socket.on('dice:roll:response', (response) => {
  if (response.status === 'error') {
    handleError(response);
  }
});
```

### 3. Rate Limiting
```javascript
// Prevent rapid dice rolling
let canRoll = true;
function rollDice() {
  if (!canRoll) return;
  
  canRoll = false;
  socket.emit('dice:roll', diceData);
  
  setTimeout(() => {
    canRoll = true;
  }, 1000); // 1 second cooldown
}
```

## 📱 Mobile Integration

### React Native Example

```javascript
import io from 'socket.io-client';

const socket = io('ws://your-server:8080');

// Roll dice
const rollDice = () => {
  socket.emit('dice:roll', {
    game_id: 'dice_game_001',
    contest_id: 'contest_456',
    session_token: 'your_session_token',
    device_id: 'your_device_id',
    jwt_token: 'your_jwt_token'
  });
};

// Listen for response
socket.on('dice:roll:response', (response) => {
  if (response.status === 'success') {
    // Update React Native state
    setDiceNumber(response.dice_number);
    setBonusPoints(response.data.bonus_points);
  }
});
```

## 🔧 Configuration

### Environment Variables

```bash
# Cassandra Configuration
CASSANDRA_HOST=172.31.4.229
CASSANDRA_USERNAME=cassandra
CASSANDRA_PASSWORD=cassandra
KEYSPACE=myapp

# Socket.IO Configuration
SOCKET_PORT=8080
SOCKET_CORS_ORIGIN=*
```

### Server Configuration

```go
// In your main.go
socketService := services.NewSocketService(cassandraSession)
gameboardHandler := config.NewGameboardSocketHandler(socketService)

// Setup handlers
gameboardHandler.SetupGameboardHandlers(socket, authFunc)
```

## 📈 Monitoring

### Log Messages

The system logs important events:

```
🎲 Dice roll request from socket_123
🎲 Dice roll stored - GameID: dice_game_001, UserID: user_123, DiceID: uuid, Number: 6
🎲 Dice roll completed - User: user_123, Game: dice_game_001, Number: 6
📊 Dice history request from socket_123
📊 Dice history retrieved - User: user_123, Game: dice_game_001, Total Rolls: 5
```

### Metrics to Monitor

- Dice roll frequency per user
- Average dice numbers
- Error rates
- Response times
- Database query performance

## 🎮 Game Integration Ideas

### 1. Multiplayer Dice Games
```javascript
// Broadcast dice roll to other players
socket.emit('dice:roll', diceData);
socket.on('dice:roll:broadcast', (data) => {
  // Show other player's dice roll
  showOpponentRoll(data);
});
```

### 2. Leaderboards
```javascript
// Get user's dice statistics
socket.emit('dice:stats', {
  game_id: 'dice_game_001',
  session_token: 'your_session_token'
});
```

### 3. Achievements
```javascript
// Check for special dice combinations
if (response.dice_number === 6) {
  unlockAchievement('perfect_roll');
}
```

## 📞 Support

For issues or questions:

1. Check the error logs
2. Verify database connectivity
3. Test with the provided test script
4. Review the `DICE_FUNCTIONALITY.md` for detailed implementation

---

**Happy Rolling! 🎲** 