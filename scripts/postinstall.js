#!/usr/bin/env node

const path = require('path');
const fs = require('fs');

// Check if binary exists and is executable
const binDir = path.join(__dirname, '..', 'bin');
const binaryName = process.platform === 'win32' ? 'nexora.exe' : 'nexora';
const binaryPath = path.join(binDir, binaryName);

if (fs.existsSync(binaryPath)) {
  console.log('✓ Nexora CLI is ready to use!');
  console.log('  Run "nexora" to start.');
} else {
  console.log('⚠️  Nexora binary not found. Installing...');
  
  // Try to run the installation script
  try {
    require('./install.js');
  } catch (error) {
    console.error('❌ Installation failed. Please install manually from:');
    console.log('   https://github.com/nexora/cli/releases');
  }
}