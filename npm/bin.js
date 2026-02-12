#!/usr/bin/env node

const { spawn } = require('child_process');
const { join } = require('path');

const BINARY_NAME = 'misaki-banner';

function getBinaryPath() {
    const platform = process.platform;
    const binaryName = platform === 'win32' ? `${BINARY_NAME}.exe` : BINARY_NAME;
    return join(__dirname, binaryName);
}

function run() {
    const binaryPath = getBinaryPath();
    const args = process.argv.slice(2);

    const child = spawn(binaryPath, args, {
        stdio: 'inherit',
        windowsHide: false
    });

    child.on('exit', (code, signal) => {
        if (signal) {
            process.kill(process.pid, signal);
        } else {
            process.exit(code || 0);
        }
    });

    process.on('SIGINT', () => {
        child.kill('SIGINT');
        child.kill('SIGTERM');
    });
}

run();
