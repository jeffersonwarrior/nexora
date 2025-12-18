#!/usr/bin/env node

const path = require('path');
const fs = require('fs');
const { spawn } = require('child_process');

// Get platform-specific binary path
const platform = process.platform;
const arch = process.arch;

let binaryName = 'nexora';
if (platform === 'win32') {
  binaryName = 'nexora.exe';
}

// Look for binary in package bin directory
const binDir = path.join(__dirname, 'bin');
const binaryPath = path.join(binDir, binaryName);

// Check if binary exists
if (!fs.existsSync(binaryPath)) {
  console.error('Error: Nexora binary not found. Please run npm install again.');
  process.exit(1);
}

// Spawn the binary
const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  cwd: process.cwd()
});

child.on('exit', (code) => {
  process.exit(code);
});

child.on('error', (err) => {
  console.error('Failed to start nexora:', err.message);
  process.exit(1);
});