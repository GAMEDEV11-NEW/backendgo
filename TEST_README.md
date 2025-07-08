# 🧪 GOSOCKET Login Flow Test

A comprehensive Node.js test suite for testing the authentication flow of the GOSOCKET backend application.

## 📋 Test Coverage

This test suite covers the complete login flow:

1. **Health Check** - Server and database connectivity
2. **Version Check** - API version endpoint
3. **Login** - Initial login with mobile number and device info
4. **OTP Verification** - Verify the received OTP
5. **Profile Setup** - Set up user profile (optional)
6. **Language Setup** - Configure language preferences (optional)
7. **Protected Endpoints** - Test JWT-protected endpoints
8. **Logout** - Test logout functionality

## 🚀 Quick Start

### Prerequisites

1. **Node.js 14+** installed
2. **GOSOCKET backend** running on port 8088
3. **Required databases** (Cassandra, Redis) running

### Installation

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Run the test:**
   ```bash
   npm test
   ```

### Command Line Options

```bash
# Basic test
node test_login_flow.js

# Custom server URL
node test_login_flow.js --url http://localhost:8088

# Custom test data
node test_login_flow.js --mobile 9876543210 --device custom_device_456

# Show help
node test_login_flow.js --help
```

## 📊 Test Flow

```
┌─────────────────┐
│   Health Check  │ ✅ Server & DB connectivity
└─────────┬───────┘
          │
┌─────────▼─────────┐
│   Version Check   │ ✅ API version info
└─────────┬─────────┘
          │
┌─────────▼─────────┐
│      Login        │ ✅ Send mobile + device info
└─────────┬─────────┘
          │
┌─────────▼─────────┐
│  OTP Verification │ ✅ Verify received OTP
└─────────┬─────────┘
          │
┌─────────▼─────────┐
│  Profile Setup    │ ✅ Set user profile (optional)
└─────────┬─────────┘
          │
┌─────────▼─────────┐
│ Language Setup    │ ✅ Configure language (optional)
└─────────┬─────────┘
          │
┌─────────▼─────────┐
│Protected Endpoints│ ✅ Test JWT-protected routes
└─────────┬─────────┘
          │
┌─────────▼─────────┐
│     Logout        │ ✅ Test logout functionality
└───────────────────┘
```

## 🔧 Configuration

### Test Data

Default test data in `test_login_flow.js`:

```javascript
const TEST_DATA = {
    mobileNo: '1234567890',
    deviceId: 'test_device_123',
    fcmToken: 'fcm_token_that_is_at_least_100_characters_long...',
    email: 'test@example.com',
    fullName: 'Test User',
    state: 'Test State',
    languageCode: 'en',
    languageName: 'English'
};
```

### Environment Variables

The test uses these default settings:
- **Base URL**: `http://localhost:8088`
- **Timeout**: 10 seconds per request
- **Test Mobile**: `1234567890`
- **Test Device ID**: `test_device_123`

## 📝 Expected API Endpoints

The test expects these HTTP endpoints to be available:

### Public Endpoints
- `GET /health` - Health check
- `GET /api/version` - API version
- `POST /api/auth/login` - Login
- `POST /api/auth/verify-otp` - OTP verification
- `POST /api/auth/set-profile` - Profile setup
- `POST /api/auth/set-language` - Language setup

### Protected Endpoints (require JWT)
- `GET /api/auth/profile` - Get user profile
- `POST /api/auth/main-screen` - Main screen data
- `POST /api/auth/logout` - Logout

## 🎯 Test Scenarios

### 1. Happy Path
- All endpoints working correctly
- Proper authentication flow
- JWT token generation and validation

### 2. Error Handling
- Invalid mobile numbers
- Missing required fields
- Invalid OTP codes
- Expired session tokens

### 3. Edge Cases
- Network timeouts
- Server errors
- Database connectivity issues

## 📊 Output Example

```
🚀 Starting GOSOCKET Login Flow Tests
📱 Test Mobile: 1234567890
📱 Test Device ID: test_device_123
🌐 Base URL: http://localhost:8088
============================================================

1. Health Check
ℹ️  Testing server health and database connectivity...
✅ Health check passed
Health Status: {
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z",
  "services": {
    "cassandra": "ok",
    "redis": "ok"
  }
}

2. Version Check
ℹ️  Testing API version endpoint...
✅ Version check passed
Version Info: {
  "version": "1.0.0",
  "name": "GOSOCKET",
  "timestamp": "2024-01-15T10:30:00Z"
}

3. Login Test
ℹ️  Testing login endpoint...
✅ Login successful
Session Token: abc123def456
OTP: 123456

4. OTP Verification Test
ℹ️  Testing OTP verification endpoint...
✅ OTP verification successful
JWT Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

============================================================
📊 Test Results Summary
============================================================
✅ PASS - Health Check
✅ PASS - Version Check
✅ PASS - Login
✅ PASS - OTP Verification
✅ PASS - Profile Setup
✅ PASS - Language Setup
✅ PASS - Protected Endpoints
✅ PASS - Logout

============================================================
🎯 Overall Result: 8/8 tests passed
🎉 All tests passed! Login flow is working correctly.
```

## 🔍 Troubleshooting

### Common Issues

1. **Connection Refused**
   ```
   ❌ Health check failed: connect ECONNREFUSED 127.0.0.1:8088
   ```
   **Solution**: Ensure the Go backend is running on port 8088

2. **Missing Dependencies**
   ```
   ❌ axios package is required. Please install it with: npm install axios
   ```
   **Solution**: Run `npm install`

3. **Database Connection Issues**
   ```
   ❌ Health check failed: Cassandra connection failed
   ```
   **Solution**: Ensure Cassandra and Redis are running

4. **Route Not Found**
   ```
   ❌ Login failed: 404 Not Found
   ```
   **Solution**: Check if auth routes are properly registered in the Go backend

### Debug Mode

To see detailed request/response data, the test automatically logs:
- Request payloads
- Response data
- Error details
- Status codes

## 🛠️ Customization

### Adding New Tests

1. Create a new test function:
   ```javascript
   async function testNewFeature() {
       logStep('9. New Feature Test', 'Testing new feature...');
       
       const result = await makeRequest('POST', '/api/new-feature', data);
       
       if (result.success) {
           logSuccess('New feature test passed');
           return true;
       } else {
           logError('New feature test failed');
           return false;
       }
   }
   ```

2. Add to the main test runner:
   ```javascript
   results.newFeature = await testNewFeature();
   ```

### Modifying Test Data

Edit the `TEST_DATA` object in `test_login_flow.js`:

```javascript
const TEST_DATA = {
    mobileNo: 'your_test_mobile',
    deviceId: 'your_test_device',
    // ... other fields
};
```

## 📚 Related Files

- `test_login_flow.js` - Main test file
- `package.json` - Node.js dependencies
- `test_otp_flow.py` - Python Socket.IO test (alternative)
- `AUTHENTICATION_GUIDE.md` - Authentication documentation

## 🤝 Contributing

To add new test cases or improve existing ones:

1. Fork the repository
2. Add your test cases
3. Update this README
4. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details. 