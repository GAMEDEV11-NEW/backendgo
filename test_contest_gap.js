const io = require('socket.io-client');

// Connect to the socket server
const socket = io('http://localhost:3000');

// Test data for contest gap request
const testGapRequest = {
    mobile_no: "1234567890",
    fcm_token: "fcm_token_1234567890_abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890",
    jwt_token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJtb2JpbGVfbm8iOiIxMjM0NTY3ODkwIiwiZGV2aWNlX2lkIjoiZGV2aWNlXzEyMyIsImZjbV90b2tlbiI6ImZjbV90b2tlbl8xMjM0NTY3ODkwX2FiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6MTIzNDU2Nzg5MGFiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6MTIzNDU2Nzg5MGFiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6MTIzNDU2Nzg5MCIsImlhdCI6MTczNTY3ODkwMCwiZXhwIjoxNzM1NzY1MzAwfQ.test_signature",
    device_id: "device_123",
    message_type: "win_price_gap"
};

// Test different price gap scenarios
const testScenarios = [
    {
        name: "Win Price Gap Analysis",
        data: {
            ...testGapRequest,
            message_type: "win_price_gap"
        }
    },
    {
        name: "Entry Fee Gap Analysis",
        data: {
            ...testGapRequest,
            message_type: "entry_fee_gap"
        }
    },
    {
        name: "Combined Price Gap Analysis",
        data: {
            ...testGapRequest,
            message_type: "price_gap"
        }
    }
];

socket.on('connect', () => {
    console.log('✅ Connected to server');
    console.log('Socket ID:', socket.id);
    
    // Test each scenario
    testScenarios.forEach((scenario, index) => {
        setTimeout(() => {
            console.log(`\n🧪 Testing: ${scenario.name}`);
            socket.emit('list:contest:gap', scenario.data);
        }, index * 2000); // Send each test 2 seconds apart
    });
});

socket.on('list:contest:gap:response', (response) => {
    console.log('\n💰 Contest Gap Response Received:');
    console.log('Status:', response.status);
    console.log('Message:', response.message);
    console.log('Message Type:', response.message_type);
    
    if (response.data && response.data.price_gaps) {
        console.log('\n📊 Price Gap Analysis:');
        console.log('Total Gaps Found:', response.data.price_gaps.length);
        
        response.data.price_gaps.forEach((gap, index) => {
            console.log(`\n${index + 1}. ${gap.type.toUpperCase()} - ${gap.range_name}`);
            console.log(`   Price Range: ${gap.min_price} - ${gap.max_price}`);
            console.log(`   Contests: ${gap.contest_count}`);
            console.log(`   Percentage: ${gap.percentage.toFixed(1)}%`);
        });
        
        if (response.data.summary) {
            console.log('\n📈 Summary Statistics:');
            console.log('Total Contests:', response.data.summary.total_contests);
            console.log('Filter Type:', response.data.summary.filter_type);
            
            if (response.data.summary.win_price_range) {
                console.log('Win Price Range:', response.data.summary.win_price_range);
            }
            
            if (response.data.summary.entry_fee_range) {
                console.log('Entry Fee Range:', response.data.summary.entry_fee_range);
            }
            
            if (response.data.summary.avg_win_price) {
                console.log('Average Win Price:', response.data.summary.avg_win_price.toFixed(2));
            }
            
            if (response.data.summary.avg_entry_fee) {
                console.log('Average Entry Fee:', response.data.summary.avg_entry_fee.toFixed(2));
            }
        }
    } else {
        console.log('❌ No gap data found');
    }
});

socket.on('connection_error', (error) => {
    console.error('❌ Connection Error:', error);
});

socket.on('disconnect', () => {
    console.log('🔌 Disconnected from server');
});

// Handle any other errors
socket.on('error', (error) => {
    console.error('❌ Socket Error:', error);
});

// Cleanup after 10 seconds
setTimeout(() => {
    console.log('\n🏁 Test completed, disconnecting...');
    socket.disconnect();
    process.exit(0);
}, 15000); 