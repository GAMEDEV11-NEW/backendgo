# Socket.IO Authentication Process Explained

## 🔐 **Complete Authentication Flow**

This document explains the step-by-step authentication process in your Socket.IO application, from initial connection to accessing protected features.

---

## 📋 **Authentication Process Overview**

```
1. Socket Connection → 2. Device Info → 3. Login → 4. OTP Verification → 5. Profile Setup → 6. Access Protected Events
```

---

## 🚀 **Step-by-Step Authentication Process**

### **Step 1: Socket Connection (No Authentication Required)**

```javascript
// Client connects to Socket.IO server
const socket = io('http://localhost:3000');

// Server response
socket.on('connect', (data) => {
    console.log('Connected with token:', data.token);
    // Token: 123456
    // SocketID: abc123
});
```

**Server Side:**
```go
// Socket connection handler
h.io.OnConnection(func(socket *socketio.Socket) {
    log.Printf("✅ Socket connected: %s", socket.Id)
    
    // Send welcome message
    welcome := h.socketService.HandleWelcome()
    socket.Emit("connect_response", welcome)
    
    // Send connect response with token
    connectResp := models.ConnectResponse{
        Token:     123456, // Random token
        Message:   "Welcome to the Game Admin Server!",
        SocketID:  socket.Id,
        Status:    "connected",
        Event:     "connect",
    }
    socket.Emit("connect", connectResp)
})
```

**Database State:** No user session exists yet.

---

### **Step 2: Device Information (No Authentication Required)**

```javascript
// Client sends device information
socket.emit('device:info', {
    device_id: 'device123',
    device_type: 'mobile',
    manufacturer: 'Apple',
    model: 'iPhone 14',
    firmware_version: '16.0'
});
```

**Server Side:**
```go
socket.On("device:info", func(event *socketio.EventPayload) {
    // Parse device info
    deviceInfoData := event.Data[0].(map[string]interface{})
    
    // Convert to DeviceInfo struct
    var deviceInfo models.DeviceInfo
    json.Unmarshal(deviceInfoJSON, &deviceInfo)
    
    // Process device info
    response := h.socketService.HandleDeviceInfo(deviceInfo, socket.Id)
    socket.Emit("device:info:ack", response)
})
```

**Database State:** Device information stored, but no user session yet.

---

### **Step 3: User Login (No Authentication Required)**

```javascript
// Client initiates login
socket.emit('login', {
    mobile_no: '+1234567890',
    device_id: 'device123',
    fcm_token: 'fcm_token_123'
});
```

**Server Side:**
```go
socket.On("login", func(event *socketio.EventPayload) {
    // Parse login request
    loginData := event.Data[0].(map[string]interface{})
    loginData["socket_id"] = socket.Id
    
    // Convert to LoginRequest struct
    var loginReq models.LoginRequest
    json.Unmarshal(loginJSON, &loginReq)
    
    // Process login and generate OTP
    response, err := h.socketService.HandleLogin(loginReq)
    if err != nil {
        socket.Emit("connection_error", errorResp)
        return
    }
    
    // Send OTP to user
    response.SocketID = socket.Id
    socket.Emit("otp:sent", response)
})
```

**Database State:**
- OTP generated and stored
- Session token created
- User status: "pending_otp"

---

### **Step 4: OTP Verification (No Authentication Required)**

```javascript
// Client verifies OTP
socket.emit('verify:otp', {
    mobile_no: '+1234567890',
    session_token: 'session123',
    otp: '123456'
});
```

**Server Side:**
```go
socket.On("verify:otp", func(event *socketio.EventPayload) {
    // Parse OTP request
    otpData := event.Data[0].(map[string]interface{})
    
    // Convert to OTPVerificationRequest struct
    var otpReq models.OTPVerificationRequest
    json.Unmarshal(otpJSON, &otpReq)
    
    // Process OTP verification
    response, err := h.socketService.HandleOTPVerification(otpReq)
    if err != nil {
        socket.Emit("connection_error", errorResp)
        return
    }
    
    // Send verification success
    response.SocketID = socket.Id
    socket.Emit("otp:verified", response)
})
```

**Database State:**
- User session activated
- JWT token generated
- User status: "authenticated"
- Session stored in `sessions` table

---

### **Step 5: Profile Setup (Now Requires Authentication)**

