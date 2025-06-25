# Game Admin Socket.IO Backend

A comprehensive Socket.IO server implementation in Go that provides real-time communication for a game admin platform. This implementation follows the exact event structure and data formats specified in the Socket.IO Events Documentation.

## 🚀 Features

- **Real-time Communication**: Full Socket.IO implementation with WebSocket support
- **Authentication System**: Mobile number + OTP verification flow
- **User Management**: Profile setup and language preferences
- **Device Management**: Device information tracking and validation
- **Error Handling**: Comprehensive error responses with detailed codes
- **MongoDB Integration**: Persistent storage for users and sessions
- **Multi-namespace Support**: Separate gameplay namespace
- **Localization**: Multi-language support with localized messages

## 📋 Event Structure

The server implements all events from the documentation:

### Connection Events
- `connect` - Automatic connection response with token
- `disconnect` - Client disconnection handling

### Device Management
- `device:info` → `device:info:ack` - Device information validation

### Authentication
- `login` → `login:success` - User login with OTP generation
- `verify:otp` → `otp:verified` - OTP verification

### User Profile
- `set:profile` → `profile:set` - User profile setup
- `set:language` → `language:set` - Language preferences

### Error Handling
- `connection_error` - Standardized error responses

## 🛠️ Installation

### Prerequisites
- Go 1.22+
- MongoDB 4.4+
- Node.js (for testing)

### Setup

1. **Clone the repository**
```bash
git clone <repository-url>
cd GOSOCEKT
```

2. **Install Go dependencies**
```bash
go mod tidy
```

3. **Install Node.js dependencies (for testing)**
```bash
npm install
```

4. **Configure MongoDB**
```bash
# Update config/config.go with your MongoDB connection string
MongoDBURL = "mongodb://localhost:27017"
DatabaseName = "game_admin"
```

5. **Run the server**
```bash
go run main.go
```

The server will start on port 3000 by default.

## 📡 API Endpoints

### Socket.IO Endpoints
- **Main namespace**: `ws://localhost:3000/socket.io/`
- **Gameplay namespace**: `ws://localhost:3000/socket.io/gameplay`

### HTTP Endpoints
- **Health check**: `GET /health`
- **Static files**: `GET /` (serves from WEBSITE directory)

## 🔧 Configuration

### Environment Variables
```bash
# Server configuration
SERVER_PORT=3000
MONGODB_URL=mongodb://localhost:27017
DATABASE_NAME=game_admin

# Socket.IO configuration
SOCKET_CORS_ORIGIN=*
SOCKET_PING_TIMEOUT=60000
SOCKET_PING_INTERVAL=25000
```

### Database Collections
The server automatically creates the following MongoDB collections:
- `users` - User profiles and preferences
- `sessions` - Active user sessions
- `connect_events` - Connection logs
- `device_info_events` - Device information logs
- `login_events` - Login attempts
- `login_success_events` - Successful logins
- `otp_verification_events` - OTP verifications
- `user_profile_events` - Profile updates
- `language_setting_events` - Language preferences
- `connection_error_events` - Error logs

## 🧪 Testing

### Run the test suite
```bash
node socket_test.js
```

### Individual tests
```bash
# Test connection
node -e "require('./socket_test.js').testConnection()"

# Test device info
node -e "require('./socket_test.js').testDeviceInfo()"

# Test complete login flow
node -e "require('./socket_test.js').testLoginFlow()"

# Test error handling
node -e "require('./socket_test.js').testErrorHandling()"

# Test gameplay namespace
node -e "require('./socket_test.js').testGameplayNamespace()"
```

## 📊 Event Flow Examples

