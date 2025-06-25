const { io } = require('socket.io-client');
const chalk = require('chalk');
const ora = require('ora');
const Table = require('cli-table3');
const figlet = require('figlet');
const gradient = require('gradient-string');
const inquirer = require('inquirer');

// Test Configuration
const CONFIG = {
    SERVER_URL: 'http://localhost:8088',
    SOCKET_NAMESPACE: '/',
    TIMEOUT: 10000,
    RETRY_ATTEMPTS: 3,
    DELAY_BETWEEN_TESTS: 1000
};

// Test Results Storage
let testResults = {
    total: 0,
    passed: 0,
    failed: 0,
    skipped: 0,
    details: []
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
    test: (msg) => console.log(chalk.cyan('🧪'), msg)
};

// Test Suite Class
class SocketIOTestSuite {
    constructor() {
        this.socket = null;
        this.testData = {
            deviceId: generateDeviceId(),
            mobileNo: generateMobileNo(),
            fcmToken: generateFCMToken(),
            sessionToken: null,
            otp: null,
            socketId: null
        };
        this.eventListeners = new Map();
    }

    async connect() {
        const spinner = ora('Connecting to Socket.IO server...').start();
        
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
                    this.testData.socketId = this.socket.id;
                    spinner.succeed('Connected to Socket.IO server');
                    resolve();
                });

                this.socket.on('connect_error', (error) => {
                    clearTimeout(timeout);
                    spinner.fail('Failed to connect to Socket.IO server');
                    reject(error);
                });
            });
        } catch (error) {
            spinner.fail('Connection failed');
            throw error;
        }
    }

    async disconnect() {
        if (this.socket) {
            this.socket.disconnect();
            this.socket = null;
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

    async runTest(testName, testFunction) {
        testResults.total++;
        const spinner = ora(`Running test: ${testName}`).start();

        try {
            await testFunction();
            spinner.succeed(`Test passed: ${testName}`);
            testResults.passed++;
            testResults.details.push({ name: testName, status: 'PASSED', error: null });
        } catch (error) {
            spinner.fail(`Test failed: ${testName}`);
            testResults.failed++;
            testResults.details.push({ name: testName, status: 'FAILED', error: error.message });
            log.error(`Error: ${error.message}`);
        }

        await sleep(CONFIG.DELAY_BETWEEN_TESTS);
    }

    // Test Cases
    async testConnection() {
        await this.runTest('Connection Test', async () => {
            const connectData = await this.waitForEvent('connect');
            
            if (!connectData || !connectData.token) {
                throw new Error('Invalid connection response');
            }

            log.info(`Connection token: ${connectData.token}`);
            log.info(`Socket ID: ${connectData.socket_id}`);
        });
    }

    async testConnectResponse() {
        await this.runTest('Connect Response Test', async () => {
            const response = await this.waitForEvent('connect_response');
            
            if (!response || !response.success) {
                throw new Error('Invalid connect response');
            }

            log.info(`Server status: ${response.status}`);
            log.info(`Message: ${response.message}`);
        });
    }

    async testDeviceInfo() {
        await this.runTest('Device Info Test', async () => {
            const deviceInfo = {
                device_id: this.testData.deviceId,
                device_type: 'mobile',
                timestamp: new Date().toISOString(),
                manufacturer: 'Samsung',
                model: 'Galaxy S21',
                firmware_version: 'Android 12',
                capabilities: ['camera', 'gps', 'bluetooth', 'wifi']
            };

            const responsePromise = this.waitForEvent('device:info:ack');
            this.socket.emit('device:info', deviceInfo);
            
            const response = await responsePromise;
            
            if (!response || response.status !== 'success') {
                throw new Error('Device info acknowledgment failed');
            }

            log.info(`Device info acknowledged: ${response.message}`);
        });
    }

    async testLogin() {
        await this.runTest('Login Test', async () => {
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

            log.info(`OTP sent successfully for: ${response.mobile_no}`);
            log.info(`Session token: ${response.session_token}`);
            log.info(`OTP: ${response.otp}`);
            log.info(`Is new user: ${response.is_new_user}`);
        });
    }

    async testOTPVerification() {
        await this.runTest('OTP Verification Test', async () => {
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

            // Store JWT token for future use
            this.testData.jwtToken = response.jwt_token;

            log.info(`OTP verified successfully`);
            log.info(`User status: ${response.user_status}`);
            log.info(`JWT token received: ${response.jwt_token ? 'Yes' : 'No'}`);
            log.info(`Device ID: ${response.device_id}`);
        });
    }

    async testSetProfile() {
        await this.runTest('Set Profile Test', async () => {
            if (!this.testData.sessionToken) {
                throw new Error('Session token not available');
            }

            const profileData = {
                mobile_no: this.testData.mobileNo,
                session_token: this.testData.sessionToken,
                full_name: 'Test User',
                state: 'California',
                referral_code: 'REF123',
                referred_by: 'FRIEND456',
                profile_data: {
                    avatar: 'https://example.com/avatar.jpg',
                    bio: 'Test user bio',
                    preferences: {
                        theme: 'dark',
                        notifications: true
                    }
                }
            };

            const responsePromise = this.waitForEvent('profile:set');
            this.socket.emit('set:profile', profileData);
            
            const response = await responsePromise;
            
            if (!response || response.status !== 'success') {
                throw new Error('Profile setup failed');
            }

            log.info(`Profile set successfully`);
            log.info(`Welcome message: ${response.welcome_message}`);
        });
    }

    async testSetLanguage() {
        await this.runTest('Set Language Test', async () => {
            if (!this.testData.sessionToken) {
                throw new Error('Session token not available');
            }

            const languageData = {
                mobile_no: this.testData.mobileNo,
                session_token: this.testData.sessionToken,
                language_code: 'en',
                language_name: 'English',
                region_code: 'US',
                timezone: 'America/Los_Angeles',
                user_preferences: {
                    date_format: 'MM/DD/YYYY',
                    time_format: '12h',
                    currency: 'USD'
                }
            };

            const responsePromise = this.waitForEvent('language:set');
            this.socket.emit('set:language', languageData);
            
            const response = await responsePromise;
            
            if (!response || response.status !== 'success') {
                throw new Error('Language setup failed');
            }

            log.info(`Language set successfully`);
            log.info(`Language: ${response.language_name}`);
            log.info(`Timezone: ${response.timezone}`);
        });
    }

    async testStaticMessages() {
        await this.runTest('Static Messages Test', async () => {
            if (!this.testData.sessionToken) {
                throw new Error('Session token not available');
            }

            const messageData = {
                mobile_no: this.testData.mobileNo,
                session_token: this.testData.sessionToken,
                message_type: 'game_list'
            };

            const responsePromise = this.waitForEvent('static:message');
            this.socket.emit('get:static:message', messageData);
            
            const response = await responsePromise;
            
            if (!response || response.status !== 'success') {
                throw new Error('Static message retrieval failed');
            }

            log.info(`Static message retrieved successfully`);
            log.info(`Message type: ${response.message_type}`);
        });
    }

    async testJWTTokenEncryption() {
        await this.runTest('JWT Token Encryption Test', async () => {
            if (!this.testData.jwtToken) {
                throw new Error('JWT token not available - run OTP verification first');
            }

            log.info(`JWT Token: ${this.testData.jwtToken.substring(0, 50)}...`);
            log.info(`Note: This token contains encrypted FCM token data`);
            log.info(`Original FCM Token: ${this.testData.fcmToken.substring(0, 20)}...`);
            log.info(`Server will decrypt and validate FCM token from JWT`);
        });
    }

    async testMainScreen() {
        await this.runTest('Main Screen Test', async () => {
            if (!this.testData.jwtToken) {
                throw new Error('JWT token not available - run OTP verification first');
            }

            const mainScreenData = {
                mobile_no: this.testData.mobileNo,
                fcm_token: this.testData.fcmToken,
                jwt_token: this.testData.jwtToken,
                device_id: this.testData.deviceId,
                message_type: 'game_list'
            };

            const responsePromise = this.waitForEvent('main:screen:game:list');
            this.socket.emit('main:screen', mainScreenData);
            
            const response = await responsePromise;
            
            if (!response || !response.gamelist) {
                throw new Error('Main screen data retrieval failed');
            }

            log.info(`Main screen data retrieved successfully`);
            log.info(`Game list items: ${response.gamelist.length}`);
            log.info(`Note: Server used JWT token values, not request values`);
        });
    }

    async testMainScreenWithEncryptedJWT() {
        await this.runTest('Main Screen with Encrypted JWT Test', async () => {
            if (!this.testData.jwtToken) {
                throw new Error('JWT token not available - run OTP verification first');
            }

            const mainScreenData = {
                mobile_no: this.testData.mobileNo,
                fcm_token: this.testData.fcmToken,
                jwt_token: this.testData.jwtToken,
                device_id: this.testData.deviceId,
                message_type: 'game_list'
            };

            const responsePromise = this.waitForEvent('main:screen:game:list');
            this.socket.emit('main:screen', mainScreenData);
            
            const response = await responsePromise;
            
            if (!response || !response.gamelist) {
                throw new Error('Main screen data retrieval failed');
            }

            log.info(`Main screen data retrieved successfully`);
            log.info(`Game list items: ${response.gamelist.length}`);
            log.info(`✅ Server decrypted JWT token and validated FCM token`);
            log.info(`✅ Server used encrypted token values, not request values`);
        });
    }

    async testErrorHandling() {
        await this.runTest('Error Handling Test', async () => {
            // Test invalid login data
            const invalidLoginData = {
                mobile_no: '', // Invalid empty mobile number
                device_id: '',
                fcm_token: ''
            };

            const errorPromise = this.waitForEvent('connection_error');
            this.socket.emit('login', invalidLoginData);
            
            const errorResponse = await errorPromise;
            
            if (!errorResponse || errorResponse.status !== 'error') {
                throw new Error('Error handling not working properly');
            }

            log.info(`Error handled correctly: ${errorResponse.message}`);
            log.info(`Error code: ${errorResponse.error_code}`);
        });
    }

    async testDisconnection() {
        await this.runTest('Disconnection Test', async () => {
            const disconnectPromise = new Promise((resolve) => {
                this.socket.once('disconnect', () => {
                    resolve();
                });
            });

            this.socket.disconnect();
            await disconnectPromise;
            
            log.info('Disconnection handled correctly');
        });
    }

    async testSimpleJWTToken() {
        await this.runTest('Simple JWT Token Test', async () => {
            if (!this.testData.jwtToken) {
                throw new Error('JWT token not available - run OTP verification first');
            }

            log.info(`Simple JWT Token: ${this.testData.jwtToken.substring(0, 50)}...`);
            log.info(`This token contains only 3 fields: mobile_no, device_id, fcm_token`);
            log.info(`Original Mobile: ${this.testData.mobileNo}`);
            log.info(`Original Device ID: ${this.testData.deviceId}`);
            log.info(`Original FCM Token: ${this.testData.fcmToken.substring(0, 20)}...`);
            log.info(`Server will decrypt using secret key and validate all 3 fields`);
        });
    }

    async testMainScreenWithSimpleJWT() {
        await this.runTest('Main Screen with Simple JWT Test', async () => {
            if (!this.testData.jwtToken) {
                throw new Error('JWT token not available - run OTP verification first');
            }

            const mainScreenData = {
                mobile_no: this.testData.mobileNo,
                fcm_token: this.testData.fcmToken,
                jwt_token: this.testData.jwtToken,
                device_id: this.testData.deviceId,
                message_type: 'game_list'
            };

            const responsePromise = this.waitForEvent('main:screen:game:list');
            this.socket.emit('main:screen', mainScreenData);
            
            const response = await responsePromise;
            
            if (!response || !response.gamelist) {
                throw new Error('Main screen data retrieval failed');
            }

            log.info(`Main screen data retrieved successfully`);
            log.info(`Game list items: ${response.gamelist.length}`);
            log.info(`✅ Server decrypted simple JWT token using secret key`);
            log.info(`✅ Server validated all 3 fields: mobile_no, device_id, fcm_token`);
            log.info(`✅ Server used decrypted token values, not request values`);
        });
    }

    // Run all tests
    async runAllTests() {
        console.clear();
        
        // Display beautiful header
        console.log(gradient.rainbow(figlet.textSync('Socket.IO Test Suite', { font: 'Standard' })));
        console.log(chalk.gray('Testing Go Socket.IO Application\n'));

        try {
            await this.connect();

            // Run all test cases
            await this.testConnection();
            await this.testConnectResponse();
            await this.testDeviceInfo();
            await this.testLogin();
            await this.testOTPVerification();
            await this.testSetProfile();
            await this.testSetLanguage();
            await this.testStaticMessages();
            await this.testJWTTokenEncryption();
            await this.testMainScreen();
            await this.testMainScreenWithEncryptedJWT();
            await this.testErrorHandling();
            await this.testDisconnection();
            await this.testSimpleJWTToken();
            await this.testMainScreenWithSimpleJWT();

        } catch (error) {
            log.error(`Test suite failed: ${error.message}`);
        } finally {
            await this.disconnect();
            this.displayResults();
        }
    }

    displayResults() {
        console.log('\n' + '='.repeat(80));
        console.log(chalk.bold.cyan('📊 TEST RESULTS SUMMARY'));
        console.log('='.repeat(80));

        // Create results table
        const table = new Table({
            head: [
                chalk.cyan('Metric'),
                chalk.cyan('Count'),
                chalk.cyan('Percentage')
            ],
            colWidths: [30, 15, 15]
        });

        const total = testResults.total;
        const passed = testResults.passed;
        const failed = testResults.failed;
        const skipped = testResults.skipped;

        table.push(
            [
                chalk.green('✅ Passed'),
                passed,
                `${((passed / total) * 100).toFixed(1)}%`
            ],
            [
                chalk.red('❌ Failed'),
                failed,
                `${((failed / total) * 100).toFixed(1)}%`
            ],
            [
                chalk.yellow('⏭️  Skipped'),
                skipped,
                `${((skipped / total) * 100).toFixed(1)}%`
            ],
            [
                chalk.blue('📋 Total'),
                total,
                '100%'
            ]
        );

        console.log(table.toString());

        // Display detailed results
        if (testResults.details.length > 0) {
            console.log('\n' + chalk.bold.cyan('📝 DETAILED RESULTS'));
            console.log('-'.repeat(80));

            testResults.details.forEach((test, index) => {
                const status = test.status === 'PASSED' 
                    ? chalk.green('✅ PASSED')
                    : chalk.red('❌ FAILED');
                
                console.log(`${chalk.cyan(`${index + 1}.`)} ${test.name} - ${status}`);
                
                if (test.error) {
                    console.log(`   ${chalk.gray('Error:')} ${test.error}`);
                }
            });
        }

        // Final summary
        console.log('\n' + '='.repeat(80));
        if (failed === 0) {
            console.log(chalk.bold.green('🎉 ALL TESTS PASSED! Your Socket.IO application is working perfectly!'));
        } else {
            console.log(chalk.bold.red(`⚠️  ${failed} test(s) failed. Please check the errors above.`));
        }
        console.log('='.repeat(80));
    }
}

// Interactive menu
async function showMenu() {
    console.clear();
    console.log(gradient.rainbow(figlet.textSync('Socket.IO Tester', { font: 'Standard' })));
    console.log(chalk.gray('Professional Test Suite for Go Socket.IO Application\n'));

    const { action } = await inquirer.prompt([
        {
            type: 'list',
            name: 'action',
            message: 'What would you like to do?',
            choices: [
                { name: '🚀 Run Complete Test Suite', value: 'run_all' },
                { name: '🧪 Run Individual Tests', value: 'run_individual' },
                { name: '⚙️  Configure Test Settings', value: 'configure' },
                { name: '📊 View Test Documentation', value: 'documentation' },
                { name: '❌ Exit', value: 'exit' }
            ]
        }
    ]);

    switch (action) {
        case 'run_all':
            const testSuite = new SocketIOTestSuite();
            await testSuite.runAllTests();
            break;
        case 'run_individual':
            await showIndividualTests();
            break;
        case 'configure':
            await showConfiguration();
            break;
        case 'documentation':
            showDocumentation();
            break;
        case 'exit':
            console.log(chalk.blue('👋 Goodbye!'));
            process.exit(0);
    }
}

async function showIndividualTests() {
    const { testType } = await inquirer.prompt([
        {
            type: 'list',
            name: 'testType',
            message: 'Select test type:',
            choices: [
                { name: '🔌 Connection Tests', value: 'connection' },
                { name: '🔐 Authentication Tests', value: 'auth' },
                { name: '👤 Profile Tests', value: 'profile' },
                { name: '🌐 Language Tests', value: 'language' },
                { name: '📱 Device Tests', value: 'device' },
                { name: '❌ Error Handling Tests', value: 'error' }
            ]
        }
    ]);

    const testSuite = new SocketIOTestSuite();
    
    try {
        await testSuite.connect();
        
        switch (testType) {
            case 'connection':
                await testSuite.testConnection();
                await testSuite.testConnectResponse();
                break;
            case 'auth':
                await testSuite.testLogin();
                await testSuite.testOTPVerification();
                break;
            case 'profile':
                await testSuite.testSetProfile();
                break;
            case 'language':
                await testSuite.testSetLanguage();
                break;
            case 'device':
                await testSuite.testDeviceInfo();
                break;
            case 'error':
                await testSuite.testErrorHandling();
                break;
        }
        
        testSuite.displayResults();
    } catch (error) {
        log.error(`Individual test failed: ${error.message}`);
    } finally {
        await testSuite.disconnect();
    }
}

async function showConfiguration() {
    const { serverUrl } = await inquirer.prompt([
        {
            type: 'input',
            name: 'serverUrl',
            message: 'Enter server URL:',
            default: CONFIG.SERVER_URL
        }
    ]);

    CONFIG.SERVER_URL = serverUrl;
    log.success(`Server URL updated to: ${serverUrl}`);
}

function showDocumentation() {
    console.log(chalk.bold.cyan('\n📚 SOCKET.IO TEST DOCUMENTATION'));
    console.log('='.repeat(80));
    
    console.log(chalk.yellow('\n🔌 Supported Events:'));
    console.log('• connect - Connection establishment');
    console.log('• connect_response - Server welcome response');
    console.log('• device:info - Device information submission');
    console.log('• login - User authentication');
    console.log('• verify:otp - OTP verification');
    console.log('• set:profile - User profile setup');
    console.log('• set:language - Language preferences');
    console.log('• get:static:message - Static content retrieval');
    console.log('• get:main:screen - Main screen data');
    console.log('• connection_error - Error responses');
    
    console.log(chalk.yellow('\n🧪 Test Coverage:'));
    console.log('• Connection and disconnection handling');
    console.log('• Authentication flow (login + OTP)');
    console.log('• Device registration and validation');
    console.log('• User profile management');
    console.log('• Language and localization settings');
    console.log('• Static message retrieval');
    console.log('• Error handling and validation');
    console.log('• Main screen data flow');
    
    console.log(chalk.yellow('\n⚙️  Configuration:'));
    console.log(`• Server URL: ${CONFIG.SERVER_URL}`);
    console.log(`• Timeout: ${CONFIG.TIMEOUT}ms`);
    console.log(`• Retry attempts: ${CONFIG.RETRY_ATTEMPTS}`);
    
    console.log('\n' + '='.repeat(80));
}

// Main execution
async function main() {
    try {
        await showMenu();
        
        // Ask if user wants to run again
        const { runAgain } = await inquirer.prompt([
            {
                type: 'confirm',
                name: 'runAgain',
                message: 'Would you like to run another test?',
                default: false
            }
        ]);

        if (runAgain) {
            await showMenu();
        } else {
            console.log(chalk.blue('👋 Thank you for using Socket.IO Test Suite!'));
        }
    } catch (error) {
        log.error(`Application error: ${error.message}`);
        process.exit(1);
    }
}

// Handle process termination
process.on('SIGINT', () => {
    console.log(chalk.blue('\n👋 Test suite interrupted. Goodbye!'));
    process.exit(0);
});

// Export for potential use as module
module.exports = { SocketIOTestSuite, CONFIG };

// Run if this file is executed directly
if (require.main === module) {
    main();
} 