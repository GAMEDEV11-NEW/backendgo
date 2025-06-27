# Contest List Testing Guide

This guide explains how to test the `list:contest` event and related contest functionality in your Go Socket.IO application.

## 🏆 Test Scripts Available

### 1. `test_list_contest.js` - Full Authentication + Contest Test
This script performs a complete authentication flow and then tests the contest list functionality.

**Features:**
- Complete login flow (login → OTP verification)
- JWT token generation and validation
- Contest list retrieval
- Contest join functionality
- Detailed logging and error handling

**Usage:**
```bash
# Run contest list only test
node test_list_contest.js

# Run full test (including contest join)
node test_list_contest.js --full
```

### 2. `test_contest_custom.js` - Custom Data Testing
This script allows you to test with your own custom data without going through the full authentication flow.

**Features:**
- Test with custom JWT tokens
- Test with custom contest data
- No authentication required (uses provided tokens)
- Perfect for debugging specific issues

**Usage:**
```bash
# Run with sample data
node test_contest_custom.js
```

### 3. `socket_test_suite.js` - Complete Test Suite
The main test suite that includes contest tests along with all other functionality.

**Usage:**
```bash
# Run complete test suite
node socket_test_suite.js

# Then select "🏆 Contest Tests" from the menu
```

## 🔧 Prerequisites

Before running the tests, make sure:

1. **Go Server is Running:**
   ```bash
   go run main.go
   ```
   Server should be running on `http://localhost:8088`

2. **Redis is Running:**
   ```bash
   # Start Redis server
   redis-server
   ```

3. **Contest Data is Available:**
   ```bash
   # Generate contest data in Redis
   python create_contest_list.py
   ```

4. **Node.js Dependencies:**
   ```bash
   npm install socket.io-client chalk
   ```

## 🧪 Testing Scenarios

### Scenario 1: Basic Contest List Test
```bash
node test_list_contest.js
```
This will:
1. Connect to the server
2. Perform login and OTP verification
3. Test the `list:contest` event
4. Display contest data received

### Scenario 2: Custom Data Test
```bash
node test_contest_custom.js
```
This will:
1. Connect to the server
2. Test with predefined sample data
3. Test both contest list and contest join
4. Show detailed results

### Scenario 3: Full Test Suite
```bash
node socket_test_suite.js
```
Then select "🏆 Contest Tests" from the menu.

## 📊 Expected Output

### Successful Contest List Response
```
🏆 Testing Contest List Event...
ℹ Sending contest data:
  Mobile: +1234567890
  Device ID: test_device
  Message Type: contest_list
  JWT Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
  FCM Token: fcm_test_token_123456789012345678901234567890...

✅ Contest list data retrieved successfully!
ℹ Status: success
ℹ Message: Contest list data retrieved successfully
ℹ Event: contest:list:response
ℹ Mobile: +1234567890
ℹ Device ID: test_device
ℹ Contest list items: 6

🏆 Contest 1:
  Name: Weekly Algorithm Challenge
  ID: 1
  Status: upcoming
  Difficulty: medium
  Duration: 7 days
  Participants: 847/1000
  Reward: USD 5000
  Categories: algorithms, data-structures, dynamic-programming
```

### Error Response Example
```
❌ Contest list data retrieval failed
Error: simple JWT token validation failed: invalid token
```

## 🔍 Troubleshooting

### Issue 1: "Connection timeout"
**Solution:** Make sure the Go server is running on `http://localhost:8088`

### Issue 2: "JWT token validation failed"
**Solution:** 
- Use the full authentication flow in `test_list_contest.js`
- Or generate a valid JWT token using the login process

### Issue 3: "No contest data found"
**Solution:** 
- Run `python create_contest_list.py` to populate Redis
- Check if Redis is running: `redis-cli ping`

### Issue 4: "Contest list data retrieval failed"
**Solution:**
- Check server logs for authentication errors
- Verify the JWT token is valid and not expired
- Ensure all required fields are provided

## 📝 Custom Data Format

### Contest List Request
```javascript
const contestData = {
    mobile_no: "+1234567890",
    fcm_token: "your_fcm_token_here",
    jwt_token: "your_jwt_token_here",
    device_id: "your_device_id",
    message_type: "contest_list"
};
```

### Contest Join Request
```javascript
const joinData = {
    mobile_no: "+1234567890",
    fcm_token: "your_fcm_token_here",
    jwt_token: "your_jwt_token_here",
    device_id: "your_device_id",
    contest_id: "contest_123",
    team_name: "Team Name",
    team_size: 2
};
```

## 🎯 Key Events to Monitor

### Client → Server Events
- `list:contest` - Request contest list
- `contest:join` - Join a contest

### Server → Client Events
- `contest:list:response` - Contest list data
- `contest:join:response` - Contest join confirmation
- `connection_error` - Error responses

## 🔧 Debugging Tips

1. **Check Server Logs:** Look for authentication and validation errors
2. **Verify Redis Data:** Use `redis-cli get listcontest:current`
3. **Test Authentication:** Ensure JWT token is valid
4. **Check Event Names:** Verify event names match exactly
5. **Monitor Network:** Use browser dev tools to see WebSocket traffic

## 📈 Performance Testing

For performance testing, you can modify the scripts to:
- Send multiple requests rapidly
- Test with different data sizes
- Monitor response times
- Test concurrent connections

## 🚀 Next Steps

After successful testing:
1. Integrate contest functionality into your client application
2. Add real-time updates for contest status
3. Implement contest result notifications
4. Add contest analytics and reporting

---

**Happy Testing! 🏆** 