### Complete User Registration Flow
```javascript
// 1. Connect to server
const socket = io('http://localhost:3000');

// 2. Send device info
socket.emit('device:info', {
    device_id: "device_123456789",
    device_type: "mobile",
    timestamp: new Date().toISOString(),
    manufacturer: "Samsung",
    model: "Galaxy S21"
});

// 3. Login with mobile number
socket.emit('login', {
    mobile_no: "+1234567890",
    device_id: "device_123456789",
    fcm_token: "fcm_token_123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
    email: "user@example.com"
});

// 4. Verify OTP
socket.emit('verify:otp', {
    mobile_no: "+1234567890",
    session_token: "session_token_from_login",
    otp: "123456"
});

// 5. Set profile
socket.emit('set:profile', {
    mobile_no: "+1234567890",
    session_token: "session_token",
    full_name: "John Doe",
    state: "California",
    referral_code: "JOHN123"
});

// 6. Set language
socket.emit('set:language', {
    mobile_no: "+1234567890",
    session_token: "session_token",
    language_code: "en",
    language_name: "English",
    region_code: "US"
});
```

## 🔐 Security Features

- **Session Management**: Secure session tokens with expiration
- **Input Validation**: Comprehensive validation for all input data
- **Error Handling**: Detailed error codes for debugging
- **Rate Limiting**: Built-in protection against abuse
- **Data Sanitization**: All input data is sanitized before processing

## 🌐 Supported Languages

- English (en)
- Spanish (es)
- French (fr)
- German (de)
- Hindi (hi)
- Chinese (zh)
- Japanese (ja)
- Korean (ko)
- Arabic (ar)
- Portuguese (pt)
- Russian (ru)

## 📝 Error Codes

### Field Errors
- `MISSING_FIELD` - Required field is missing
- `EMPTY_FIELD` - Field cannot be empty
- `INVALID_FORMAT` - Data format is invalid
- `INVALID_TYPE` - Field has wrong data type

### Authentication Errors
- `INVALID_SESSION` - Session token is invalid
- `INVALID_OTP` - OTP verification failed
- `MAX_ATTEMPTS_EXCEEDED` - Too many OTP attempts
- `SESSION_VERIFICATION_ERROR` - Session verification failed

### System Errors
- `VERIFICATION_ERROR` - System verification error
- `REFERRAL_CODE_EXISTS` - Referral code already exists

## 🏗️ Architecture

```
GOSOCEKT/
├── main.go                 # Application entry point
├── config/
│   ├── config.go          # Configuration settings
│   └── socket_handler.go  # Socket.IO event handlers
├── app/
│   ├── models/
│   │   ├── socket_models.go  # Data structures
│   │   └── loginmodel.go     # Login models
│   ├── services/
│   │   └── socket_service.go # Business logic
│   ├── controllers/          # HTTP controllers
│   ├── middlewares/          # HTTP middlewares
│   └── routes/              # HTTP routes
├── database/
│   └── database.go         # MongoDB connection
├── data/                   # Data files
├── socket_test.js          # Test client
└── README.md              # This file
```

## 🔄 Event Response Format

All events follow a consistent response format:

### Success Response
```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "timestamp": "2024-01-15T10:30:00Z",
  "socket_id": "socket_123456",
  "event": "event_name"
}
```

### Error Response
```json
{
  "status": "error",
  "error_code": "ERROR_CODE",
  "error_type": "ERROR_TYPE",
  "field": "field_name",
  "message": "Human readable error message",
  "details": {
    "additional_info": "Additional error details"
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "socket_id": "socket_123456",
  "event": "connection_error"
}
```

## 🚀 Deployment

### Docker Deployment
```bash
# Build the image
docker build -t game-admin-socket .

# Run the container
docker run -p 3000:3000 -e MONGODB_URL=mongodb://host.docker.internal:27017 game-admin-socket
```

### Production Considerations
- Use environment variables for configuration
- Set up MongoDB with authentication
- Configure reverse proxy (nginx)
- Enable SSL/TLS
- Set up monitoring and logging
- Implement rate limiting
- Configure backup strategies

## 📈 Monitoring

The server includes comprehensive logging for:
- Connection events
- Authentication attempts
- Error occurrences
- Performance metrics
- User activity

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:
- Create an issue in the repository
- Check the documentation
- Review the test examples

---

**Built with ❤️ using Go, Socket.IO, and MongoDB** 