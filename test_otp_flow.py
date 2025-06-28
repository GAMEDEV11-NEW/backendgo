#!/usr/bin/env python3
"""
Test script for OTP flow in GOSOCKET application
Tests: Login -> OTP Verification -> Resend OTP
"""

import socketio
import time
import json

# Create Socket.IO client
sio = socketio.Client()

# Test data
TEST_MOBILE = "1234567890"
TEST_DEVICE_ID = "test_device_123"
TEST_FCM_TOKEN = "fcm_token_that_is_at_least_100_characters_long_for_testing_purposes_and_validation_check_1234567890"

# Global variables to store responses
login_response = None
otp_response = None
resend_response = None

@sio.event
def connect():
    print("✅ Connected to server")
    
    # Send device info
    device_info = {
        "device_id": TEST_DEVICE_ID,
        "device_type": "mobile",
        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
        "manufacturer": "Test Manufacturer",
        "model": "Test Model",
        "firmware_version": "1.0.0",
        "capabilities": ["touch", "camera", "gps"]
    }
    sio.emit("device:info", device_info)

@sio.event
def disconnect():
    print("❌ Disconnected from server")

@sio.on("device:info:ack")
def on_device_info_ack(data):
    print(f"📱 Device info acknowledged: {data}")
    
    # Send login request
    login_data = {
        "mobile_no": TEST_MOBILE,
        "device_id": TEST_DEVICE_ID,
        "fcm_token": TEST_FCM_TOKEN,
        "email": "test@example.com"
    }
    print(f"🔐 Sending login request: {login_data}")
    sio.emit("login", login_data)

@sio.on("otp:sent")
def on_otp_sent(data):
    global login_response
    login_response = data
    print(f"📨 OTP sent response: {json.dumps(data, indent=2)}")
    
    # Extract OTP and session token
    otp = data.get("otp")
    session_token = data.get("session_token")
    
    if otp and session_token:
        print(f"🔢 Received OTP: {otp}")
        print(f"🔑 Session Token: {session_token}")
        
        # Wait a moment then verify OTP
        time.sleep(2)
        
        # Send OTP verification
        otp_verify_data = {
            "mobile_no": TEST_MOBILE,
            "session_token": session_token,
            "otp": str(otp)
        }
        print(f"✅ Sending OTP verification: {otp_verify_data}")
        sio.emit("verify:otp", otp_verify_data)
    else:
        print("❌ Missing OTP or session token in response")

@sio.on("otp:verified")
def on_otp_verified(data):
    global otp_response
    otp_response = data
    print(f"✅ OTP verified response: {json.dumps(data, indent=2)}")
    
    # Extract JWT token
    jwt_token = data.get("jwt_token")
    session_token = data.get("session_token")
    
    if jwt_token and session_token:
        print(f"🎫 JWT Token: {jwt_token}")
        
        # Test resend OTP functionality
        time.sleep(2)
        resend_data = {
            "mobile_no": TEST_MOBILE,
            "session_token": session_token
        }
        print(f"🔄 Sending resend OTP request: {resend_data}")
        sio.emit("resend:otp", resend_data)
    else:
        print("❌ Missing JWT token in response")

@sio.on("otp:resent")
def on_otp_resent(data):
    global resend_response
    resend_response = data
    print(f"🔄 OTP resent response: {json.dumps(data, indent=2)}")
    
    # Test completed
    print("\n🎉 OTP flow test completed successfully!")
    print("📊 Summary:")
    print(f"  - Login: {'✅' if login_response else '❌'}")
    print(f"  - OTP Verification: {'✅' if otp_response else '❌'}")
    print(f"  - Resend OTP: {'✅' if resend_response else '❌'}")
    
    # Disconnect after test
    time.sleep(1)
    sio.disconnect()

@sio.on("connection_error")
def on_connection_error(data):
    print(f"❌ Connection error: {json.dumps(data, indent=2)}")

def main():
    print("🚀 Starting OTP flow test...")
    print(f"📱 Test Mobile: {TEST_MOBILE}")
    print(f"📱 Test Device ID: {TEST_DEVICE_ID}")
    print("=" * 50)
    
    try:
        # Connect to the server
        sio.connect("http://localhost:8088")
        
        # Keep the script running
        sio.wait()
        
    except Exception as e:
        print(f"❌ Test failed: {e}")
    finally:
        if sio.connected:
            sio.disconnect()

if __name__ == "__main__":
    main() 