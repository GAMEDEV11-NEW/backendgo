const { io } = require('socket.io-client');
const chalk = require('chalk');

// Test Configuration
const CONFIG = {
    SERVER_URL: 'http://localhost:8088',
    TIMEOUT: 10000
};

// Utility Functions
const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

const generateDeviceId = () => `device_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
const generateMobileNo = () => `+1${Math.floor(Math.random() * 9000000000) + 1000000000}`;
const generateFCMToken = () => `fcm_${Math.random().toString(36).substr(2, 20)}`;

const log = {
    info: (msg) => console.log(chalk.blue('ℹ'), msg),
    success: (msg) => console.log(chalk.green('✅'), msg),
    error: (msg) => console.log(chalk.red('❌'), msg),
    warning: (msg) => console.log(chalk.yellow('⚠'), msg),
    test: (msg) => console.log(chalk.cyan('🧪'), msg),
    contest: (msg) => console.log(chalk.magenta('🏆'), msg)
};

class ContestListTester {
    constructor() {
        this.socket = null;
        this.testData = {
            deviceId: generateDeviceId(),
            mobileNo: generateMobileNo(),
            fcmToken: generateFCMToken(),
            sessionToken: null,
            otp: null,
            jwtToken: null
        };
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

    async testLogin() {
        log.test('Testing Login...');
        
        const loginData = {
            mobile_no: this.testData.mobileNo,
            device_id: this.testData.deviceId,
            fcm_token: this.testData.fcmToken,
            email: 'test@example.com'
        };

        const responsePromise = this.waitForEvent('otp:sent');
        this.socket.emit('login', loginData);
        
        const response = await responsePromise;
        
        if (!response || response.status !== 'success') {
            throw new Error('Login failed');
        }

        this.testData.sessionToken = response.session_token;
        this.testData.otp = response.otp;

        log.success(`Login successful for: ${response.mobile_no}`);
        log.info(`Session token: ${response.session_token}`);
        log.info(`OTP: ${response.otp}`);
        log.info(`Is new user: ${response.is_new_user}`);
    }

    async testOTPVerification() {
        log.test('Testing OTP Verification...');
        
        if (!this.testData.sessionToken || !this.testData.otp) {
            throw new Error('Session token or OTP not available');
        }

        const otpData = {
            mobile_no: this.testData.mobileNo,
            session_token: this.testData.sessionToken,
            otp: this.testData.otp.toString()
        };

        const responsePromise = this.waitForEvent('otp:verified');
        this.socket.emit('verify:otp', otpData);
        
        const response = await responsePromise;
        
        if (!response || response.status !== 'success') {
            throw new Error('OTP verification failed');
        }

        this.testData.jwtToken = response.jwt_token;

        log.success('OTP verification successful');
        log.info(`User status: ${response.user_status}`);
        log.info(`JWT token received: ${response.jwt_token ? 'Yes' : 'No'}`);
        log.info(`Device ID: ${response.device_id}`);
    }

    async testContestList() {
        log.contest('Testing Contest List Event...');
        
        if (!this.testData.jwtToken) {
            throw new Error('JWT token not available - run authentication first');
        }

        const contestData = {
            mobile_no: this.testData.mobileNo,
            fcm_token: this.testData.fcmToken,
            jwt_token: this.testData.jwtToken,
            device_id: this.testData.deviceId,
            message_type: 'contest_list'
        };

        log.info('Sending contest data:');
        log.info(`  Mobile: ${contestData.mobile_no}`);
        log.info(`  Device ID: ${contestData.device_id}`);
        log.info(`  Message Type: ${contestData.message_type}`);
        log.info(`  JWT Token: ${contestData.jwt_token.substring(0, 50)}...`);
        log.info(`  FCM Token: ${contestData.fcm_token.substring(0, 50)}...`);

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

    async testContestJoin() {
        log.contest('Testing Contest Join Event...');
        
        if (!this.testData.jwtToken) {
            throw new Error('JWT token not available - run authentication first');
        }

        const joinData = {
            mobile_no: this.testData.mobileNo,
            fcm_token: this.testData.fcmToken,
            jwt_token: this.testData.jwtToken,
            device_id: this.testData.deviceId,
            contest_id: 'contest_123',
            team_name: 'Test Team',
            team_size: 2
        };

        log.info('Sending contest join data:');
        log.info(`  Contest ID: ${joinData.contest_id}`);
        log.info(`  Team Name: ${joinData.team_name}`);
        log.info(`  Team Size: ${joinData.team_size}`);

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

    async runFullTest() {
        console.log(chalk.bold.magenta('\n🏆 CONTEST LIST TEST SUITE'));
        console.log('='.repeat(60));
        
        try {
            // Step 1: Connect
            await this.connect();
            await sleep(1000);

            // Step 2: Login
            await this.testLogin();
            await sleep(1000);

            // Step 3: OTP Verification
            await this.testOTPVerification();
            await sleep(1000);

            // Step 4: Test Contest List
            const contestListResponse = await this.testContestList();
            await sleep(1000);

            // Step 5: Test Contest Join
            const contestJoinResponse = await this.testContestJoin();
            await sleep(1000);

            // Summary
            console.log('\n' + '='.repeat(60));
            console.log(chalk.bold.green('🎉 ALL TESTS PASSED!'));
            console.log('='.repeat(60));
            console.log(chalk.green('✅ Authentication successful'));
            console.log(chalk.green('✅ Contest list retrieved'));
            console.log(chalk.green('✅ Contest join successful'));
            console.log(`📊 Contests found: ${contestListResponse.data?.gamelist?.length || 0}`);
            console.log(`🏆 Contest joined: ${contestJoinResponse.contest_id}`);

        } catch (error) {
            console.log('\n' + '='.repeat(60));
            console.log(chalk.bold.red('❌ TEST FAILED'));
            console.log('='.repeat(60));
            console.log(chalk.red(`Error: ${error.message}`));
            console.log(chalk.yellow('💡 Make sure the server is running on http://localhost:8088'));
            console.log(chalk.yellow('💡 Check if Redis is running and has contest data'));
        } finally {
            await this.disconnect();
        }
    }

    async runContestListOnly() {
        console.log(chalk.bold.magenta('\n🏆 CONTEST LIST ONLY TEST'));
        console.log('='.repeat(60));
        
        try {
            await this.connect();
            await sleep(1000);
            await this.testLogin();
            await sleep(1000);
            await this.testOTPVerification();
            await sleep(1000);
            await this.testContestList();
            
            console.log('\n' + '='.repeat(60));
            console.log(chalk.bold.green('✅ CONTEST LIST TEST PASSED!'));
            console.log('='.repeat(60));

        } catch (error) {
            console.log('\n' + '='.repeat(60));
            console.log(chalk.bold.red('❌ CONTEST LIST TEST FAILED'));
            console.log('='.repeat(60));
            console.log(chalk.red(`Error: ${error.message}`));
        } finally {
            await this.disconnect();
        }
    }
}

// Main execution
async function main() {
    const tester = new ContestListTester();
    
    // Check command line arguments
    const args = process.argv.slice(2);
    
    if (args.includes('--full') || args.includes('-f')) {
        await tester.runFullTest();
    } else {
        await tester.runContestListOnly();
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

module.exports = { ContestListTester }; 