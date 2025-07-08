# Socket.IO Authentication System Guide

## Overview

This application implements a **centralized authentication system** that validates user authentication for all Socket.IO events. The system ensures that only authenticated users can access protected endpoints while allowing authentication-related events to proceed without validation.

## Authentication Flow

### 1. Initial Connection
- Socket connects to the server
- Authorization handler allows all connections (configurable for production)
- Welcome message and connection response sent

### 2. Authentication Events (No Validation Required)
The following events are **exempt** from authentication checks:
- `device:info` - Device information setup
- `login` - User login with OTP
- `verify:otp` - OTP verification
- `verify:user` - JWT token verification
- `connect` - Connection events
- `disconnect` - Disconnection events
- `connect_response` - Connection responses

### 3. Protected Events (Authentication Required)
All other events require valid authentication:
- `set:profile` - User profile setup
- `set:language` - Language preference setup
- `main:screen` - Main screen data
- `trigger_game_list_update` - Game list updates
- `list:contest` - Contest listings
- `list:contest:gap` - Contest price gaps
- `contest:join` - Contest joining
- `check:opponent` - Opponent checking
- `cancel:find` - Cancel matchmaking
- `heartbeat` - Health monitoring
- `health_check` - System health
- `ping` - Connection testing
- `gameboard:init` - Gameboard initialization
- And all other game-related events

## Implementation Details

### Centralized Authentication Function

```go
func (h *SocketIoHandler) authenticateUser(socket *socketio.Socket, eventName string) (*models.User, error)
```

This function:
1. **Checks if the event is exempt** from authentication
2. **Queries the sessions_by_socket table** to get user session info
3. **Validates user exists** in the database
4. **Checks session expiration** and active status
5. **Verifies session token** matches stored token
6. **Returns user object** if authentication succeeds

### Authentication Error Handling

The system provides detailed error responses:

```go
type AuthenticationError struct {
    *models.ConnectionError
}

func (e *AuthenticationError) Error() string {
    return e.Message
}
```

### Event Handler Integration

Each protected event handler includes authentication:

```go
socket.On("main:screen", func(event *socketio.EventPayload) {
    // Authenticate user
    _, err := authFunc(socket, "main:screen")
    if err != nil {
        if authErr, ok := err.(*AuthenticationError); ok {
            socket.Emit("authentication_error", authErr.ConnectionError)
        } else {
            socket.Emit("connection_error", models.ConnectionError{
                Status:    "error",
                ErrorCode: models.ErrorCodeInvalidSession,
                ErrorType: models.ErrorTypeAuthentication,
                Field:     "authentication",
                Message:   err.Error(),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
                SocketID:  socket.Id,
                Event:     "connection_error",
            })
        }
        return
    }
    
    // Proceed with event handling...
})
```

## Database Schema

### Sessions Table
```sql
CREATE TABLE sessions (
    mobile_no text,
    device_id text,
    session_token text,
    jwt_token text,
    fcm_token text,
    expires_at timestamp,
    is_active boolean,
    PRIMARY KEY (mobile_no, device_id)
);
```

### Sessions by Socket Table
```sql
CREATE TABLE sessions_by_socket (
    socket_id text PRIMARY KEY,
    mobile_no text,
    user_id text,
    session_token text,
    created_at timestamp
);
```

### Users Table
```sql
CREATE TABLE users (
    id text,
    mobile_no text,
    full_name text,
    status text,
    language_code text,
    PRIMARY KEY (id)
);
```

## Error Codes

The system uses standardized error codes:

- `ErrorCodeInvalidSession` - Session not found or expired
- `ErrorCodeMissingField` - Required field missing
- `ErrorCodeInvalidFormat` - Invalid data format
- `ErrorCodeVerificationError` - Verification failed
- `ErrorCodeInvalidOTP` - Invalid OTP code

## Security Features

### ✅ Implemented
- **Session-based authentication** with expiration
- **Device ID validation** to prevent session hijacking
- **Mobile number verification** against stored data
- **JWT token validation** for returning users
- **Comprehensive error handling** with detailed messages
- **Centralized authentication** for all events
- **Session cleanup** on disconnect

### 🔧 Production Recommendations

1. **Rate Limiting**
   ```go
   // Add rate limiting for authentication attempts
   rateLimiter := rate.NewLimiter(rate.Every(time.Minute), 5)
   ```

2. **IP-based Restrictions**
   ```go
   // Validate client IP addresses
   if !isAllowedIP(clientIP) {
       return false
   }
   ```

3. **Audit Logging**
   ```go
   // Log all authentication attempts
   log.Printf("Auth attempt: %s from IP: %s", eventName, clientIP)
   ```

4. **Session Hijacking Protection**
   ```go
   // Add additional security headers
   // Implement token rotation
   ```

## Usage Examples

### Client-Side Authentication Flow

```javascript
// 1. Connect to socket
const socket = io('http://localhost:3000');

// 2. Send device info (no auth required)
socket.emit('device:info', {
    device_id: 'device123',
    device_type: 'mobile'
});

// 3. Login (no auth required)
socket.emit('login', {
    mobile_no: '+1234567890',
    device_id: 'device123'
});

// 4. Verify OTP (no auth required)
socket.emit('verify:otp', {
    mobile_no: '+1234567890',
    session_token: 'session123',
    otp: '123456'
});

// 5. Now all other events require authentication
socket.emit('main:screen', {
    // This will be authenticated automatically
    message_type: 'game_list'
});
```

### Error Handling

```javascript
// Listen for authentication errors
socket.on('authentication_error', (error) => {
    console.error('Authentication failed:', error);
    // Redirect to login
});

// Listen for general connection errors
socket.on('connection_error', (error) => {
    console.error('Connection error:', error);
});
```

## Testing

### Test Authentication Flow

1. **Connect without authentication**
   ```bash
   # Should work - connection allowed
   curl -X POST http://localhost:3000/socket.io/
   ```

2. **Try protected event without auth**
   ```javascript
   // Should fail with authentication_error
   socket.emit('main:screen', {});
   ```

3. **Complete authentication flow**
   ```javascript
   // Should succeed after proper authentication
   socket.emit('device:info', {...});
   socket.emit('login', {...});
   socket.emit('verify:otp', {...});
   socket.emit('main:screen', {...}); // Now works
   ```

## Monitoring

### Log Messages

The system provides detailed logging:

```
✅ Socket connected: abc123 (namespace: /)
✅ User authenticated for event main:screen: +1234567890 (socket: abc123)
❌ Authentication failed for event main:screen: User not authenticated
```

### Metrics to Monitor

- Authentication success/failure rates
- Session expiration patterns
- Device ID validation failures
- JWT token validation errors
- Database query performance

## Troubleshooting

### Common Issues

1. **"User not authenticated"**
   - Check if user completed login flow
   - Verify session exists in database
   - Check session expiration

2. **"Session expired"**
   - User needs to re-authenticate
   - Check session cleanup processes

3. **"Device ID mismatch"**
   - Verify device ID consistency
   - Check for session hijacking attempts

### Debug Commands

```sql
-- Check active sessions
SELECT * FROM sessions WHERE is_active = true;

-- Check socket mappings
SELECT * FROM sessions_by_socket;

-- Check user data
SELECT * FROM users WHERE mobile_no = 'your_mobile';
```

This authentication system provides a robust, secure foundation for your Socket.IO application while maintaining flexibility for different types of events. 