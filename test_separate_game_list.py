#!/usr/bin/env python3
"""
Test script for separate game list handler functionality
This script demonstrates the new HandleGameListRequest function
"""

import socketio
import json
import time
import asyncio
from datetime import datetime

# Create Socket.IO client
sio = socketio.AsyncClient()

# Test configuration
SERVER_URL = "http://localhost:8088"

@sio.event
async def connect():
    print("🔗 Connected to server")
    print(f"📱 Socket ID: {sio.sid}")
    
    # Start the test flow
    await test_separate_game_list_flow()

@sio.event
async def disconnect():
    print("❌ Disconnected from server")

@sio.event
async def game_list_response(data):
    """Handle game list response from separate handler"""
    print("\n🎮 SEPARATE GAME LIST RESPONSE:")
    print("=" * 50)
    
    # Print basic response info
    print(f"Status: {data.get('status')}")
    print(f"Message: {data.get('message')}")
    print(f"Event: {data.get('event')}")
    print(f"Message Type: {data.get('message_type')}")
    print(f"Timestamp: {data.get('timestamp')}")
    
    # Print data structure
    if 'data' in data:
        data_content = data['data']
        print(f"\n📊 DATA STRUCTURE:")
        
        # Print metadata if available
        if 'metadata' in data_content:
            metadata = data_content['metadata']
            print(f"  Total Contests: {metadata.get('total_contests')}")
            print(f"  Processed At: {metadata.get('processed_at')}")
            print(f"  Version: {metadata.get('version')}")
        
        # Print gamelist info
        if 'gamelist' in data_content:
            gamelist = data_content['gamelist']
            print(f"\n🎯 GAMELIST ({len(gamelist)} contests):")
            
            for i, contest in enumerate(gamelist[:3], 1):  # Show first 3 contests
                print(f"\n  Contest {i}:")
                print(f"    Contest ID: {contest.get('contestId')}")
                print(f"    Name: {contest.get('contestName')}")
                print(f"    Status: {contest.get('status')} (Priority: {contest.get('status_priority')})")
                print(f"    Difficulty: {contest.get('difficulty')} (Score: {contest.get('difficulty_score')})")
                print(f"    Participants: {contest.get('currentParticipants')}/{contest.get('maxParticipants')}")
                print(f"    Participation: {contest.get('participation_percentage', 0)}%")
                print(f"    Spots Remaining: {contest.get('spots_remaining', 0)}")
                print(f"    Time Until Start: {contest.get('time_until_start', 'N/A')}")
                print(f"    Reward Category: {contest.get('reward_category', 'N/A')}")
                print(f"    Has Popular Category: {contest.get('has_popular_category', False)}")
                print(f"    Processed: {contest.get('processed', False)}")
    
    print("\n" + "=" * 50)

@sio.event
async def error_response(data):
    """Handle error responses"""
    print(f"\n❌ ERROR RESPONSE:")
    print(f"Error: {data.get('message')}")
    print(f"Code: {data.get('error_code')}")
    print(f"Type: {data.get('error_type')}")

async def test_separate_game_list_flow():
    """Test the separate game list handler"""
    print("\n🚀 STARTING SEPARATE GAME LIST TEST")
    print("=" * 50)
    
    try:
        # Step 1: Send separate game list request
        print("\n1️⃣ Sending separate game list request...")
        await sio.emit('get_game_list_separate', {})
        await asyncio.sleep(3)
        
        # Step 2: Send another request to test Redis caching
        print("\n2️⃣ Sending second request (testing Redis cache)...")
        await sio.emit('get_game_list_separate', {})
        await asyncio.sleep(3)
        
        print("\n✅ Separate game list test completed successfully!")
        
    except Exception as e:
        print(f"\n❌ Error during test: {e}")
    
    finally:
        # Disconnect after a delay
        await asyncio.sleep(2)
        await sio.disconnect()

async def main():
    """Main function to run the test"""
    print("🎮 Separate Game List Handler Test")
    print("=" * 50)
    print(f"Server URL: {SERVER_URL}")
    print("=" * 50)
    
    try:
        await sio.connect(SERVER_URL)
        await sio.wait()
    except Exception as e:
        print(f"❌ Connection failed: {e}")

if __name__ == "__main__":
    asyncio.run(main()) 