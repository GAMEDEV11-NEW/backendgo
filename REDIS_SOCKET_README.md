# Redis + Socket.IO Real-Time Game List Updates

This system demonstrates real-time updates between Redis cache and Socket.IO clients. When the Python script updates the game list in Redis, all connected Socket.IO clients receive instant notifications.

## 🏗️ System Architecture

```
Python Script → Redis Cache → Go Server → Socket.IO Clients
     ↓              ↓            ↓            ↓
Updates Game   Stores Data   Broadcasts   Receive Real-
List Data      with TTL      to Clients   Time Updates
```

## 📋 Prerequisites

1. **Redis Server** running on `127.0.0.1:6379`
2. **Go Server** running on port `3000`
3. **Python** with required dependencies

## 🚀 Setup Instructions

### 1. Install Python Dependencies

```bash
pip install -r requirements.txt
```

### 2. Start Redis Server

**Windows:**
```bash
# If Redis is installed via chocolatey
redis-server

# Or download Redis for Windows and run
redis-server.exe
```

**Linux/Mac:**
```bash
redis-server
```

### 3. Start Go Server

```bash
go run main.go
```

The server will start on `http://localhost:3000`

### 4. Test Redis Connection

```bash
python redis_test.py
```

This will test basic Redis connectivity and operations.

## 🎮 Usage

### Option 1: Interactive Game List Updater

```bash
python game_list_updater.py
```

This provides an interactive menu with options:
- **Single update**: Update game list once
- **Continuous updates**: Update every 30 seconds
- **Custom continuous updates**: Set your own interval
- **Test Redis connection**: Verify Redis is working

### Option 2: Direct Python Script Usage

```python
from game_list_updater import GameListUpdater

# Initialize updater
updater = GameListUpdater()

# Single update
updater.update_and_notify()

# Continuous updates every 60 seconds
updater.continuous_updates(60, max_updates=10)
```

### Option 3: Test with Web Client

1. Open `test_client.html` in your browser
2. Click "Connect to Socket.IO"
3. Run the Python updater script
4. Watch real-time updates in the browser!

## 📡 How It Works

### 1. Python Script Updates Redis

```python
# Updates game list in Redis with 5-minute TTL
updater.update_game_list_in_redis(cache_duration=300)

# Sends notification to all connected Socket.IO clients
updater.send_game_list_update_notification(game_list_data)
```

### 2. Go Server Receives and Broadcasts

The Go Socket.IO server:
- Listens for `game_list:updated` events
- Broadcasts updates to all connected clients
- Works in both main namespace and `/gameplay` namespace

### 3. Clients Receive Real-Time Updates

Connected clients receive:
```json
{
  "status": "success",
  "message": "Game list has been updated",
  "data": {
    "gamelist": [...],
    "updated_at": "2024-01-01T12:00:00Z",
    "total_games": 5,
    "active_games": 3,
    "total_players": 15000
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "event": "game_list:updated"
}
```

## 🎯 Features

### Redis Caching
- **Automatic TTL**: Game lists expire after 5 minutes
- **JSON Storage**: Complex data structures supported
- **Connection Pooling**: Efficient Redis connections

### Real-Time Updates
- **Instant Broadcasting**: All clients notified immediately
- **Multiple Namespaces**: Works in main and gameplay namespaces
- **Error Handling**: Graceful connection failures

### Game List Generation
- **Random Games**: Generates realistic game data
- **Player Counts**: Simulates active and live players
- **Game Status**: Active, maintenance, coming soon states

## 🔧 Configuration

### Redis Settings
```python
# In game_list_updater.py
updater = GameListUpdater(
    redis_host='127.0.0.1',
    redis_port=6379,
    socketio_url='http://localhost:3000'
)
```

### Go Server Settings
```go
// In redis/redis_service.go
const (
    RedisURL      = "localhost:6379"
    RedisPassword = ""
    RedisDB       = 0
)
```

## 📊 Monitoring

### Redis Commands
```bash
# Check if Redis is running
redis-cli ping

# View current game list
redis-cli get gamelist:current

# Monitor Redis operations
redis-cli monitor

# Check memory usage
redis-cli info memory
```

### Go Server Logs
The Go server logs all Redis operations:
- `📝 Redis SET: gamelist:current`
- `📖 Redis GET: gamelist:current`
- `📡 Game list update broadcasted to all connected clients`

## 🐛 Troubleshooting

### Redis Connection Issues
```bash
# Check if Redis is running
redis-cli ping

# Start Redis if not running
redis-server

# Check Redis logs
redis-server --loglevel verbose
```

### Socket.IO Connection Issues
1. Ensure Go server is running on port 3000
2. Check browser console for connection errors
3. Verify CORS settings if needed

### Python Dependencies
```bash
# Install missing dependencies
pip install redis python-socketio requests

# Check installed versions
pip list | grep -E "(redis|socketio|requests)"
```

## 🎨 Customization

### Add New Game Types
```python
# In game_list_updater.py
self.game_templates = [
    {"name": "Your Game", "type": "custom"},
    # ... existing games
]
```

### Modify Update Intervals
```python
# Update every 2 minutes instead of 30 seconds
updater.continuous_updates(120)
```

### Custom Game List Data
```python
custom_games = [
    {
        "game_name": "Custom Game",
        "active_gamepalye": 1000,
        "livegameplaye": 500,
        "status": "active"
    }
]
updater.update_game_list_in_redis(custom_games)
```

## 📈 Performance Tips

1. **Redis TTL**: Set appropriate cache durations
2. **Batch Updates**: Update multiple games at once
3. **Connection Pooling**: Reuse Redis connections
4. **Error Handling**: Implement retry logic for failed operations

## 🔒 Security Considerations

1. **Redis Security**: Set up Redis authentication
2. **Socket.IO Security**: Implement proper authentication
3. **Input Validation**: Validate all incoming data
4. **Rate Limiting**: Prevent abuse of update endpoints

## 📝 API Reference

### Redis Keys
- `gamelist:current` - Current game list data
- `session:{session_id}` - User session data
- `user:{user_id}` - User data cache

### Socket.IO Events
- `game_list:updated` - Game list update notification
- `main:screen` - Request game list data
- `device:info` - Send device information

### Game List Data Structure
```json
{
  "gamelist": [
    {
      "game_id": "game_1_1234567890",
      "game_name": "Poker",
      "game_type": "card",
      "active_gamepalye": 15000,
      "livegameplaye": 8000,
      "min_players": 2,
      "max_players": 8,
      "status": "active",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "updated_at": "2024-01-01T12:00:00Z",
  "cache_duration": "300 seconds",
  "total_games": 1,
  "active_games": 1,
  "total_players": 15000
}
```

---

**Happy Gaming! 🎮** 