```javascript
// Client sets up profile (requires authentication)
socket.emit('set:profile', {
    mobile_no: '+1234567890',
    session_token: 'session123',
    full_name: 'John Doe',
    state: 'California',
    referral_code: 'REF123'
});
```

**Server Side:**
```go
socket.On("set:profile", func(event *socketio.EventPayload) {
    // 🔐 AUTHENTICATION CHECK
    _, err := authFunc(socket, "set:profile")
    if err != nil {
        socket.Emit("authentication_error", authErr.ConnectionError)
        return
    }
    
    // Parse profile request
    profileData := event.Data[0].(map[string]interface{})
    
    // Convert to SetProfileRequest struct
    var profileReq models.SetProfileRequest
    json.Unmarshal(profileJSON, &profileReq)
    
    // Process profile setup
    response, err := h.socketService.HandleSetProfile(profileReq)
    if err != nil {
        socket.Emit("connection_error", errorResp)
        return
    }
    
    // Send profile setup success
    response.SocketID = socket.Id
    socket.Emit("profile:set", response)
})
```

**Authentication Process:**
```go
func (h *SocketIoHandler) authenticateUser(socket *socketio.Socket, eventName string) (*models.User, error) {
    // 1. Check if event is exempt from authentication
    authEvents := map[string]bool{
        "device:info": true,
        "login": true,
        "verify:otp": true,
        "verify:user": true,
        "connect": true,
        "disconnect": true,
        "connect_response": true,
    }
    
    if authEvents[eventName] {
        return nil, nil // Skip authentication
    }
    
    // 2. Get user session from sessions_by_socket table
    var mobileNo, userID, sessionToken string
    err := h.socketService.GetCassandraSession().Query(`
        SELECT mobile_no, user_id, session_token, created_at
        FROM sessions_by_socket
        WHERE socket_id = ?
    `, socket.Id).Scan(&mobileNo, &userID, &sessionToken, &createdAt)
    
    if err != nil {
        return nil, &AuthenticationError{
            ConnectionError: &models.ConnectionError{
                Status: "error",
                ErrorCode: models.ErrorCodeInvalidSession,
                Message: "User not authenticated. Please login first.",
                SocketID: socket.Id,
                Event: "authentication_error",
            },
        }
    }
    
    // 3. Get user details from users table
    var user models.User
    err = h.socketService.GetCassandraSession().Query(`
        SELECT id, mobile_no, full_name, status, language_code
        FROM users
        WHERE mobile_no = ?
        ALLOW FILTERING
    `).Bind(mobileNo).Scan(&user.ID, &user.MobileNo, &user.FullName, &user.Status, &user.LanguageCode)
    
    if err != nil {
        return nil, &AuthenticationError{
            ConnectionError: &models.ConnectionError{
                Status: "error",
                ErrorCode: models.ErrorCodeInvalidSession,
                Message: "User not found in database",
                SocketID: socket.Id,
                Event: "authentication_error",
            },
        }
    }
    
    // 4. Check if session is still active
    var session models.Session
    err = h.socketService.GetCassandraSession().Query(`
        SELECT session_token, mobile_no, device_id, fcm_token, expires_at, is_active, jwt_token
        FROM sessions
        WHERE mobile_no = ? AND device_id = ? AND is_active = true AND expires_at > ?
        ALLOW FILTERING
    `).Bind(mobileNo, userID, time.Now()).Scan(&session.SessionToken, &session.MobileNo, &session.DeviceID, &session.FCMToken, &session.ExpiresAt, &session.IsActive, &session.JWTToken)
    
    if err != nil {
        return nil, &AuthenticationError{
            ConnectionError: &models.ConnectionError{
                Status: "error",
                ErrorCode: models.ErrorCodeInvalidSession,
                Message: "Session expired or invalid",
                SocketID: socket.Id,
                Event: "authentication_error",
            },
        }
    }
    
    // 5. Verify session token matches
    if session.SessionToken != sessionToken {
        return nil, &AuthenticationError{
            ConnectionError: &models.ConnectionError{
                Status: "error",
                ErrorCode: models.ErrorCodeInvalidSession,
                Message: "Session token mismatch",
                SocketID: socket.Id,
                Event: "authentication_error",
            },
        }
    }
    
    log.Printf("✅ User authenticated for event %s: %s (socket: %s)", eventName, user.MobileNo, socket.Id)
    return &user, nil
}
```

**Database State:**
- User profile updated
- Session remains active
- User status: "profile_complete"

