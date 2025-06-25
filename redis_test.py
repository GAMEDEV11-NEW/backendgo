#!/usr/bin/env python3
"""
Redis Connection Test Script
Tests Redis connectivity and basic operations on 127.0.0.1:6379
"""

import redis
import json
import time
from datetime import datetime

def test_redis_connection():
    """Test basic Redis connection and operations"""
    print("🔍 Testing Redis Connection...")
    print("=" * 50)
    
    try:
        # Create Redis client
        r = redis.Redis(
            host='127.0.0.1',
            port=6379,
            db=0,
            decode_responses=True,
            socket_connect_timeout=5,
            socket_timeout=5
        )
        
        # Test 1: Ping Redis
        print("1️⃣ Testing PING...")
        response = r.ping()
        print(f"   ✅ PING Response: {response}")
        
        # Test 2: Set and Get a simple key
        print("\n2️⃣ Testing SET/GET operations...")
        test_key = "test:connection"
        test_value = f"Hello Redis! Time: {datetime.now().isoformat()}"
        
        r.set(test_key, test_value)
        retrieved_value = r.get(test_key)
        print(f"   ✅ SET: {test_key} = {test_value}")
        print(f"   ✅ GET: {test_key} = {retrieved_value}")
        
        # Test 3: Test JSON operations (like your Go app)
        print("\n3️⃣ Testing JSON operations...")
        json_key = "test:json"
        json_data = {
            "user_id": "12345",
            "mobile_no": "9876543210",
            "status": "active",
            "timestamp": datetime.now().isoformat(),
            "data": {
                "game_list": [
                    {"game_id": 1, "name": "Game 1", "active_players": 100},
                    {"game_id": 2, "name": "Game 2", "active_players": 200}
                ]
            }
        }
        
        # Store JSON data
        r.set(json_key, json.dumps(json_data))
        print(f"   ✅ JSON SET: {json_key}")
        
        # Retrieve and parse JSON data
        retrieved_json = r.get(json_key)
        parsed_data = json.loads(retrieved_json)
        print(f"   ✅ JSON GET: {json_key}")
        print(f"   📊 Parsed data keys: {list(parsed_data.keys())}")
        
        # Test 4: Test expiration (TTL)
        print("\n4️⃣ Testing TTL (Time To Live)...")
        ttl_key = "test:ttl"
        r.setex(ttl_key, 10, "This will expire in 10 seconds")
        ttl = r.ttl(ttl_key)
        print(f"   ✅ TTL SET: {ttl_key} (expires in {ttl} seconds)")
        
        # Test 5: Test exists and delete
        print("\n5️⃣ Testing EXISTS and DELETE...")
        exists_before = r.exists(test_key)
        print(f"   ✅ EXISTS before delete: {test_key} = {exists_before}")
        
        r.delete(test_key)
        exists_after = r.exists(test_key)
        print(f"   ✅ EXISTS after delete: {test_key} = {exists_after}")
        
        # Test 6: Test Redis info
        print("\n6️⃣ Testing Redis INFO...")
        info = r.info()
        print(f"   ✅ Redis Version: {info.get('redis_version', 'Unknown')}")
        print(f"   ✅ Connected Clients: {info.get('connected_clients', 'Unknown')}")
        print(f"   ✅ Used Memory: {info.get('used_memory_human', 'Unknown')}")
        
        # Test 7: Test game list caching (like your Go app)
        print("\n7️⃣ Testing Game List Caching...")
        game_list_key = "gamelist:current"
        game_list_data = {
            "gamelist": [
                {
                    "active_gamepalye": 12313,
                    "livegameplaye": 12313,
                    "game name": "newgame"
                },
                {
                    "active_gamepalye": 45678,
                    "livegameplaye": 45678,
                    "game name": "testgame"
                },
                {
                    "active_gamepalye": 45678,
                    "livegameplaye": 45678,
                    "game name": "testgame"
                },
                {
                    "active_gamepalye": 45678,
                    "livegameplaye": 45678,
                    "game name": "testgame"
                },
                {
                    "active_gamepalye": 45678,
                    "livegameplaye": 45678,
                    "game name": "testgame"
                }
            ],
            "cached_at": datetime.now().isoformat(),
            "expires_in": "5 minutes"
        }
        
        # Cache game list for 5 minutes (300 seconds)
        r.setex(game_list_key, 300, json.dumps(game_list_data))
        print(f"   ✅ Game list cached: {game_list_key}")
        
        # Retrieve cached game list
        cached_game_list = r.get(game_list_key)
        if cached_game_list:
            parsed_game_list = json.loads(cached_game_list)
            print(f"   ✅ Cached game list retrieved: {len(parsed_game_list.get('gamelist', []))} games")
        
        # Test 8: Test session caching
        print("\n8️⃣ Testing Session Caching...")
        session_id = "session:test123"
        session_data = {
            "user_id": "user123",
            "mobile_no": "9876543210",
            "device_id": "device123",
            "fcm_token": "fcm_token_123",
            "created_at": datetime.now().isoformat(),
            "expires_at": (datetime.now().timestamp() + 86400)  # 24 hours
        }
        
        r.setex(session_id, 86400, json.dumps(session_data))
        print(f"   ✅ Session cached: {session_id}")
        
        # Retrieve session
        cached_session = r.get(session_id)
        if cached_session:
            parsed_session = json.loads(cached_session)
            print(f"   ✅ Session retrieved: {parsed_session.get('mobile_no')}")
        
        print("\n" + "=" * 50)
        print("🎉 All Redis tests completed successfully!")
        print("✅ Redis is working properly on 127.0.0.1:6379")
        
        return True
        
    except redis.ConnectionError as e:
        print(f"❌ Redis Connection Error: {e}")
        print("💡 Make sure Redis server is running on 127.0.0.1:6379")
        return False
        
    except Exception as e:
        print(f"❌ Unexpected Error: {e}")
        return False

def test_redis_performance():
    """Test Redis performance with multiple operations"""
    print("\n🚀 Testing Redis Performance...")
    print("=" * 50)
    
    try:
        r = redis.Redis(host='127.0.0.1', port=6379, db=0, decode_responses=True)
        
        # Test bulk operations
        start_time = time.time()
        
        # Set multiple keys
        for i in range(100):
            r.set(f"perf:key:{i}", f"value:{i}")
        
        # Get multiple keys
        for i in range(100):
            r.get(f"perf:key:{i}")
        
        end_time = time.time()
        duration = end_time - start_time
        
        print(f"✅ 200 operations (100 SET + 100 GET) completed in {duration:.3f} seconds")
        print(f"✅ Average: {200/duration:.1f} operations/second")
        
        # Cleanup
        for i in range(100):
            r.delete(f"perf:key:{i}")
        
        return True
        
    except Exception as e:
        print(f"❌ Performance test failed: {e}")
        return False

if __name__ == "__main__":
    print("🔧 Redis Test Script")
    print("Testing Redis on 127.0.0.1:6379")
    print("=" * 50)
    
    # Test basic functionality
    basic_test = test_redis_connection()
    
    if basic_test:
        # Test performance
        test_redis_performance()
    
    print("\n📋 Test Summary:")
    if basic_test:
        print("✅ Redis is working correctly!")
        print("✅ Your Go application should be able to use Redis for caching")
    else:
        print("❌ Redis is not working properly")
        print("💡 Please check if Redis server is running")
        print("💡 Try: redis-server (to start Redis)")
        print("💡 Or install Redis if not installed") 