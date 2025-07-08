const WebSocket = require('ws');


//**************************************************** */
// node test_debug.js restore 39df02ce-421a-4e98-b576-c546b9642fae
//**************************************************** */

//**************************************************** */
// node test_debug.js full
//**************************************************** */

// Get test mode from command line arguments
const testMode = process.argv[2] || 'full'; // 'full' or 'restore'
const sessionToken = process.argv[3] || ''; // Session token for restore mode

console.log('🔍 Authentication Test - Mode:', testMode.toUpperCase());
console.log('==================================================');

if (testMode === 'restore' && !sessionToken) {
  console.log('❌ Error: Session token required for restore mode');
  console.log('Usage: node test_debug.js restore <session_token>');
  process.exit(1);
}

const ws = new WebSocket('ws://localhost:8088/socket.io/?EIO=4&transport=websocket');

let currentSessionToken = sessionToken;
let otpCode = '';
let jwtToken = '';

const FCM_TOKEN = 'fcm_token_that_is_at_least_120_characters_long_for_testing_purposes_and_validation_check_1234567890_abcdefghijklmnopqrstuvwxyz_ABCDEFGHIJKLMNOPQRSTUVWXYZ_1234567890';

ws.on('open', function open() {
  console.log('✅ WebSocket connected');
  
  // Send Socket.IO handshake
  ws.send('40');
});

