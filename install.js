#!/usr/bin/env node

const { execSync } = require('child_process');
const chalk = require('chalk');
const ora = require('ora');
const gradient = require('gradient-string');
const figlet = require('figlet');

console.log(gradient.rainbow(figlet.textSync('Socket.IO Test Suite', { font: 'Standard' })));
console.log(chalk.gray('Professional Test Suite Installation\n'));

const spinner = ora('Checking prerequisites...').start();

// Check Node.js version
try {
    const nodeVersion = process.version;
    const majorVersion = parseInt(nodeVersion.slice(1).split('.')[0]);
    
    if (majorVersion < 16) {
        spinner.fail('Node.js version 16.0.0 or higher is required');
        console.log(chalk.yellow(`Current version: ${nodeVersion}`));
        console.log(chalk.blue('Please upgrade Node.js and try again'));
        process.exit(1);
    }
    
    spinner.succeed(`Node.js version check passed: ${nodeVersion}`);
} catch (error) {
    spinner.fail('Failed to check Node.js version');
    process.exit(1);
}

// Install dependencies
const installSpinner = ora('Installing dependencies...').start();

try {
    execSync('npm install', { stdio: 'pipe' });
    installSpinner.succeed('Dependencies installed successfully');
} catch (error) {
    installSpinner.fail('Failed to install dependencies');
    console.log(chalk.red('Error:'), error.message);
    process.exit(1);
}

// Check if Go server is running
const serverSpinner = ora('Checking Go server connection...').start();

try {
    const { io } = require('socket.io-client');
    const socket = io('http://localhost:8088', { timeout: 5000 });
    
    socket.on('connect', () => {
        serverSpinner.succeed('Go Socket.IO server is running');
        socket.disconnect();
        showSuccessMessage();
    });
    
    socket.on('connect_error', () => {
        serverSpinner.warn('Go Socket.IO server is not running');
        showServerInstructions();
    });
    
    setTimeout(() => {
        serverSpinner.warn('Connection timeout - server may not be running');
        showServerInstructions();
    }, 5000);
    
} catch (error) {
    serverSpinner.fail('Failed to check server connection');
    showServerInstructions();
}

function showSuccessMessage() {
    console.log('\n' + '='.repeat(80));
    console.log(chalk.bold.green('🎉 Installation completed successfully!'));
    console.log('='.repeat(80));
    
    console.log(chalk.cyan('\n🚀 Quick Start:'));
    console.log(chalk.white('  npm start          - Launch interactive test suite'));
    console.log(chalk.white('  npm test           - Run complete test suite'));
    console.log(chalk.white('  npm run test:quick - Run quick tests'));
    
    console.log(chalk.cyan('\n📚 Documentation:'));
    console.log(chalk.white('  README.md          - Complete documentation'));
    console.log(chalk.white('  TEST_SUITE_README.md - Test suite guide'));
    
    console.log(chalk.cyan('\n🔧 Configuration:'));
    console.log(chalk.white('  Edit socket_test_suite.js to modify test settings'));
    console.log(chalk.white('  Server URL: http://localhost:8088'));
    
    console.log('\n' + '='.repeat(80));
    console.log(chalk.gray('Happy testing! 🧪'));
}

function showServerInstructions() {
    console.log('\n' + '='.repeat(80));
    console.log(chalk.bold.yellow('⚠️  Go Server Not Running'));
    console.log('='.repeat(80));
    
    console.log(chalk.cyan('\n📋 To start your Go Socket.IO server:'));
    console.log(chalk.white('  1. Navigate to your Go project directory'));
    console.log(chalk.white('  2. Run: go run main.go'));
    console.log(chalk.white('  3. Ensure MongoDB is running'));
    console.log(chalk.white('  4. Server should start on port 8088'));
    
    console.log(chalk.cyan('\n🔍 Verify server is running:'));
    console.log(chalk.white('  curl http://localhost:8088/socket.io/'));
    
    console.log(chalk.cyan('\n✅ Once server is running:'));
    console.log(chalk.white('  npm start          - Launch test suite'));
    console.log(chalk.white('  npm test           - Run tests'));
    
    console.log('\n' + '='.repeat(80));
    console.log(chalk.gray('Installation completed. Start your Go server to begin testing! 🚀'));
} 