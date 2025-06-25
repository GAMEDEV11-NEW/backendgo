# 🧪 Socket.IO Test Suite

A professional, comprehensive Node.js test suite for testing Go Socket.IO applications with beautiful UI and complete automation.

## ✨ Features

- **🎨 Beautiful UI**: Professional CLI interface with colors, animations, and progress indicators
- **🔍 Complete Coverage**: Tests all Socket.IO events and scenarios
- **⚡ Automated Testing**: Full automation with configurable timeouts and retries
- **📊 Detailed Reporting**: Comprehensive test results with statistics
- **🔄 Interactive Mode**: User-friendly interactive menu system
- **🛠️ Configurable**: Easy configuration for different environments
- **📈 Real-time Feedback**: Live progress updates and status indicators

## 🚀 Quick Start

### Prerequisites

- Node.js 16.0.0 or higher
- Go Socket.IO server running (default: `http://localhost:8088`)
- MongoDB running (for the Go application)

### Installation

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Start your Go Socket.IO server:**
   ```bash
   go run main.go
   ```

3. **Run the test suite:**
   ```bash
   npm test
   ```

## 📋 Test Coverage

### 🔌 Connection Events
- ✅ Socket connection establishment
- ✅ Connection response validation
- ✅ Welcome message handling
- ✅ Disconnection handling

### 📱 Device Management
- ✅ Device information submission
- ✅ Device validation and acknowledgment
- ✅ Device capabilities handling

### 🔐 Authentication Flow
- ✅ User login with mobile number
- ✅ OTP generation and verification
- ✅ Session token management
- ✅ JWT token handling

### 👤 User Profile
- ✅ Profile setup and validation
- ✅ User preferences management
- ✅ Referral code handling
- ✅ Avatar and bio management

### 🌐 Localization
- ✅ Language settings configuration
- ✅ Timezone and region handling
- ✅ User preferences for formatting
- ✅ Localized message responses

### 📨 Static Messages
- ✅ Game list retrieval
- ✅ Announcements handling
- ✅ Update notifications
- ✅ Dashboard data

### ❌ Error Handling
- ✅ Invalid data validation
- ✅ Missing field detection
- ✅ Format error handling
- ✅ Authentication error responses

## 🎮 Usage

### Interactive Mode (Recommended)

```bash
npm start
```

This launches the interactive menu where you can:
- Run complete test suite
- Run individual test categories
- Configure test settings
- View documentation

### Command Line Options

```bash
# Run complete test suite
npm test

# Quick test run
npm run test:quick

# Verbose output
npm run test:verbose
```

### Programmatic Usage

```javascript
const { SocketIOTestSuite } = require('./socket_test_suite.js');

async function runTests() {
    const testSuite = new SocketIOTestSuite();
    await testSuite.runAllTests();
}

runTests();
```

## ⚙️ Configuration

### Environment Variables

```bash
# Server configuration
SOCKET_SERVER_URL=http://localhost:8088
SOCKET_NAMESPACE=/
SOCKET_TIMEOUT=10000
SOCKET_RETRY_ATTEMPTS=3
```

### Test Configuration Object

```javascript
const CONFIG = {
    SERVER_URL: 'http://localhost:8088',
    SOCKET_NAMESPACE: '/',
    TIMEOUT: 10000,
    RETRY_ATTEMPTS: 3,
    DELAY_BETWEEN_TESTS: 1000
};
```

## 📊 Test Results

The test suite provides comprehensive reporting:

### Summary Statistics
- Total tests executed
- Passed tests count
- Failed tests count
- Success percentage

### Detailed Results
- Individual test status
- Error messages for failed tests
- Execution time per test
- Test category breakdown

### Example Output
```
📊 TEST RESULTS SUMMARY
================================================================================
┌──────────────────────────────┬───────────────┬───────────────┐
│ Metric                       │ Count         │ Percentage    │
├──────────────────────────────┼───────────────┼───────────────┤
│ ✅ Passed                    │ 12            │ 100.0%        │
│ ❌ Failed                    │ 0             │ 0.0%          │
│ ⏭️  Skipped                  │ 0             │ 0.0%          │
│ 📋 Total                     │ 12            │ 100%          │
└──────────────────────────────┴───────────────┴───────────────┘

🎉 ALL TESTS PASSED! Your Socket.IO application is working perfectly!
```

## 🧪 Individual Test Categories

### Connection Tests
```bash
# Test socket connection and basic communication
npm run test:connection
```

### Authentication Tests
```bash
# Test login, OTP verification, and session management
npm run test:auth
```

### Profile Tests
```bash
# Test user profile setup and management
npm run test:profile
```

### Language Tests
```bash
# Test localization and language settings
npm run test:language
```

### Device Tests
```bash
# Test device registration and validation
npm run test:device
```

### Error Handling Tests
```bash
# Test error scenarios and validation
npm run test:error
```

## 🔧 Customization

### Adding New Tests

1. **Create test method in SocketIOTestSuite class:**
   ```javascript
   async testCustomFeature() {
       await this.runTest('Custom Feature Test', async () => {
           // Your test logic here
           const response = await this.waitForEvent('custom:event');
           // Validate response
       });
   }
   ```

2. **Add to runAllTests method:**
   ```javascript
   await this.testCustomFeature();
   ```

### Custom Test Data

```javascript
// Modify test data in constructor
this.testData = {
    deviceId: 'custom_device_id',
    mobileNo: '+1234567890',
    fcmToken: 'custom_fcm_token',
    // ... other custom data
};
```

## 🐛 Troubleshooting

### Common Issues

1. **Connection Failed**
   - Ensure Go server is running on correct port
   - Check firewall settings
   - Verify server URL in configuration

2. **Timeout Errors**
   - Increase timeout value in configuration
   - Check server performance
   - Verify network connectivity

3. **Authentication Failures**
   - Check MongoDB connection
   - Verify user data in database
   - Check OTP generation logic

### Debug Mode

Enable verbose logging:
```bash
DEBUG=socket.io* npm test
```

## 📚 API Reference

### SocketIOTestSuite Class

#### Methods
- `connect()` - Establish socket connection
- `disconnect()` - Close socket connection
- `runTest(name, function)` - Execute individual test
- `waitForEvent(eventName, timeout)` - Wait for specific event
- `runAllTests()` - Execute complete test suite

#### Properties
- `socket` - Socket.IO client instance
- `testData` - Test data storage
- `eventListeners` - Event listener management

### Test Data Structure

```javascript
{
    deviceId: 'device_123456789',
    mobileNo: '+1234567890',
    fcmToken: 'fcm_token_123456',
    sessionToken: 'session_123456789',
    otp: 123456,
    socketId: 'socket_123456'
}
```

## 🤝 Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit pull request

## 📄 License

MIT License - see LICENSE file for details

## 🆘 Support

For issues and questions:
- Create GitHub issue
- Check troubleshooting section
- Review API documentation

---

**Made with ❤️ for Go Socket.IO applications** 