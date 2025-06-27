const { io } = require('socket.io-client');
const chalk = require('chalk');

// Test Configuration
const CONFIG = {
    SERVER_URL: 'http://localhost:8088',
    TIMEOUT: 10000
};

const log = {
    info: (msg) => console.log(chalk.blue('ℹ'), msg),
    success: (msg) => console.log(chalk.green('✅'), msg),
    error: (msg) => console.log(chalk.red('❌'), msg),
    warning: (msg) => console.log(chalk.yellow('⚠'), msg),
    contest: (msg) => console.log(chalk.magenta('🏆'), msg)
};

class CustomContestTester {
    constructor() {
        this.socket = null;
    }

    async connect() {
        log.info('Connecting to Socket.IO server...');
        
        try {
            this.socket = io(CONFIG.SERVER_URL, {
                transports: ['websocket', 'polling'],
                timeout: CONFIG.TIMEOUT,
                forceNew: true
            });

            return new Promise((resolve, reject) => {
                const timeout = setTimeout(() => {
                    reject(new Error('Connection timeout'));
                }, CONFIG.TIMEOUT);

                this.socket.on('connect', () => {
                    clearTimeout(timeout);
                    log.success('Connected to Socket.IO server');
                    log.info(`Socket ID: ${this.socket.id}`);
                    resolve();
                });

                this.socket.on('connect_error', (error) => {
                    clearTimeout(timeout);
                    log.error('Failed to connect to Socket.IO server');
                    reject(error);
                });
            });
        } catch (error) {
            log.error('Connection failed');
            throw error;
        }
    }

    async disconnect() {
        if (this.socket) {
            this.socket.disconnect();
            this.socket = null;
            log.info('Disconnected from server');
        }
    }

    async waitForEvent(eventName, timeout = CONFIG.TIMEOUT) {
        return new Promise((resolve, reject) => {
            const timeoutId = setTimeout(() => {
                reject(new Error(`Timeout waiting for event: ${eventName}`));
            }, timeout);

            this.socket.once(eventName, (data) => {
                clearTimeout(timeoutId);
                resolve(data);
            });
        });
    }

    async testContestListWithCustomData(contestData) {
        log.contest('Testing Contest List with Custom Data...');
        
        log.info('Sending contest data:');
        log.info(`  Mobile: ${contestData.mobile_no}`);
        log.info(`  Device ID: ${contestData.device_id}`);
        log.info(`  Message Type: ${contestData.message_type}`);
        log.info(`  JWT Token: ${contestData.jwt_token ? contestData.jwt_token.substring(0, 50) + '...' : 'NOT PROVIDED'}`);
        log.info(`  FCM Token: ${contestData.fcm_token ? contestData.fcm_token.substring(0, 50) + '...' : 'NOT PROVIDED'}`);

        const responsePromise = this.waitForEvent('contest:list:response');
        this.socket.emit('list:contest', contestData);
        
        const response = await responsePromise;
        
        if (!response || response.status !== 'success') {
            throw new Error('Contest list data retrieval failed');
        }

        log.success('Contest list data retrieved successfully!');
        log.info(`Status: ${response.status}`);
        log.info(`Message: ${response.message}`);
        log.info(`Event: ${response.event}`);
        log.info(`Mobile: ${response.mobile_no}`);
        log.info(`Device ID: ${response.device_id}`);
        
        if (response.data && response.data.gamelist) {
            log.info(`Contest list items: ${response.data.gamelist.length}`);
            
            // Display contest details
            response.data.gamelist.forEach((contest, index) => {
                log.contest(`\nContest ${index + 1}:`);
                log.info(`  Name: ${contest.contestName}`);
                log.info(`  ID: ${contest.contestId}`);
                log.info(`  Status: ${contest.status}`);
                log.info(`  Difficulty: ${contest.difficulty}`);
                log.info(`  Duration: ${contest.duration}`);
                log.info(`  Participants: ${contest.currentParticipants}/${contest.maxParticipants}`);
                log.info(`  Reward: ${contest.reward?.currency} ${contest.reward?.amount}`);
                log.info(`  Categories: ${contest.categories?.join(', ')}`);
            });
        } else {
            log.warning('No contest data found in response');
        }

        if (response.user_info) {
            log.info('\nUser Info:');
            log.info(`  User ID: ${response.user_info.user_id}`);
            log.info(`  Full Name: ${response.user_info.full_name}`);
            log.info(`  Status: ${response.user_info.status}`);
            log.info(`  Language: ${response.user_info.language}`);
        }

        return response;
    }

