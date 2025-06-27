const io = require('socket.io-client');
const chalk = require('chalk');

const log = {
    info: (msg) => console.log(chalk.blue('ℹ'), msg),
    success: (msg) => console.log(chalk.green('✅'), msg),
    error: (msg) => console.log(chalk.red('❌'), msg),
    warning: (msg) => console.log(chalk.yellow('⚠'), msg),
    contest: (msg) => console.log(chalk.magenta('🏆'), msg)
};

class ContestDataTester {
    constructor() {
        this.socket = null;
        this.testData = {
            deviceId: 'test_device_123',
            mobileNo: '1234567890',
            fcmToken: 'fcm_token_test_123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890',
            jwtToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJtb2JpbGVfbm8iOiIxMjM0NTY3ODkwIiwiZGV2aWNlX2lkIjoidGVzdF9kZXZpY2VfMTIzIiwiZmNtX3Rva2VuIjoiZmNtX3Rva2VuX3Rlc3RfMTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTAiLCJleHAiOjE3MzU2NzY4MDB9.test_signature'
        };
    }

    async connect() {
        return new Promise((resolve, reject) => {
            log.info('Connecting to socket server...');
            
            this.socket = io('http://localhost:8088', {
                transports: ['websocket'],
                timeout: 5000
            });

            this.socket.on('connect', () => {
                log.success(`Connected to server with ID: ${this.socket.id}`);
                resolve();
            });

            this.socket.on('connect_error', (error) => {
                log.error(`Connection failed: ${error.message}`);
                reject(error);
            });

            this.socket.on('disconnect', (reason) => {
                log.warning(`Disconnected: ${reason}`);
            });
        });
    }

    waitForEvent(eventName, timeout = 10000) {
        return new Promise((resolve, reject) => {
            const timer = setTimeout(() => {
                reject(new Error(`Timeout waiting for event: ${eventName}`));
            }, timeout);

            this.socket.once(eventName, (data) => {
                clearTimeout(timer);
                resolve(data);
            });
        });
    }

    async testContestListData() {
        log.contest('Testing Contest List Data Reception...');
        
        const contestData = {
            mobile_no: this.testData.mobileNo,
            fcm_token: this.testData.fcmToken,
            jwt_token: this.testData.jwtToken,
            device_id: this.testData.deviceId,
            message_type: 'contest_list'
        };

        log.info('Sending contest request:');
        log.info(`  Mobile: ${contestData.mobile_no}`);
        log.info(`  Device ID: ${contestData.device_id}`);
        log.info(`  Message Type: ${contestData.message_type}`);
        log.info(`  JWT Token: ${contestData.jwt_token.substring(0, 50)}...`);
        log.info(`  FCM Token: ${contestData.fcm_token.substring(0, 50)}...`);

        try {
            const responsePromise = this.waitForEvent('contest:list:response');
            this.socket.emit('list:contest', contestData);
            
            const response = await responsePromise;
            
            log.success('🎉 Contest list response received!');
            log.info(`Status: ${response.status}`);
            log.info(`Message: ${response.message}`);
            log.info(`Event: ${response.event}`);
            log.info(`Mobile: ${response.mobile_no}`);
            log.info(`Device ID: ${response.device_id}`);
            
            // Check if data exists
            if (response.data) {
                log.success('✅ Response.data exists');
                log.info(`Data type: ${typeof response.data}`);
                log.info(`Data keys: ${Object.keys(response.data)}`);
                
                if (response.data.gamelist) {
                    log.success(`✅ Found gamelist with ${response.data.gamelist.length} contests`);
                    
                    // Display first contest details
                    if (response.data.gamelist.length > 0) {
                        const firstContest = response.data.gamelist[0];
                        log.contest('\nFirst Contest Details:');
                        log.info(`  Contest ID: ${firstContest.contestId}`);
                        log.info(`  Name: ${firstContest.contestName}`);
                        log.info(`  Status: ${firstContest.status}`);
                        log.info(`  Difficulty: ${firstContest.difficulty}`);
                        log.info(`  Participants: ${firstContest.currentParticipants}/${firstContest.maxParticipants}`);
                        log.info(`  Reward: ${firstContest.reward?.currency} ${firstContest.reward?.amount}`);
                        log.info(`  Categories: ${firstContest.categories?.join(', ')}`);
                    }
                } else {
                    log.warning('❌ No gamelist found in response.data');
                }
            } else {
                log.error('❌ No data field in response');
            }

            if (response.user_info) {
                log.info('\nUser Info:');
                log.info(`  User ID: ${response.user_info.user_id}`);
                log.info(`  Full Name: ${response.user_info.full_name}`);
                log.info(`  Status: ${response.user_info.status}`);
                log.info(`  Language: ${response.user_info.language}`);
            }

            return response;
            
        } catch (error) {
            log.error(`Failed to receive contest list: ${error.message}`);
            throw error;
        }
    }

    async runTest() {
        try {
            await this.connect();
            await this.testContestListData();
            log.success('🎉 All tests completed successfully!');
        } catch (error) {
            log.error(`Test failed: ${error.message}`);
        } finally {
            if (this.socket) {
                this.socket.disconnect();
            }
        }
    }
}

// Run the test
const tester = new ContestDataTester();
tester.runTest(); 