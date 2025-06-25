# Simple JWT Token Implementation

## Overview

This implementation uses a simplified JWT token approach with only 3 essential fields:
- `mobile_no`
- `device_id` 
- `fcm_token`

## How It Works

### 1. Token Creation (During OTP Verification)

```go
// Generate simple JWT token with only 3 fields
jwtToken, err := utils.GenerateSimpleJWTToken(mobileNo, deviceID, fcmToken)
```

**Process:**
1. Create `SimpleJWTData` structure with 3 fields
2. Convert to JSON
3. Encrypt using AES-256-GCM with secret key
4. Create JWT claims with encrypted data
5. Sign with JWT secret key
6. Store token in database

### 2. Token Storage

Only **one token** is stored in the database:
```go
session.JWTToken = jwtToken
```

### 3. Token Validation (During main:screen)

```go
// Decrypt and validate simple JWT token
simpleJWTData, err := utils.ValidateSimpleJWTToken(jwtToken)
```

**Process:**
1. Parse JWT token
2. Decrypt encrypted data using secret key
3. Extract 3 fields: `mobile_no`, `device_id`, `fcm_token`
4. Validate all fields are present
5. Use decrypted values for validation (not request values)

## Security Benefits

1. **Single Source of Truth**: Only one token stored in database
2. **Encrypted Data**: All 3 fields are encrypted using AES-256-GCM
3. **Secret Key Protection**: Uses same secret key for encryption/decryption
4. **Tamper Prevention**: Server uses decrypted token values, not request values
5. **No Session ID/User ID**: Simplified approach with only essential fields

## Code Flow

### OTP Verification → Token Creation
```
1. User sends OTP
2. Server validates OTP
3. Server generates simple JWT token (3 fields only)
4. Server stores token in session
5. Server sends token to client
```

### Main Screen → Token Validation
```
1. Client sends main:screen request with JWT token
2. Server decrypts JWT token using secret key
3. Server extracts 3 fields from decrypted data
4. Server validates using decrypted values (not request values)
5. Server processes request and sends response
```

## Key Functions

### Token Generation
```go
func GenerateSimpleJWTToken(mobileNo, deviceID, fcmToken string) (string, error)
```

### Token Validation
```go
func ValidateSimpleJWTToken(tokenString string) (*SimpleJWTData, error)
```

### Data Structure
```go
type SimpleJWTData struct {
    MobileNo string `json:"mobile_no"`
    DeviceID string `json:"device_id"`
    FCMToken string `json:"fcm_token"`
}
```

## Testing

Run the test suite to see the simplified JWT token in action:

```bash
node socket_test_suite.js
```

The test will show:
- Token creation with 3 fields
- Token decryption using secret key
- Validation using decrypted values
- Security against tampering

## Advantages

1. **Simplified**: Only 3 essential fields
2. **Secure**: Encrypted data with secret key
3. **Efficient**: Single token storage
4. **Tamper-proof**: Uses decrypted values for validation
5. **Consistent**: Same secret key for encryption/decryption

This approach provides a clean, secure, and efficient JWT token system focused on the essential authentication data. 