# GOSOCKET - Real-time Gaming Platform

A high-performance real-time gaming platform built with Go, Socket.IO, and Apache Cassandra for scalable multiplayer gaming experiences.

## 🚀 Features

- **Real-time Communication**: Socket.IO for bidirectional communication
- **User Authentication**: Mobile-based login with OTP verification
- **Session Management**: JWT-based session handling with device tracking
- **Gaming Platform**: Contest management, game listings, and real-time gameplay
- **Scalable Database**: Apache Cassandra for high-performance data storage
- **Caching Layer**: Redis for performance optimization
- **Push Notifications**: FCM integration for real-time notifications

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │   Web Clients   │    │   Admin Panel   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │    GOSOCKET Server        │
                    │  (Go + Fiber + Socket.IO) │
                    └─────────────┬─────────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
┌─────────▼─────────┐  ┌─────────▼─────────┐  ┌─────────▼─────────┐
│   Apache Cassandra │  │      Redis        │  │   FCM Service     │
│   (Primary DB)     │  │   (Caching)       │  │ (Notifications)   │
└────────────────────┘  └────────────────────┘  └────────────────────┘
```

## 📋 Prerequisites

- **Go 1.22+** - [Download](https://golang.org/dl/)
- **Apache Cassandra 4.x** - [Download](https://cassandra.apache.org/download/)
- **Redis 6.x+** - [Download](https://redis.io/download)
- **Python 3.8+** (for database setup)
- **Git**

## 🛠️ Installation & Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd GOSOCKEKT
```

### 2. Install Go Dependencies

```bash
go mod download
```

### 3. Install Python Dependencies (for database setup)

```bash
pip install cassandra-driver
```

### 4. Configure Environment

Copy the environment template and configure your settings:

```bash
cp env.example .env
```

Edit `.env` file with your configuration:

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

# Application Configuration
APP_ENV=development
APP_DEBUG=true

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRY=24h
```

### 5. Start Cassandra

#### Using Docker (Recommended)

```bash
docker run -d --name cassandra \
  -p 9042:9042 \
  -e CASSANDRA_USER=cassandra \
  -e CASSANDRA_PASSWORD=cassandra \
  cassandra:4.1
```

#### Using Local Installation

1. Download and install Cassandra
2. Start the service:
   ```bash
   sudo systemctl start cassandra
   # or
   cassandra
   ```

### 6. Start Redis

#### Using Docker

```bash
docker run -d --name redis \
  -p 6379:6379 \
  redis:7-alpine
```

#### Using Local Installation

```bash
redis-server
```

### 7. Setup Database

Run the database setup script:

```bash
python setup_cassandra.py
```

This will:
- Create the keyspace
- Create all required tables
- Set up secondary indexes
- Insert sample data
- Verify the setup

### 8. Run the Application

```bash
go run main.go
```

The server will start on port 8088 (or your configured port).

## 📊 Database Schema

### Tables

1. **users** - User profiles and preferences
2. **sessions** - Active user sessions with composite primary key
3. **games** - Game catalog and metadata
4. **contests** - Contest information and pricing
5. **server_announcements** - System announcements
6. **game_updates** - Game version updates
7. **sessions_by_token** - Quick session lookups by token

### Key Design Patterns

- **Composite Primary Keys** for efficient querying
- **Secondary Indexes** for flexible queries
- **TTL Support** for session expiration
- **Clustering Keys** for ordered data access

## 🔌 API Endpoints

### Socket.IO Events

| Event | Description | Payload |
|-------|-------------|---------|
| `device:info` | Register device | `DeviceInfo` |
| `login` | User authentication | `LoginRequest` |
| `verify:otp` | OTP verification | `OTPVerificationRequest` |
| `set:profile` | Profile setup | `SetProfileRequest` |
| `set:language` | Language preferences | `SetLanguageRequest` |
| `main:screen` | Main dashboard | `MainScreenRequest` |
| `contest:list` | Contest listings | `ContestRequest` |
| `contest:join` | Join contest | `ContestJoinRequest` |
| `contest:gap` | Price analysis | `ContestGapRequest` |

### HTTP Endpoints

Currently minimal - focus is on Socket.IO for real-time features.

## 🎮 Usage Examples

### Connect to Socket.IO

```javascript
const socket = io('http://localhost:8088');

