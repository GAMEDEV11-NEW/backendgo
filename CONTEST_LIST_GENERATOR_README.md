# Contest List Generator for Redis

This Python script generates contest list data and stores it in Redis under the key `listcontest:current`, which is used by the Go Socket.IO service.

## Features

- 🏆 Generates realistic contest data with various types (upcoming, active, completed)
- 💾 Stores data in Redis with configurable expiration time
- 🔄 Retrieves and verifies data from Redis
- 📊 Displays formatted contest information
- 🧪 Includes comprehensive testing functionality

## Prerequisites

1. **Redis Server**: Make sure Redis is running on `localhost:6379`
2. **Python 3.7+**: Required for running the script
3. **Python Dependencies**: Install required packages

## Installation

1. **Install Python dependencies**:
   ```bash
   pip install -r requirements_contest.txt
   ```

   Or install manually:
   ```bash
   pip install redis==5.0.1
   ```

2. **Verify Redis is running**:
   ```bash
   redis-cli ping
   ```
   Should return `PONG`

## Usage

### 1. Generate Contest List Data

Run the main script to create contest list data:

```bash
python create_contest_list.py
```

This will:
- Connect to Redis
- Generate 6 sample contests with realistic data
- Store the data under key `listcontest:current`
- Set expiration to 5 minutes
- Display the generated contests

### 2. Test the Implementation

Run the test script to verify everything works:

```bash
python test_contest_list.py
```

This will:
- Test Redis connection
- Generate and store contest data
- Retrieve and verify the data
- Check Redis key structure
- Display test results

### 3. Manual Redis Verification

You can also manually check the data in Redis:

```bash
# Connect to Redis CLI
redis-cli

# Check if key exists
EXISTS listcontest:current

# Get the data
GET listcontest:current

# Check TTL (time to live)
TTL listcontest:current
```

## Data Structure

The contest list data is stored in the following format:

```json
{
  "gamelist": [
    {
      "contestId": 1,
      "contestName": "Weekly Algorithm Challenge",
      "description": "Solve complex algorithmic problems...",
      "startTime": "2025-01-15T09:00:00Z",
      "endTime": "2025-01-22T23:59:59Z",
      "status": "upcoming",
      "maxParticipants": 1000,
      "currentParticipants": 847,
      "categories": ["algorithms", "data-structures"],
      "difficulty": "medium",
      "duration": "7 days",
      "reward": {
        "currency": "USD",
        "amount": 5000,
        "distribution": {
          "1st": 2500,
          "2nd": 1500,
          "3rd": 1000
        }
      },
      "rules": ["Individual participation only", ...],
      "tags": ["competitive", "weekly"],
      "sponsor": "TechCorp Inc.",
      "registrationDeadline": "2025-01-14T23:59:59Z"
    }
  ]
}
```

## Contest Types

The script generates contests with different statuses:

1. **Upcoming Contests**: Future contests that users can register for
2. **Active Contests**: Currently running contests with live stats
3. **Completed Contests**: Finished contests with results

## Configuration

You can modify the script to change:

- **Redis Connection**: Update host, port, and database in the constructor
- **Expiration Time**: Change the default 5-minute expiration
- **Contest Data**: Modify the contest generation logic
- **Number of Contests**: Add or remove contests from the list

## Integration with Go Service

The Go Socket.IO service expects this data structure:

```go
// In getGameSubListData() function
gameListData := map[string]interface{}{
    "gamelist": gamelist,
}
```

The Python script creates exactly this structure, so the Go service can retrieve it using:

```go
// In Redis service
func (r *Service) GetListContest() (map[string]interface{}, error) {
    key := "listcontest:current"
    var listContest map[string]interface{}
    err := r.Get(key, &listContest)
    if err != nil {
        return nil, err
    }
    return listContest, nil
}
```

## Error Handling

The script includes comprehensive error handling:

- **Redis Connection Errors**: Graceful handling of connection failures
- **Data Validation**: Verifies data structure before storing
- **JSON Serialization**: Handles encoding/decoding errors
- **Missing Data**: Fallback behavior for missing keys

## Logging

The script uses Python's logging module with informative messages:

- ✅ Success operations
- ❌ Error conditions
- ⚠️ Warnings
- 📊 Data statistics
- 🔑 Redis key information

## Troubleshooting

### Common Issues

1. **Redis Connection Failed**:
   - Ensure Redis is running: `redis-server`
   - Check Redis port: `redis-cli -p 6379 ping`

2. **Import Errors**:
   - Install dependencies: `pip install redis`
   - Check Python version: `python --version`

3. **Data Not Found**:
   - Check if key exists: `redis-cli EXISTS listcontest:current`
   - Verify TTL: `redis-cli TTL listcontest:current`

### Debug Mode

To enable debug logging, modify the script:

```python
logging.basicConfig(
    level=logging.DEBUG,  # Change from INFO to DEBUG
    format='%(asctime)s - %(levelname)s - %(message)s'
)
```

## Examples

### Basic Usage

```bash
# Generate contest data
python create_contest_list.py

# Test the implementation
python test_contest_list.py
```

### Custom Configuration

```python
# Create generator with custom Redis settings
generator = ContestListGenerator(
    redis_host='192.168.1.100',
    redis_port=6380,
    redis_db=1
)

# Store with custom expiration
generator.store_contest_list_in_redis(data, expiration_minutes=10)
```

## Contributing

To add new contest types or modify the data structure:

1. Update the `generate_contest_list_data()` method
2. Ensure the data structure matches the Go service expectations
3. Test with the provided test script
4. Update this documentation if needed

## License

This script is part of the Socket.IO Game Admin project and follows the same licensing terms. 