const WebSocket = require('ws');
const crypto = require('crypto');

// --- Encryption logic (matches Unity/Go) ---
function padKey(mobile) {
    let buf = Buffer.alloc(32, 0);
    Buffer.from(mobile).copy(buf);
    return buf;
}

function encryptUserData(playerdata, mobileNo) {
    const key = padKey(mobileNo);
    const iv = Buffer.alloc(16, 0); // 16 zeros
    const json = JSON.stringify(playerdata);

    const cipher = crypto.createCipheriv('aes-256-cbc', key, iv);
    let encrypted = cipher.update(json, 'utf8', 'base64');
    encrypted += cipher.final('base64');
    return encrypted;
}

// --- Prepare payload ---
const mobileNo = "5985899652";
const playerdata = {
    mobile_no: mobileNo,
    device_id: "abc123"
};
const user_data = encryptUserData(playerdata, mobileNo);

const ws = new WebSocket('ws://localhost:8088/socket.io/?EIO=4&transport=websocket');

ws.on('open', function open() {
    console.log('✅ WebSocket connected');
    ws.send('40'); // Socket.IO handshake

    setTimeout(() => {
        // Send login event in Socket.IO format
        const loginEvent = ['login', { mobile_no: mobileNo, user_data: user_data }];
        console.log('📤 Sending login:', JSON.stringify(loginEvent));
        ws.send('42' + JSON.stringify(loginEvent));
    }, 1000);
});

ws.on('message', function message(data) {
    const messageStr = data.toString();
    console.log('📡 Received:', messageStr);

    if (messageStr.startsWith('42')) {
        try {
            const jsonData = JSON.parse(messageStr.substring(2));
            if (jsonData[0] === 'otp:sent') {
                console.log('🎉 Login successful!');
                console.log('🔑 Session Token:', jsonData[1].session_token);
                console.log('🔢 OTP Code:', jsonData[1].otp);

                // Now send OTP verification
                setTimeout(() => {
                    const otpEvent = ['verify:otp', {
                        mobile_no: mobileNo,
                        session_token: jsonData[1].session_token,
                        otp: jsonData[1].otp.toString()
                    }];
                    console.log('📤 Sending OTP verification:', JSON.stringify(otpEvent));
                    ws.send('42' + JSON.stringify(otpEvent));
                }, 1000);
            } else if (jsonData[0] === 'otp:verified') {
                console.log('🎉 OTP verification successful!');
                ws.close();
            } else if (jsonData[0] === 'connection_error' || jsonData[0] === 'authentication_error') {
                console.log('❌ Error received:', jsonData[1]);
                ws.close();
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