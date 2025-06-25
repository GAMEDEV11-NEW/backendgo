# 🔐 Encrypted JWT Authentication System

This document describes the enhanced JWT authentication system with encryption capabilities and mobile number validation.

## 🚀 Features

### ✨ Core Features
- **AES-256-GCM Encryption**: JWT tokens contain encrypted data for enhanced security
- **Mobile Number Validation**: Automatic validation that mobile number in token matches the request
- **Session Management**: Proper session handling for both new and existing users
- **Backward Compatibility**: Support for both encrypted and regular JWT tokens
- **Token Refresh**: Ability to refresh tokens with extended expiry
- **Device Tracking**: Device-specific session management

### 🔒 Security Features
- **Dual-layer Security**: JWT signature + AES encryption
- **Mobile Number Matching**: Prevents token reuse across different users
- **Session Consistency**: Ensures session tokens match JWT tokens
- **Expiration Validation**: Multiple layers of expiration checking
- **Device Binding**: Tokens are bound to specific devices

## 📋 API Endpoints

### 🔑 Login Flow

#### 1. Login Request
```http
POST /login
Content-Type: application/json

{
  "mobile_no": "9876543210",
  "device_id": "device_12345",
  "fcm_token": "fcm_token_here...",
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Login successful",
  "mobile_no": "9876543210",
  "device_id": "device_12345",
  "session_token": "session_abc123",
  "jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "otp": 123456,
  "is_new_user": false,
  "timestamp": "2024-01-15T10:30:00Z",
  "event": "login:success"
}
```

#### 2. OTP Verification
```http
POST /verify-otp
Content-Type: application/json

{
  "mobile_no": "9876543210",
  "session_token": "session_abc123",
  "otp": "123456"
}
```

### 🏠 Main Screen (Protected Endpoint)
```http
POST /main-screen
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "mobile_no": "9876543210",
  "fcm_token": "fcm_token_here...",
  "jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "device_id": "device_12345",
  "message_type": "game_list"
}
```

## 🔧 Implementation Details

### JWT Token Structure

#### Encrypted JWT Token
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "mobile_no": "9876543210",
    "device_id": "device_12345",
    "user_id": "user_67890",
    "encrypted_data": "base64_encrypted_data_here...",
    "exp": 1705312200,
    "iat": 1705225800,
    "nbf": 1705225800,
    "iss": "game-admin-backend",
    "sub": "9876543210"
  }
}
```

#### Encrypted Data Structure
```json
{
  "mobile_no": "9876543210",
  "device_id": "device_12345",
  "user_id": "user_67890",
  "session_id": "session_abc123",
  "created_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-01-16T10:30:00Z"
}
```

### 🔐 Encryption Process

1. **Data Preparation**: User data is marshaled to JSON
2. **AES-256-GCM Encryption**: Data is encrypted using AES-256-GCM with a random nonce
3. **Base64 Encoding**: Encrypted data is base64 encoded
4. **JWT Creation**: Encrypted data is embedded in JWT claims
5. **JWT Signing**: Token is signed with HMAC-SHA256

### 🔍 Validation Process

1. **JWT Signature Verification**: Verify JWT signature using secret key
2. **Mobile Number Matching**: Compare mobile number in token with request
3. **Data Decryption**: Decrypt the encrypted data using AES-256-GCM
4. **Claim Validation**: Verify all claims match expected values
5. **Expiration Check**: Ensure token hasn't expired
6. **Session Validation**: Verify session exists and is active

## 🛠️ Usage Examples

### Generate Encrypted JWT Token
```go
import "gofiber/app/utils"

// Generate encrypted JWT token
token, err := utils.GenerateEncryptedJWTToken(
    "9876543210",    // mobile number
    "device_12345",  // device ID
    "user_67890",    // user ID
    "session_abc123" // session ID
)
if err != nil {
    log.Fatal(err)
}
```

### Validate Encrypted JWT Token
```go
// Validate token and get decrypted data
encryptedData, err := utils.ValidateEncryptedJWTToken(token)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Mobile: %s\n", encryptedData.MobileNo)
fmt.Printf("Device: %s\n", encryptedData.DeviceID)
fmt.Printf("User: %s\n", encryptedData.UserID)
fmt.Printf("Session: %s\n", encryptedData.SessionID)
```

### Validate Mobile Number in Token
```go
// Validate that mobile number in token matches expected
err := utils.ValidateMobileNumberInToken(token, "9876543210")
if err != nil {
    log.Fatal("Mobile number mismatch:", err)
}
```

### Refresh JWT Token
```go
// Refresh token with extended expiry
newToken, err := utils.RefreshJWTToken(oldToken)
if err != nil {
    log.Fatal(err)
}
```

## 🔧 Middleware Usage

### Required JWT Middleware
```go
import "gofiber/app/middlewares"

