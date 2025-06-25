#!/usr/bin/env python3
"""
Quick Test Script
Tests Redis connection and basic functionality
"""

def test_redis_import():
    """Test if Redis can be imported and used"""
    print("🔍 Testing Redis import...")
    
    try:
        import redis
        print("✅ Redis module imported successfully")
        
        # Test creating Redis client
        redis_client = redis.Redis(host='127.0.0.1', port=6379, db=0, decode_responses=True)
        print("✅ Redis client created successfully")
        
        # Test connection
        redis_client.ping()
        print("✅ Redis connection successful")
        
        return True
        
    except ImportError as e:
        print(f"❌ Redis import failed: {e}")
        print("💡 Please run: pip install redis==5.0.1")
        return False
        
    except Exception as e:
        print(f"❌ Redis connection failed: {e}")
        print("💡 Make sure Redis server is running on 127.0.0.1:6379")
        return False

def test_socketio_import():
    """Test if Socket.IO can be imported"""
    print("\n🔍 Testing Socket.IO import...")
    
    try:
        import socketio
        print("✅ Socket.IO module imported successfully")
        return True
        
    except ImportError as e:
        print(f"❌ Socket.IO import failed: {e}")
        print("💡 Please run: pip install python-socketio==5.10.0")
        return False

if __name__ == "__main__":
    print("🔧 Quick Test Script")
    print("=" * 40)
    
    redis_ok = test_redis_import()
    socketio_ok = test_socketio_import()
    
    print("\n📋 Test Summary:")
    if redis_ok and socketio_ok:
        print("✅ All tests passed! You can now run the main scripts.")
        print("🚀 Try running: python test_trigger_flow.py")
    else:
        print("❌ Some tests failed. Please fix the issues above.") 