---

### **Step 6: Language Setup (Requires Authentication)**

```javascript
// Client sets language preference (requires authentication)
socket.emit('set:language', {
    mobile_no: '+1234567890',
    session_token: 'session123',
    language_code: 'en',
    language_name: 'English',
    region_code: 'US',
    timezone: 'America/Los_Angeles'
});
```

**Server Side:**
```go
socket.On("set:language", func(event *socketio.EventPayload) {
    // 🔐 AUTHENTICATION CHECK
    _, err := authFunc(socket, "set:language")
    if err != nil {
        socket.Emit("authentication_error", authErr.ConnectionError)
        return
    }
    
    // Parse language request
    langData := event.Data[0].(map[string]interface{})
    
    // Convert to SetLanguageRequest struct
    var langReq models.SetLanguageRequest
    json.Unmarshal(langJSON, &langReq)
    
    // Process language setup
    response, err := h.socketService.HandleSetLanguage(langReq)
    if err != nil {
        socket.Emit("connection_error", errorResp)
        return
    }
    
    // Send language setup success
    response.SocketID = socket.Id
    socket.Emit("language:set", response)
})
```

**Database State:**
- Language preferences updated
- User fully configured
- Ready for game features

---

### **Step 7: Access Protected Events (Requires Authentication)**

```javascript
// Now user can access all protected events
socket.emit('main:screen', {
    message_type: 'game_list'
});

socket.emit('list:contest', {
    mobile_no: '+1234567890',
    fcm_token: 'fcm_token_123',
    jwt_token: 'jwt_token_123',
    device_id: 'device123',
    message_type: 'contest_list'
});

socket.emit('contest:join', {
    mobile_no: '+1234567890',
    fcm_token: 'fcm_token_123',
    jwt_token: 'jwt_token_123',
    device_id: 'device123',
    contest_id: 'contest123'
});
```

**Server Side:**
```go
// All protected events follow the same pattern
socket.On("main:screen", func(event *socketio.EventPayload) {
    // 🔐 AUTHENTICATION CHECK
    _, err := authFunc(socket, "main:screen")
    if err != nil {
        socket.Emit("authentication_error", authErr.ConnectionError)
        return
    }
    
    // Process the event...
    response := h.socketService.HandleMainScreen(mainReq)
    socket.Emit("main:screen:game:list", response.Data)
})
```

---

## 🗄️ **Database Schema & State Changes**

### **Tables Involved:**

1. **`sessions_by_socket`** - Maps socket IDs to user sessions
   ```sql
   CREATE TABLE sessions_by_socket (
       socket_id text PRIMARY KEY,
       mobile_no text,
       user_id text,
       session_token text,
       created_at timestamp
   );
   ```

2. **`sessions`** - Active user sessions
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

3. **`users`** - User information
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

### **State Progression:**

| Step | `sessions_by_socket` | `sessions` | `users` | Authentication Status |
|------|---------------------|------------|---------|---------------------|
| 1. Connect | ❌ Empty | ❌ Empty | ❌ Empty | ❌ Not Authenticated |
| 2. Device Info | ❌ Empty | ❌ Empty | ❌ Empty | ❌ Not Authenticated |
| 3. Login | ❌ Empty | ✅ Created | ❌ Empty | ❌ Pending OTP |
| 4. OTP Verify | ✅ Created | ✅ Active | ✅ Created | ✅ Authenticated |
| 5. Profile Setup | ✅ Exists | ✅ Active | ✅ Updated | ✅ Authenticated |
| 6. Language Setup | ✅ Exists | ✅ Active | ✅ Updated | ✅ Authenticated |
| 7. Protected Events | ✅ Exists | ✅ Active | ✅ Complete | ✅ Authenticated |

---

## 🔍 **Authentication Validation Process**

### **What Happens During Authentication:**

1. **Event Check**: Is this event exempt from authentication?
2. **Session Lookup**: Find user session in `sessions_by_socket` table
3. **User Validation**: Verify user exists in `users` table
4. **Session Validation**: Check session is active and not expired
5. **Token Verification**: Ensure session token matches stored token
6. **Success**: Return user object for event processing

### **Error Scenarios:**