    async testContestJoinWithCustomData(joinData) {
        log.contest('Testing Contest Join with Custom Data...');
        
        log.info('Sending contest join data:');
        log.info(`  Contest ID: ${joinData.contest_id}`);
        log.info(`  Team Name: ${joinData.team_name || 'N/A'}`);
        log.info(`  Team Size: ${joinData.team_size || 'N/A'}`);

        const responsePromise = this.waitForEvent('contest:join:response');
        this.socket.emit('contest:join', joinData);
        
        const response = await responsePromise;
        
        if (!response || response.status !== 'success') {
            throw new Error('Contest join failed');
        }

        log.success('Contest join successful!');
        log.info(`Status: ${response.status}`);
        log.info(`Message: ${response.message}`);
        log.info(`Contest ID: ${response.contest_id}`);
        log.info(`Team ID: ${response.team_id || 'N/A'}`);
        log.info(`Join Time: ${response.join_time}`);

        if (response.data) {
            log.info('\nJoin Data:');
            log.info(`  Contest ID: ${response.data.contest_id}`);
            log.info(`  Team Name: ${response.data.team_name}`);
            log.info(`  Team Size: ${response.data.team_size}`);
            log.info(`  Join Status: ${response.data.join_status}`);
            log.info(`  Next Steps: ${response.data.next_steps}`);
        }

        return response;
    }
}

// Example usage and main function
async function main() {
    const tester = new CustomContestTester();
    
    try {
        await tester.connect();
        
        // Example 1: Test with sample data
        console.log(chalk.bold.magenta('\n🏆 TESTING WITH SAMPLE DATA'));
        console.log('='.repeat(60));
        
        const sampleContestData = {
            mobile_no: "+1234567890",
            fcm_token: "fcm_test_token_1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
            jwt_token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJtb2JpbGVfbm8iOiIrMTIzNDU2Nzg5MCIsImRldmljZV9pZCI6InRlc3RfZGV2aWNlIiwiZmNtX3Rva2VuIjoiZmNtX3Rlc3RfdG9rZW5fMTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTAiLCJleHAiOjE3MzU2ODgwMDAsImlhdCI6MTczNTYwMTYwMCwibmJmIjoxNzM1NjAxNjAwLCJpc3MiOiJnYW1lLWFkbWluLWJhY2tlbmQiLCJzdWIiOiIrMTIzNDU2Nzg5MCJ9.test_signature",
            device_id: "test_device",
            message_type: "contest_list"
        };

        await tester.testContestListWithCustomData(sampleContestData);
        
        // Example 2: Test contest join
        const sampleJoinData = {
            mobile_no: "+1234567890",
            fcm_token: "fcm_test_token_1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
            jwt_token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJtb2JpbGVfbm8iOiIrMTIzNDU2Nzg5MCIsImRldmljZV9pZCI6InRlc3RfZGV2aWNlIiwiZmNtX3Rva2VuIjoiZmNtX3Rlc3RfdG9rZW5fMTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTAiLCJleHAiOjE3MzU2ODgwMDAsImlhdCI6MTczNTYwMTYwMCwibmJmIjoxNzM1NjAxNjAwLCJpc3MiOiJnYW1lLWFkbWluLWJhY2tlbmQiLCJzdWIiOiIrMTIzNDU2Nzg5MCJ9.test_signature",
            device_id: "test_device",
            contest_id: "contest_123",
            team_name: "Test Team",
            team_size: 2
        };

        await tester.testContestJoinWithCustomData(sampleJoinData);
        
        console.log('\n' + '='.repeat(60));
        console.log(chalk.bold.green('✅ ALL CUSTOM TESTS PASSED!'));
        console.log('='.repeat(60));

    } catch (error) {
        console.log('\n' + '='.repeat(60));
        console.log(chalk.bold.red('❌ CUSTOM TEST FAILED'));
        console.log('='.repeat(60));
        console.log(chalk.red(`Error: ${error.message}`));
        console.log(chalk.yellow('💡 Make sure the server is running on http://localhost:8088'));
        console.log(chalk.yellow('💡 Check if Redis is running and has contest data'));
    } finally {
        await tester.disconnect();
    }
}

// Handle process termination
process.on('SIGINT', () => {
    console.log(chalk.blue('\n👋 Test interrupted. Goodbye!'));
    process.exit(0);
});

// Run if this file is executed directly
if (require.main === module) {
    main().catch(error => {
        console.error(chalk.red(`Fatal error: ${error.message}`));
        process.exit(1);
    });
}

module.exports = { CustomContestTester }; 