# 🚀 GOSOCKET - Go Fiber Socket.IO Backend

A high-performance, real-time gaming backend built with Go Fiber, Socket.IO, Cassandra, and Redis. This application provides a complete authentication system, real-time communication, and gaming infrastructure.

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)
- [Installation & Setup](#installation--setup)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Socket.IO Events](#socketio-events)
- [Database Schema](#database-schema)
- [Authentication Flow](#authentication-flow)
- [Real-time Features](#real-time-features)
- [Background Services](#background-services)
- [Development](#development)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │   Web Clients   │    │   Mobile Apps   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │    Go Fiber Server        │
                    │  (Port: 8088)             │
                    └─────────────┬─────────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
┌─────────▼─────────┐  ┌─────────▼─────────┐  ┌─────────▼─────────┐
│   Socket.IO       │  │   HTTP Routes     │  │   Background      │
│   Real-time       │  │   REST API        │  │   Services        │
└─────────┬─────────┘  └─────────┬─────────┘  └─────────┬─────────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │    Service Layer          │
                    │  - Auth Service           │
                    │  - Socket Service         │
                    │  - Game Service           │
                    └─────────────┬─────────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
┌─────────▼─────────┐  ┌─────────▼─────────┐  ┌─────────▼─────────┐
│   Cassandra DB    │  │   Redis Cache     │  │   MongoDB         │
│   (Primary DB)    │  │   (Session/OTP)   │  │   (User Data)     │
└───────────────────┘  └───────────────────┘  └───────────────────┘
```

## 🛠️ Technology Stack

### Core Framework
- **Go Fiber v2.52.5** - High-performance HTTP framework
- **Socket.IO v4.0.8** - Real-time bidirectional communication
- **Go 1.22.2** - Programming language

### Database & Cache
- **Apache Cassandra** - Primary database for scalability
- **Redis** - Caching and session management
- **MongoDB** - User data storage

### Authentication & Security
- **JWT (JSON Web Tokens)** - Stateless authentication
- **OTP (One-Time Password)** - Two-factor authentication
- **Session Management** - Secure session handling

### Additional Libraries
- **gocql** - Cassandra driver
- **go-redis** - Redis client
- **mongo-driver** - MongoDB driver
- **godotenv** - Environment configuration

## 📁 Project Structure

```
backendgo/
├── main.go                 # Application entry point
├── go.mod                  # Go module dependencies
├── go.sum                  # Dependency checksums
├── .env.example           # Environment configuration template
├── README.md              # This documentation
│
├── app/                   # Application logic
│   ├── controllers/       # HTTP request handlers
│   │   └── auth_controller.go
│   ├── middlewares/       # HTTP middleware
│   ├── models/           # Data models
│   │   ├── socket_models.go
│   │   └── loginmodel.go
│   ├── routes/           # HTTP route definitions
│   │   └── routes.go
│   ├── services/         # Business logic
│   │   └── socket_service.go
│   └── utils/            # Utility functions
│
├── config/               # Configuration management
│   ├── config.go         # Environment configuration
│   └── socket_handler.go # Socket.IO event handlers
│
├── database/             # Database connections
│   └── database.go       # Cassandra connection setup
│
├── redis/               # Redis cache
│   └── redis_service.go # Redis service implementation
│
└── setup_scripts/       # Database setup scripts
    ├── setup_cassandra.py
    └── DATABASESETUP.py
```

## 🚀 Installation & Setup

### Prerequisites

1. **Go 1.22.2+** installed
2. **Apache Cassandra** running
3. **Redis** server running
4. **MongoDB** (optional, for user data)

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd backendgo
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Setup database**
   ```bash
   python3 setup_cassandra.py
   ```

5. **Run the application**
   ```bash
   go run main.go
   ```

The server will start on port 8088 with comprehensive logging.

## ⚙️ Configuration

### Environment Variables

Create a `.env` file based on `.env.example`:

```env
# Cassandra Configuration
CASSANDRA_HOST=localhost
CASSANDRA_PORT=9042
CASSANDRA_USERNAME=cassandra
CASSANDRA_PASSWORD=cassandra
CASSANDRA_KEYSPACE=myapp

# Redis Configuration
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Server Configuration
SERVER_PORT=8088

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRY=24h

# Application Configuration
APP_ENV=development
APP_DEBUG=true
```

## 📚 API Documentation

### HTTP Endpoints

#### Health Check
```http
GET /health
```
Returns server health status and database connectivity.

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z",
  "services": {
    "cassandra": "ok",
    "redis": "ok"
  }
}
```

#### API Version
```http
GET /api/version
```
Returns application version information.

**Response:**
```json
{
  "version": "1.0.0",
  "name": "GOSOCKET",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Authentication Endpoints

#### Login
```http
POST /auth/login
Content-Type: application/json

{
  "mobile_no": "+1234567890",
  "device_id": "device123",
  "fcm_token": "fcm_token_here"
}
```

#### Verify OTP
```http
POST /auth/verify-otp
Content-Type: application/json

{
  "mobile_no": "+1234567890",
  "session_token": "session_token_here",
  "otp": "123456"
}
```

#### Set Profile
```http
POST /auth/set-profile
Content-Type: application/json

{
  "mobile_no": "+1234567890",
  "session_token": "session_token_here",
  "full_name": "John Doe",
  "state": "California"
}
```

## 🔌 Socket.IO Events

### Connection Events

#### Connect
```javascript
// Client connects to Socket.IO
socket.connect();
```

**Server Response:**
```json
{
  "token": 12345,
  "message": "Connected successfully",
  "timestamp": "2024-01-15T10:30:00Z",
  "socket_id": "socket_123",
  "status": "connected",
  "event": "connect"
}
```

#### Device Info
```javascript
// Send device information
socket.emit('device_info', {
  "device_id": "device123",
  "device_type": "mobile",
  "timestamp": "2024-01-15T10:30:00Z",
  "manufacturer": "Apple",
  "model": "iPhone 14"
});
```

### Authentication Events

#### Login
```javascript
socket.emit('login', {
  "mobile_no": "+1234567890",
  "device_id": "device123",
  "fcm_token": "fcm_token_here"
});
```

**Server Response:**
```json
{
  "status": "success",
  "message": "OTP sent successfully",
  "mobile_no": "+1234567890",
  "device_id": "device123",
  "session_token": "session_token_here",
  "otp": 123456,
  "is_new_user": true,
  "timestamp": "2024-01-15T10:30:00Z",
  "socket_id": "socket_123",
  "event": "login"
}
```

#### Verify OTP
```javascript
socket.emit('verify_otp', {
  "mobile_no": "+1234567890",
  "session_token": "session_token_here",
  "otp": "123456"
});
```

#### Set Profile
```javascript
socket.emit('set_profile', {
  "mobile_no": "+1234567890",
  "session_token": "session_token_here",
  "full_name": "John Doe",
  "state": "California"
});
```

### Gameplay Events

#### Gameplay Namespace
Connect to the gameplay namespace for real-time gaming:
```javascript
const gameplaySocket = io('/gameplay');
```

#### Player Action
```javascript
gameplaySocket.emit('player_action', {
  "action_type": "move",
  "player_id": "player123",
  "session_token": "session_token_here",
  "coordinates": {
    "x": 100,
    "y": 200
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "game_state": {
    "level": 1,
    "score": 1000,
    "health": 100
  }
});
```

### Utility Events

#### Heartbeat
```javascript
socket.emit('heartbeat');
```

#### Disconnect
```javascript
socket.disconnect();
```

## 🗄️ Database Schema

### Cassandra Tables

#### Users Table
```sql
CREATE TABLE users (
    mobile_no text PRIMARY KEY,
    email text,
    full_name text,
    state text,
    referral_code text,
    referred_by text,
    language_code text,
    language_name text,
    status text,
    created_at timestamp,
    updated_at timestamp
);
```

#### Sessions Table
```sql
CREATE TABLE sessions (
    session_token text PRIMARY KEY,
    user_id text,
    mobile_no text,
    device_id text,
    fcm_token text,
    jwt_token text,
    created_at timestamp,
    expires_at timestamp,
    is_active boolean
);
```

#### OTP Table
```sql
CREATE TABLE otps (
    phone_or_email text,
    otp_code text,
    created_at timestamp,
    expires_at timestamp,
    purpose text,
    is_verified boolean,
    attempt_count int,
    PRIMARY KEY (phone_or_email, created_at)
);
```

### Redis Keys

- `session:{session_token}` - Session data
- `otp:{mobile_no}` - OTP data
- `user:{mobile_no}` - User cache
- `socket:{socket_id}` - Socket connection data

## 🔐 Authentication Flow

### 1. User Registration/Login
```
Client → Socket.IO Login Event → Server
Server → Generate OTP → Store in Cassandra
Server → Send OTP Response → Client
```

### 2. OTP Verification
```
Client → Socket.IO Verify OTP Event → Server
Server → Validate OTP → Generate JWT
Server → Create Session → Store in Cassandra
Server → Send JWT Response → Client
```

### 3. Profile Setup
```
Client → Socket.IO Set Profile Event → Server
Server → Update User Profile → Store in Cassandra
Server → Send Profile Response → Client
```

### 4. Session Management
```
Client → Include JWT in Requests → Server
Server → Validate JWT → Check Session
Server → Process Request → Send Response
```

## ⚡ Real-time Features

### Socket.IO Namespaces

1. **Default Namespace (`/`)**
   - Authentication events
   - Device management
   - General communication

2. **Gameplay Namespace (`/gameplay`)**
   - Real-time gaming events
   - Player actions
   - Game state synchronization

### Real-time Capabilities

- **Bidirectional Communication** - Instant message exchange
- **Room Management** - Group users in game rooms
- **Event Broadcasting** - Send events to multiple clients
- **Connection Management** - Handle disconnections gracefully
- **Heartbeat Monitoring** - Keep connections alive

## 🧹 Background Services

### Cleanup Service
The application runs a background service every 5 minutes to:

1. **Cleanup Expired Sessions**
   - Remove sessions past expiration time
   - Free up database space

2. **Cleanup Expired OTPs**
   - Remove OTPs past expiration time
   - Maintain security

### Service Configuration
```go
// Runs every 5 minutes
ticker := time.NewTicker(5 * time.Minute)

// Cleanup operations
socketService.CleanupExpiredSessions()
socketService.CleanupExpiredOTPs()
```

## 🛠️ Development

### Running in Development

1. **Start with debug logging**
   ```bash
   go run main.go
   ```

2. **Monitor logs**
   - All operations are logged with emojis
   - Debug information for troubleshooting
   - Error tracking and reporting

### Code Structure Best Practices

- **Separation of Concerns** - Clear module boundaries
- **Error Handling** - Comprehensive error management
- **Logging** - Detailed operation logging
- **Configuration** - Environment-based configuration
- **Testing** - Unit and integration tests

### Adding New Features

1. **Models** - Define data structures in `app/models/`
2. **Services** - Implement business logic in `app/services/`
3. **Controllers** - Handle HTTP requests in `app/controllers/`
4. **Routes** - Define endpoints in `app/routes/`
5. **Socket Events** - Add real-time events in `config/socket_handler.go`

## 🚀 Deployment

### Production Setup

1. **Environment Configuration**
   ```bash
   # Set production environment
   APP_ENV=production
   APP_DEBUG=false
   ```

2. **Database Setup**
   ```bash
   # Ensure Cassandra is running
   # Setup production keyspace
   # Configure Redis cluster
   ```

3. **Build Application**
   ```bash
   go build -o gosocket main.go
   ```

4. **Run with Process Manager**
   ```bash
   # Using systemd or PM2
   ./gosocket
   ```

### Docker Deployment

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o gosocket main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/gosocket .
CMD ["./gosocket"]
```

## 🔧 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check Cassandra service status
   - Verify connection credentials
   - Ensure keyspace exists

2. **Redis Connection Failed**
   - Check Redis service status
   - Verify Redis URL and credentials
   - Check network connectivity

3. **Socket.IO Connection Issues**
   - Verify client Socket.IO version
   - Check CORS configuration
   - Monitor server logs

### Debug Mode

Enable debug logging by setting:
```env
APP_DEBUG=true
LOG_LEVEL=debug
```

### Health Check

Use the health endpoint to verify service status:
```bash
curl http://localhost:8088/health
```

### Log Analysis

Look for these log patterns:
- `✅` - Successful operations
- `❌` - Errors and failures
- `⚠️` - Warnings
- `🔌` - Connection events
- `🧹` - Cleanup operations

## 📞 Support

For issues and questions:
1. Check the logs for detailed error messages
2. Verify configuration settings
3. Test database connectivity
4. Review Socket.IO client implementation

---

**GOSOCKET v1.0.0** - Built with ❤️ using Go Fiber and Socket.IO 