```go
// Scenario 1: User not authenticated
if err != nil {
    return &AuthenticationError{
        ConnectionError: &models.ConnectionError{
            Status: "error",
            ErrorCode: models.ErrorCodeInvalidSession,
            Message: "User not authenticated. Please login first.",
            SocketID: socket.Id,
            Event: "authentication_error",
        },
    }
}

// Scenario 2: User not found
if err != nil {
    return &AuthenticationError{
        ConnectionError: &models.ConnectionError{
            Status: "error",
            ErrorCode: models.ErrorCodeInvalidSession,
            Message: "User not found in database",
            SocketID: socket.Id,
            Event: "authentication_error",
        },
    }
}

// Scenario 3: Session expired
if err != nil {
    return &AuthenticationError{
        ConnectionError: &models.ConnectionError{
            Status: "error",
            ErrorCode: models.ErrorCodeInvalidSession,
            Message: "Session expired or invalid",
            SocketID: socket.Id,
            Event: "authentication_error",
        },
    }
}

// Scenario 4: Token mismatch
if session.SessionToken != sessionToken {
    return &AuthenticationError{
        ConnectionError: &models.ConnectionError{
            Status: "error",
            ErrorCode: models.ErrorCodeInvalidSession,
            Message: "Session token mismatch",
            SocketID: socket.Id,
            Event: "authentication_error",
        },
    }
}
```

---

## 📱 **Client-Side Implementation**

### **Complete Authentication Flow:**

```javascript
// 1. Connect to socket
const socket = io('http://localhost:3000');

// 2. Listen for connection
socket.on('connect', (data) => {
    console.log('Connected:', data);
});

// 3. Send device info
socket.emit('device:info', {
    device_id: 'device123',
    device_type: 'mobile'
});

// 4. Login
socket.emit('login', {
    mobile_no: '+1234567890',
    device_id: 'device123',
    fcm_token: 'fcm_token_123'
});

// 5. Listen for OTP
socket.on('otp:sent', (data) => {
    console.log('OTP sent:', data.otp);
    // Show OTP input to user
});

// 6. Verify OTP
socket.emit('verify:otp', {
    mobile_no: '+1234567890',
    session_token: data.session_token,
    otp: '123456'
});

// 7. Listen for verification success
socket.on('otp:verified', (data) => {
    console.log('OTP verified:', data);
    // Now user is authenticated
});

// 8. Set profile (requires auth)
socket.emit('set:profile', {
    mobile_no: '+1234567890',
    session_token: data.session_token,
    full_name: 'John Doe',
    state: 'California'
});

// 9. Set language (requires auth)
socket.emit('set:language', {
    mobile_no: '+1234567890',
    session_token: data.session_token,
    language_code: 'en',
    language_name: 'English'
});

// 10. Now access protected events
socket.emit('main:screen', {
    message_type: 'game_list'
});

// Error handling
socket.on('authentication_error', (error) => {
    console.error('Authentication failed:', error);
    // Redirect to login
});

socket.on('connection_error', (error) => {
    console.error('Connection error:', error);
});
```

---

## 🔐 **Security Features**

### **Session Management:**
- **Session Expiration**: Sessions expire after 24 hours
- **Device Binding**: Sessions are tied to specific device IDs
- **Token Validation**: Session tokens are verified on each request
- **Active Status**: Only active sessions are accepted

### **Error Handling:**
- **Detailed Error Messages**: Clear, actionable error responses
- **Error Classification**: Different error types for different scenarios
- **Logging**: Comprehensive logging for security monitoring
- **Graceful Degradation**: Proper error responses to clients

### **Database Security:**
- **Session Isolation**: Each socket has its own session mapping
- **User Validation**: Users are verified against database records
- **Token Verification**: Session tokens are cross-validated
- **Expiration Checks**: Automatic session expiration handling

---

## 📊 **Monitoring & Debugging**

### **Log Messages:**

```
✅ Socket connected: abc123 (namespace: /)
✅ User authenticated for event set:profile: +1234567890 (socket: abc123)
❌ Authentication failed for event main:screen: User not authenticated
```

### **Database Queries for Debugging:**

```sql
-- Check active sessions
SELECT * FROM sessions WHERE is_active = true;

-- Check socket mappings
SELECT * FROM sessions_by_socket WHERE socket_id = 'abc123';

-- Check user data
SELECT * FROM users WHERE mobile_no = '+1234567890';

-- Check session expiration
SELECT * FROM sessions WHERE expires_at > now();
```

This authentication system provides a robust, secure foundation for your Socket.IO application with comprehensive session management and error handling. 