// Apply to protected routes
app.Use("/protected", middlewares.JWTMiddleware(usersCollection, sessionsCollection))
```

### Optional JWT Middleware
```go
// Apply to routes where JWT is optional
app.Use("/optional", middlewares.OptionalJWTMiddleware(usersCollection, sessionsCollection))
```

### Access User Data in Handlers
```go
func ProtectedHandler(c *fiber.Ctx) error {
    // Get user from context
    user, err := middlewares.GetUserFromContext(c)
    if err != nil {
        return err
    }
    
    // Get session from context
    session, err := middlewares.GetSessionFromContext(c)
    if err != nil {
        return err
    }
    
    // Get encrypted JWT data from context
    jwtData, err := middlewares.GetEncryptedJWTDataFromContext(c)
    if err != nil {
        return err
    }
    
    // Use the data...
    return c.JSON(fiber.Map{
        "user_id": user.ID,
        "mobile_no": user.MobileNo,
        "session_id": jwtData.SessionID,
    })
}
```

## 🔒 Security Considerations

### Secret Keys
- **JWT Secret Key**: Used for signing JWT tokens
- **Encryption Secret Key**: 32-byte key for AES-256-GCM encryption
- **Recommendation**: Store keys in environment variables, not in code

### Token Security
- **Expiration**: Tokens expire after 24 hours
- **Device Binding**: Tokens are bound to specific devices
- **Mobile Number Validation**: Prevents token reuse across users
- **Session Consistency**: Ensures session and JWT tokens match

### Best Practices
1. **Rotate Keys**: Regularly rotate secret keys
2. **HTTPS Only**: Always use HTTPS in production
3. **Token Storage**: Store tokens securely on client side
4. **Session Cleanup**: Regularly clean up expired sessions
5. **Rate Limiting**: Implement rate limiting for login attempts

## 🧪 Testing

### Run Test Script
```bash
go run test_jwt_encryption.go
```

### Expected Output
```
🔐 Testing Encrypted JWT Token System
=====================================
📱 Mobile Number: 9876543210
📱 Device ID: device_12345
👤 User ID: user_67890
🔑 Session ID: session_abc123

🔐 Generating encrypted JWT token...
✅ Encrypted JWT Token generated: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

🔍 Validating encrypted JWT token...
✅ Encrypted JWT Token validated successfully!
📱 Decrypted Mobile Number: 9876543210
📱 Decrypted Device ID: device_12345
👤 Decrypted User ID: user_67890
🔑 Decrypted Session ID: session_abc123
🕐 Created At: 2024-01-15T10:30:00Z
⏰ Expires At: 2024-01-16T10:30:00Z

📱 Testing mobile number validation...
✅ Mobile number validation successful!

❌ Testing with wrong mobile number...
✅ Correctly rejected wrong mobile number: mobile number mismatch: expected 1234567890, got 9876543210

🔄 Testing token refresh...
✅ New token generated: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
✅ Refreshed token validated successfully!
📱 Mobile Number: 9876543210
🕐 New Created At: 2024-01-15T10:30:00Z
⏰ New Expires At: 2024-01-16T10:30:00Z

🎉 All tests completed successfully!
=====================================
📋 Summary:
✅ Encrypted JWT token generation and validation
✅ Mobile number matching and validation
✅ Token refresh functionality
✅ AES-256-GCM encryption/decryption
✅ Session ID tracking
✅ Expiration time validation
```

## 🔄 Migration from Regular JWT

The system maintains backward compatibility with regular JWT tokens. Existing tokens will continue to work, but new tokens will use the encrypted format.

### Migration Steps
1. **Deploy New Code**: Deploy the updated code with encrypted JWT support
2. **Gradual Migration**: New logins will automatically use encrypted tokens
3. **Token Refresh**: Existing users will get encrypted tokens on next login
4. **Cleanup**: Old sessions will expire naturally

## 📝 Configuration

### Environment Variables
```bash
# JWT Configuration
JWT_SECRET_KEY=your-super-secret-jwt-key-for-game-admin-backend-2024
ENCRYPTION_SECRET_KEY=your-32-byte-encryption-key-here!!

# Token Expiration (in hours)
JWT_EXPIRY_HOURS=24
```

### Database Indexes
```javascript
// Sessions collection indexes
db.sessions.createIndex({ "mobile_no": 1, "device_id": 1, "is_active": 1 })
db.sessions.createIndex({ "expires_at": 1 })
db.sessions.createIndex({ "session_token": 1 })

// Users collection indexes
db.users.createIndex({ "mobile_no": 1 })
```

## 🐛 Troubleshooting

### Common Issues

#### 1. JWT Token Validation Failed
- **Cause**: Invalid token signature or expired token
- **Solution**: Check token format and expiration time

#### 2. Mobile Number Mismatch
- **Cause**: Token was generated for different mobile number
- **Solution**: Ensure mobile number in request matches token

#### 3. Session Not Found
- **Cause**: Session expired or was deleted
- **Solution**: User needs to login again

#### 4. Encryption Error
- **Cause**: Wrong encryption key or corrupted token
- **Solution**: Check encryption key configuration

### Debug Logging
Enable debug logging to troubleshoot issues:
```go
log.SetLevel(log.DebugLevel)
```

## 📞 Support

For issues or questions about the encrypted JWT system:
1. Check the logs for detailed error messages
2. Verify configuration settings
3. Test with the provided test script
4. Review the security considerations

---

**Note**: This system provides enhanced security through encryption and mobile number validation. Always follow security best practices and keep secret keys secure. 