#!/usr/bin/env node

const path = require('path');
const fs = require('fs');
const https = require('https');
const { execSync } = require('child_process');

// Package info
const packageName = 'nexora-cli';
const version = require('../package.json').version;

// Platform detection
const platform = process.platform;
const arch = process.arch;

let platformName, archName;

switch(platform) {
  case 'linux':
    platformName = 'Linux';
    break;
  case 'darwin':
    platformName = 'Darwin';
    break;
  case 'win32':
    platformName = 'Windows';
    break;
  case 'freebsd':
    platformName = 'FreeBSD';
    break;
  case 'openbsd':
    platformName = 'OpenBSD';
    break;
  case 'netbsd':
    platformName = 'NetBSD';
    break;
  default:
    console.error(`Unsupported platform: ${platform}`);
    process.exit(1);
}

switch(arch) {
  case 'x64':
    archName = 'x86_64';
    break;
  case 'arm64':
    archName = 'arm64';
    break;
  case 'ia32':
    archName = 'i386';
    break;
  default:
    console.error(`Unsupported architecture: ${arch}`);
    process.exit(1);
}

// Binary name
const binaryName = platform === 'win32' ? 'nexora.exe' : 'nexora';

// Download URL (you'll need to adjust this based on your release structure)
const downloadUrl = `https://github.com/nexora/nexora/releases/download/v${version}/nexora_${version}_${platformName}_${archName}.tar.gz`;

// Install directory
const binDir = path.join(__dirname, '..', 'bin');
const tempFile = path.join(binDir, 'nexora.tar.gz');

// Create bin directory if it doesn't exist
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

// Download and extract
console.log(`Downloading nexora ${version} for ${platformName} ${archName}...`);

function download(url, dest, callback) {
  const file = fs.createWriteStream(dest);
  https.get(url, (response) => {
    if (response.statusCode !== 200) {
      callback(new Error(`Failed to download: ${response.statusCode}`));
      return;
    }
    response.pipe(file);
    file.on('finish', () => {
      file.close(callback);
    });
  }).on('error', (err) => {
    fs.unlink(dest, () => {
      callback(err);
    });
  });
}

download(downloadUrl, tempFile, (err) => {
  if (err) {
    console.error('Download failed:', err.message);
    console.log('You may need to install nexora manually from https://github.com/nexora/nexora/releases');
    process.exit(1);
  }
  
  // Extract the binary
  try {
    const tar = require('tar');
    tar.extract({
      file: tempFile,
      cwd: binDir,
      strip: 1
    }, () => {
      // Remove the downloaded archive
      fs.unlinkSync(tempFile);
      
      // Make binary executable on Unix systems
      if (platform !== 'win32') {
        fs.chmodSync(path.join(binDir, binaryName), '755');
      }
      
      console.log('Nexora installed successfully!');
    });
  } catch (error) {
    // If tar module not available, try system tar
    try {
      execSync(`tar -xzf "${tempFile}" -C "${binDir}" --strip-components=1`, { stdio: 'inherit' });
      fs.unlinkSync(tempFile);
      
      if (platform !== 'win32') {
        execSync(`chmod +x "${path.join(binDir, binaryName)}"`);
      }
      
      console.log('Nexora installed successfully!');
    } catch (tarErr) {
      console.error('Extraction failed:', tarErr.message);
      fs.unlinkSync(tempFile);
      process.exit(1);
    }
  }
});