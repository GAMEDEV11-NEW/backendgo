#!/usr/bin/env python3
"""
Test script to verify contest list data in Redis
"""

import redis
import json
from create_contest_list import ContestListGenerator

def test_contest_list():
    """Test contest list functionality"""
    print("🧪 Testing Contest List Functionality")
    print("=" * 50)
    
    # Create generator instance
    generator = ContestListGenerator()
    
    # Test Redis connection
    print("1. Testing Redis connection...")
    if not generator.test_redis_connection():
        print("❌ Redis connection failed")
        return False
    print("✅ Redis connection successful")
    
    # Generate and store data
    print("\n2. Generating and storing contest list data...")
    contest_data = generator.generate_contest_list_data()
    
    if generator.store_contest_list_in_redis(contest_data):
        print("✅ Contest list data stored successfully")
    else:
        print("❌ Failed to store contest list data")
        return False
    
    # Retrieve and verify data
    print("\n3. Retrieving and verifying contest list data...")
    retrieved_data = generator.retrieve_contest_list_from_redis()
    
    if retrieved_data:
        print("✅ Contest list data retrieved successfully")
        print(f"📊 Number of contests: {len(retrieved_data['gamelist'])}")
        
        # Display first contest for verification
        first_contest = retrieved_data['gamelist'][0]
        print(f"\n📋 First contest details:")
        print(f"   Name: {first_contest['contestName']}")
        print(f"   ID: {first_contest['contestId']}")
        print(f"   Status: {first_contest['status']}")
        print(f"   Difficulty: {first_contest['difficulty']}")
        
        return True
    else:
        print("❌ Failed to retrieve contest list data")
        return False

def test_redis_key_structure():
    """Test the Redis key structure"""
    print("\n4. Testing Redis key structure...")
    
    try:
        r = redis.Redis(host='localhost', port=6379, db=0, decode_responses=True)
        
        # Check if key exists
        key = "listcontest:current"
        exists = r.exists(key)
        print(f"🔑 Key '{key}' exists: {exists}")
        
        if exists:
            # Get TTL
            ttl = r.ttl(key)
            print(f"⏰ TTL: {ttl} seconds")
            
            # Get data size
            data = r.get(key)
            if data:
                data_size = len(data)
                print(f"📏 Data size: {data_size} characters")
                
                # Parse and verify structure
                parsed_data = json.loads(data)
                if 'gamelist' in parsed_data:
                    print(f"✅ Data structure is correct")
                    print(f"📊 Contains {len(parsed_data['gamelist'])} contests")
                else:
                    print("❌ Data structure is incorrect")
                    return False
            else:
                print("❌ No data found")
                return False
        else:
            print("❌ Key not found")
            return False
            
        return True
        
    except Exception as e:
        print(f"❌ Error testing Redis key structure: {e}")
        return False

def main():
    """Main test function"""
    print("🏆 Contest List Test Suite")
    print("=" * 50)
    
    # Run tests
    test1_passed = test_contest_list()
    test2_passed = test_redis_key_structure()
    
    print("\n" + "=" * 50)
    print("📊 TEST RESULTS")
    print("=" * 50)
    print(f"Contest List Functionality: {'✅ PASSED' if test1_passed else '❌ FAILED'}")
    print(f"Redis Key Structure: {'✅ PASSED' if test2_passed else '❌ FAILED'}")
    
    if test1_passed and test2_passed:
        print("\n🎉 All tests passed!")
        print("✅ Contest list data is properly stored in Redis")
        print("🔑 Key: listcontest:current")
        print("📊 Ready for use by the Go Socket.IO service")
        return 0
    else:
        print("\n❌ Some tests failed")
        return 1

if __name__ == "__main__":
    exit(main()) 