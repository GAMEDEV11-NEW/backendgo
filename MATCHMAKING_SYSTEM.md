# 🎯 Contest Matchmaking System

## Overview

The contest matchmaking system has been separated into two distinct processes:

1. **Contest Join Process** - Simple user registration for contests
2. **Background Matchmaking Process** - Automated cron job that pairs users

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Joins    │    │   Background    │    │   Match Found   │
│   Contest       │    │   Cron Job      │    │   Notification  │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  contest:join   │    │  ProcessMatch-  │    │  Update Users   │
│  Event          │    │  making()       │    │  with Opponent  │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Store in       │    │  Find Pending   │    │  Create Game    │
│  league_joins   │    │  Users          │    │  Pieces         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Process Flow

### 1. Contest Join Process (Simplified)

**Event:** `contest:join`

**What it does:**
- ✅ Validates user authentication
- ✅ Validates contest data
- ✅ Stores user in `league_joins` table
- ✅ Stores user in `pending_league_joins` table
- ✅ Returns simple success response

**What it doesn't do:**
- ❌ No opponent search
- ❌ No match creation
- ❌ No game pieces creation
- ❌ No real-time matchmaking

### 2. Background Matchmaking Process

**Cron Job:** Runs every 30 seconds

**What it does:**
- 🔄 Queries `pending_league_joins` for users with `status_id = "1"`
- 🔄 Groups users by `league_id` (contest)
- 🔄 Sorts users by join time (oldest first)
- 🔄 Pairs users in order (1st with 2nd, 3rd with 4th, etc.)
- 🔄 Creates match pairs in `match_pairs` table
- 🔄 Updates both users with opponent information
- 🔄 Creates game pieces for matched users
- 🔄 Logs matchmaking statistics

## Database Tables

### `league_joins` (Main Storage)
```sql
PRIMARY KEY ((user_id, status_id, join_month), joined_at DESC)
```

### `pending_league_joins` (Fast Lookup)
```sql
PRIMARY KEY ((status_id, join_month, join_day, league_id), joined_at)
```

### `match_pairs` (Match Storage)
```sql
PRIMARY KEY (id)
```

## API Endpoints

### Health Check
```http
GET /health
```

### Matchmaking Status
```http
GET /api/matchmaking/status
```

**Response:**
```json
{
  "status": "success",
  "cron_running": true,
  "stats": {
    "pending_users": 5,
    "active_matches": 12,
    "last_run": "2024-01-15T10:30:00Z"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Configuration

### Cron Job Intervals
- **Matchmaking:** Every 30 seconds
- **Cleanup:** Every 5 minutes (removes matches older than 24 hours)

### Environment Variables
```env
# Matchmaking Configuration
MATCHMAKING_INTERVAL=30s
CLEANUP_INTERVAL=5m
MATCH_EXPIRY=24h
```

## Benefits of This Architecture

### ✅ **Separation of Concerns**
- Join process is fast and simple
- Matchmaking runs independently
- No blocking on user requests

### ✅ **Scalability**
- Multiple users can join simultaneously
- Background process handles pairing
- No real-time performance impact

### ✅ **Reliability**
- Failed matches don't affect join process
- Retry logic in background
- Cleanup prevents data accumulation

### ✅ **Monitoring**
- Detailed logging of matchmaking process
- Statistics endpoint for monitoring
- Health check integration

## Usage Example

### 1. User Joins Contest
```javascript
socket.emit('contest:join', {
  mobile_no: '1234567890',
  fcm_token: 'fcm_token_here',
  jwt_token: 'jwt_token_here',
  device_id: 'device_123',
  contest_id: 'contest_456'
});
```

### 2. Background Matchmaking
- Cron job runs every 30 seconds
- Finds pending users in same contest
- Creates matches automatically
- Updates user records with opponent info

### 3. Check Match Status
```javascript
// Check if user has been matched
socket.emit('check:opponent', {
  user_id: 'user_123',
  contest_id: 'contest_456'
});
```

## Monitoring

### Log Messages
- `🔄 Starting matchmaking process...`
- `✅ Created match: user1 vs user2 in league contest_456`
- `🎯 Matchmaking complete: 5 total matches created`
- `📊 Stats: {pending_users: 3, active_matches: 15}`

### Statistics
- Pending users count
- Active matches count
- Last run timestamp
- Cron job status

## Troubleshooting

### Common Issues

1. **No matches being created**
   - Check if users are in `pending_league_joins` with `status_id = "1"`
   - Verify cron job is running
   - Check logs for errors

2. **Users not being paired**
   - Ensure users joined the same contest
   - Check join timestamps
   - Verify database connectivity

3. **High pending user count**
   - Check if matchmaking cron is running
   - Verify database queries
   - Check for errors in logs

### Debug Commands
```bash
# Check matchmaking status
curl http://localhost:8088/api/matchmaking/status

# Check health
curl http://localhost:8088/health
``` 