// Connect to gameplay namespace
const gameplaySocket = io('http://localhost:8088/gameplay');

// Send device info
socket.emit('device:info', {
  device_id: 'device_123',
  device_type: 'mobile',
  manufacturer: 'Samsung',
  model: 'Galaxy S21'
});

// Login
socket.emit('login', {
  mobile_no: '+1234567890',
  device_id: 'device_123',
  fcm_token: 'your-fcm-token-here'
});
```

### Go Client Example

```go
package main

import (
    "github.com/gorilla/websocket"
    "log"
)

func main() {
    // Connect to Socket.IO server
    conn, _, err := websocket.Dial("ws://localhost:8088/socket.io/", nil)
    if err != nil {
        log.Fatal("dial:", err)
    }
    defer conn.Close()

    // Send login request
    loginReq := `{
        "event": "login",
        "data": {
            "mobile_no": "+1234567890",
            "device_id": "device_123",
            "fcm_token": "your-fcm-token"
        }
    }`
    
    err = conn.WriteMessage(websocket.TextMessage, []byte(loginReq))
    if err != nil {
        log.Fatal("write:", err)
    }
}
```

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CASSANDRA_HOST` | `localhost` | Cassandra server host |
| `CASSANDRA_PORT` | `9042` | Cassandra server port |
| `CASSANDRA_USERNAME` | `cassandra` | Cassandra username |
| `CASSANDRA_PASSWORD` | `cassandra` | Cassandra password |
| `CASSANDRA_KEYSPACE` | `myapp` | Cassandra keyspace |
| `REDIS_URL` | `localhost:6379` | Redis server URL |
| `REDIS_PASSWORD` | `` | Redis password |
| `REDIS_DB` | `0` | Redis database number |
| `SERVER_PORT` | `8088` | Server port |
| `JWT_SECRET` | - | JWT secret key |
| `JWT_EXPIRY` | `24h` | JWT expiration time |

### Cassandra Configuration

The application is optimized for Cassandra with:

- **Connection Pooling**: 10 connections per node
- **Retry Policy**: Simple retry with 3 attempts
- **Consistency Level**: QUORUM for writes
- **Timeout Settings**: 10s connection, 10s query timeout

## 🚀 Deployment

### Docker Deployment

```bash
# Build the application
docker build -t gosocket .

# Run with environment variables
docker run -d \
  --name gosocket \
  -p 8088:8088 \
  -e CASSANDRA_HOST=cassandra \
  -e REDIS_URL=redis:6379 \
  gosocket
```

### Production Considerations

1. **Cassandra Cluster**: Use multiple nodes for high availability
2. **Redis Cluster**: Use Redis Cluster for scalability
3. **Load Balancer**: Use nginx or HAProxy for load balancing
4. **Monitoring**: Implement health checks and metrics
5. **Security**: Use proper authentication and SSL/TLS

## 🧪 Testing

### Health Check

```bash
curl http://localhost:8088/health
```

### Socket.IO Connection Test

```bash
# Using wscat
npm install -g wscat
wscat -c ws://localhost:8088/socket.io/
```

## 📝 Logging

The application uses structured logging with different levels:

- **INFO**: General application flow
- **WARNING**: Non-critical issues
- **ERROR**: Critical errors
- **DEBUG**: Detailed debugging information

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
- Review the code examples

## 🔄 Changelog

### v1.0.0
- Initial release
- Socket.IO integration
- Cassandra database support
- User authentication system
- Contest management
- Real-time gaming features

---

**Happy Gaming! 🎮** 