ws.on('message', function message(data) {
  const messageStr = data.toString();
  console.log('📡 Received:', messageStr);
  
  // Parse the message to extract data
  if (messageStr.startsWith('42')) {
    try {
      const jsonData = JSON.parse(messageStr.substring(2));
      
      if (jsonData[0] === 'otp:sent') {
        console.log('🎉 Login successful!');
        console.log('📱 Mobile:', jsonData[1].mobile_no);
        console.log('🔑 Session Token:', jsonData[1].session_token);
        console.log('🔢 OTP Code:', jsonData[1].otp);
        console.log('👤 Is New User:', jsonData[1].is_new_user);
        
        currentSessionToken = jsonData[1].session_token;
        otpCode = jsonData[1].otp.toString();
        
        // Step 2: Verify OTP
        setTimeout(() => {
          console.log('\n🔄 Step 2: Verifying OTP...');
          const otpEvent = ['verify:otp', {
            mobile_no: '1234567890',
            session_token: currentSessionToken,
            otp: otpCode
          }];
          
          console.log('📤 Sending OTP verification:', JSON.stringify(otpEvent));
          ws.send('42' + JSON.stringify(otpEvent));
        }, 2000);
      }
      
      else if (jsonData[0] === 'otp:verified') {
        console.log('🎉 OTP verification successful!');
        console.log('✅ User authenticated');
        console.log('🎫 JWT Token:', jsonData[1].jwt_token);
        
        jwtToken = jsonData[1].jwt_token;
        
        // Step 3: Set Profile
        setTimeout(() => {
          console.log('\n🔄 Step 3: Setting up profile...');
          const profileEvent = ['set:profile', {
            mobile_no: '1234567890',
            session_token: currentSessionToken,
            full_name: 'John Doe',
            state: 'California',
            city: 'Los Angeles',
            country: 'USA'
          }];
          
          console.log('📤 Sending profile setup:', JSON.stringify(profileEvent));
          ws.send('42' + JSON.stringify(profileEvent));
        }, 2000);
      }
      
      else if (jsonData[0] === 'profile:set') {
        console.log('🎉 Profile setup successful!');
        console.log('✅ User profile configured');
        
        // Step 4: Set Language
        setTimeout(() => {
          console.log('\n🔄 Step 4: Setting language preference...');
          const languageEvent = ['set:language', {
            mobile_no: '1234567890',
            session_token: currentSessionToken,
            language_code: 'en'
          }];
          
          console.log('📤 Sending language setup:', JSON.stringify(languageEvent));
          ws.send('42' + JSON.stringify(languageEvent));
        }, 2000);
      }
      
      else if (jsonData[0] === 'language:set') {
        console.log('🎉 Language setup successful!');
        console.log('✅ User fully configured');
        
        // Step 5: Test protected event (main:screen)
        setTimeout(() => {
          console.log('\n🔄 Step 5: Testing protected event (main:screen)...');
          const mainScreenEvent = ['main:screen', {
            mobile_no: '1234567890',
            session_token: currentSessionToken,
            jwt_token: jwtToken,
            device_id: 'test_device_123',
            fcm_token: FCM_TOKEN,
            message_type: 'game_list'
          }];
          
          console.log('📤 Sending main screen request:', JSON.stringify(mainScreenEvent));
          ws.send('42' + JSON.stringify(mainScreenEvent));
        }, 2000);
      }
      
      else if (jsonData[0] === 'session:restored') {
        console.log('✅ Session restored successfully!');
        console.log('📱 Mobile:', jsonData[1].mobile_no);
        console.log('🔑 Session Token:', jsonData[1].session_token);
        console.log('🆔 Socket ID:', jsonData[1].socket_id);
        
        // Test protected event after restoration
        setTimeout(() => {
          console.log('\n🔄 Testing protected event after restoration...');
          const mainScreenEvent = ['main:screen', {
            mobile_no: '1234567890',
            session_token: currentSessionToken,
            jwt_token: jwtToken,
            device_id: 'test_device_123',
            fcm_token: FCM_TOKEN,
            message_type: 'game_list'
          }];
          
          console.log('📤 Sending main screen request:', JSON.stringify(mainScreenEvent));
          ws.send('42' + JSON.stringify(mainScreenEvent));
        }, 2000);
      }
      
      else if (jsonData[0] === 'main:screen:game:list') {
        console.log('🎉 Protected event successful!');
        console.log('✅ User can access game features');
        console.log('📊 Game list received');
        
        if (testMode === 'restore') {
          console.log('\n🎯 Session restoration test completed successfully!');
          console.log('💡 Session token for future tests:', currentSessionToken);
        } else {
          console.log('\n🎯 Authentication flow test completed successfully!');
        }
        
        // Test complete - disconnect
        setTimeout(() => {
          ws.close();
          process.exit(0);
        }, 2000);
      }
      
      else if (jsonData[0] === 'connection_error' || jsonData[0] === 'authentication_error') {
        console.log('❌ Error received:', jsonData[1]);
        
        if (testMode === 'restore') {
          console.log('💡 Session may have expired or been logged out');
          console.log('💡 Try running the full authentication flow first');
        }
        
        setTimeout(() => {
          ws.close();
          process.exit(1);
        }, 2000);
      }
      
    } catch (error) {
      console.log('❌ Error parsing message:', error.message);
    }
  }
});

ws.on('error', function error(err) {
  console.log('❌ WebSocket error:', err.message);
});

ws.on('close', function close() {
  console.log('🔌 WebSocket closed');
});

// Send initial request based on test mode
setTimeout(() => {
  if (testMode === 'restore') {
    console.log('\n🔄 Testing session restoration...');
    console.log('🔑 Using session token:', currentSessionToken);
    
    const restoreEvent = ['restore:session', {
      session_token: currentSessionToken
    }];
    
    console.log('📤 Sending session restoration:', JSON.stringify(restoreEvent));
    ws.send('42' + JSON.stringify(restoreEvent));
  } else {
    console.log('\n🔄 Step 1: Sending login request...');
    const loginEvent = ['login', {
      mobile_no: '1234567890',
      device_id: 'test_device_123',
      fcm_token: FCM_TOKEN
    }];
    
    console.log('📤 Sending login:', JSON.stringify(loginEvent));
    ws.send('42' + JSON.stringify(loginEvent));
  }
}, 2000);

// Disconnect after 60 seconds (should complete before then)
setTimeout(() => {
  console.log('⏰ Test timeout - disconnecting');
  ws.close();
  process.exit(0);
}, 60000);