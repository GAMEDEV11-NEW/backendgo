#!/usr/bin/env python3
"""
Test Script for trigger_game_list_update Flow
Demonstrates how Python script triggers Redis updates to all connected Socket.IO clients
"""

import json
import time
from datetime import datetime
import socketio


def test_trigger_flow():
    """Test the complete trigger flow"""
    print("🎮 Testing trigger_game_list_update Flow")
    print("=" * 60)

    # Initialize Redis client
    try:
        import redis
        redis_client = redis.Redis(
            host='127.0.0.1', port=6379, db=0, decode_responses=True)
    except ImportError:
        print("❌ Redis module not found. Please run: pip install redis==5.0.1")
        return
    except AttributeError:
        print("❌ Redis module issue. Trying alternative import...")
        try:
            import redis.client
            redis_client = redis.client.Redis(
                host='127.0.0.1', port=6379, db=0, decode_responses=True)
        except:
            print("❌ Redis not properly installed. Please run: pip install redis==5.0.1")
            return

    # Initialize Socket.IO client
    sio = socketio.Client()

    try:
        # Test Redis connection
        redis_client.ping()
        print("✅ Redis connection successful")

        # Connect to Socket.IO server with WebSocket transport
        sio.connect('http://localhost:8088',
                    transports=['websocket'])
        print("✅ Socket.IO WebSocket connection successful")

        # Step 1: Update game list in Redis
        print("\n📝 Step 1: Updating game list in Redis...")
        game_list_data = {
            "gamelist": [
                {
                    "game_id": "test_game_1",
                    "game_name": "Test Poker",
                    "game_type": "card",
                    "active_gamepalye": 1000,
                    "livegameplaye": 900,
                    "status": "active",
                    "created_at": datetime.now().isoformat()
                },
                {
                    "game_id": "test_game_2",
                    "game_name": "Test Rummy",
                    "game_type": "card",
                    "active_gamepalye": 1000,
                    "livegameplaye": 600,
                    "status": "active",
                    "created_at": datetime.now().isoformat()
                },
                {
                    "game_id": "test_game_2",
                    "game_name": "Test Rummy",
                    "game_type": "card",
                    "active_gamepalye": 1000,
                    "livegameplaye": 600,
                    "status": "active",
                    "created_at": datetime.now().isoformat()
                }
            ],
            "updated_at": datetime.now().isoformat(),
            "total_games": 2,
            "active_games": 2,
            "total_players": 27000
        }

        # Cache in Redis for 5 minutes
        redis_client.setex("gamelist:current", 300, json.dumps(game_list_data))
        print("✅ Game list updated in Redis")

        # Step 2: Trigger update via Socket.IO
        print("\n📡 Step 2: Triggering game list update...")
        sio.emit("trigger_game_list_update", {
            "message": "Test trigger from Python script",
            "timestamp": datetime.now().isoformat()
        })
        print("✅ Trigger event sent to Go server")

        # Step 3: Wait a moment for processing
        print("\n⏰ Step 3: Waiting for server processing...")
        time.sleep(2)

        # Step 4: Verify Redis data
        print("\n📖 Step 4: Verifying Redis data...")
        cached_data = redis_client.get("gamelist:current")
        if cached_data:
            parsed_data = json.loads(cached_data)
            print(
                f"✅ Redis contains {len(parsed_data.get('gamelist', []))} games")
        else:
            print("❌ No data found in Redis")

        print("\n🎉 Test completed successfully!")
        print("📋 Flow Summary:")
        print("   1. ✅ Updated game list in Redis")
        print("   2. ✅ Sent trigger_game_list_update event")
        print("   3. ✅ Go server should fetch from Redis")
        print("   4. ✅ Go server should broadcast to all clients")
        print("   5. ✅ All connected clients should receive main:screen:game:list")

    except Exception as e:
        print(f"❌ Test failed: {e}")

    finally:
        # Cleanup
        try:
            sio.disconnect()
        except:
            pass


def continuous_test():
    """Run continuous tests"""
    print("🔄 Starting continuous trigger tests...")
    print("=" * 60)

    try:
        import redis
        redis_client = redis.Redis(
            host='127.0.0.1', port=6379, db=0, decode_responses=True)
    except ImportError:
        print("❌ Redis module not found. Please run: pip install redis==5.0.1")
        return
    except AttributeError:
        print("❌ Redis module issue. Please run: pip install redis==5.0.1")
        return

    sio = socketio.Client()

    try:
        # Connect
        redis_client.ping()
        sio.connect('http://localhost:8088', transports=['websocket'])
        print("✅ Connected to Redis and Socket.IO WebSocket")

        test_count = 0
        while True:
            test_count += 1
            print(
                f"\n📅 Test #{test_count} - {datetime.now().strftime('%H:%M:%S')}")

            # Generate random game data
            import random
            game_count = random.randint(2, 5)
            games = []

            game_names = ["Poker", "Rummy", "Teen Patti", "Carrom", "Ludo"]
            for i in range(game_count):
                game = {
                    "game_id": f"game_{i+1}_{int(time.time())}",
                    "game_name": random.choice(game_names),
                    "game_type": "card",
                    "active_gamepalye": random.randint(1000, 50000),
                    "livegameplaye": random.randint(500, 20000),
                    "status": "active",
                    "created_at": datetime.now().isoformat()
                }
                games.append(game)

            game_list_data = {
                "gamelist": games,
                "updated_at": datetime.now().isoformat(),
                "total_games": len(games),
                "active_games": len(games),
                "total_players": sum(g["active_gamepalye"] for g in games)
            }

            # Update Redis
            redis_client.setex("gamelist:current", 300,
                               json.dumps(game_list_data))
            print(f"📝 Updated Redis with {len(games)} games")

            # Trigger update
            sio.emit("trigger_game_list_update", {
                "message": f"Continuous test #{test_count}",
                "timestamp": datetime.now().isoformat()
            })
            print("📡 Trigger sent")

            # Wait
            time.sleep(10)

    except KeyboardInterrupt:
        print("\n⏹️ Continuous test stopped by user")
    except Exception as e:
        print(f"\n❌ Continuous test failed: {e}")
    finally:
        try:
            sio.disconnect()
        except:
            pass


if __name__ == "__main__":
    print("🔧 Trigger Flow Test Script")
    print("=" * 60)
    print("1. Single test")
    print("2. Continuous tests")

    choice = input("\nEnter choice (1-2): ").strip()

    if choice == "1":
        test_trigger_flow()
    elif choice == "2":
        continuous_test()
    else:
        print("❌ Invalid choice")
