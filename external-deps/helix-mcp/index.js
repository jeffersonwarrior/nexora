#!/usr/bin/env node
const { spawn } = require('child_process');
const readline = require('readline');

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

function execHelix(args) {
  return new Promise((resolve, reject) => {
    const helix = spawn('helix', args);
    let stdout = '';
    let stderr = '';

    helix.stdout.on('data', (data) => stdout += data);
    helix.stderr.on('data', (data) => stderr += data);

    helix.on('close', (code) => {
      if (code === 0) {
        resolve({ stdout, stderr });
      } else {
        reject(new Error(`helix exited with code ${code}: ${stderr}`));
      }
    });
  });
}

rl.on('line', async (line) => {
  try {
    const request = JSON.parse(line);

    if (request.method === 'initialize') {
      console.log(JSON.stringify({
        jsonrpc: '2.0',
        id: request.id,
        result: {
          protocolVersion: '2024-11-05',
          serverInfo: { name: 'helix-mcp', version: '1.0.0' },
          capabilities: { tools: {} }
        }
      }));
    } else if (request.method === 'tools/list') {
      console.log(JSON.stringify({
        jsonrpc: '2.0',
        id: request.id,
        result: {
          tools: [
            { name: 'helix_query', description: 'Execute HelixQL query', inputSchema: { type: 'object', properties: { query: { type: 'string' } }, required: ['query'] } },
            { name: 'helix_check', description: 'Validate helix schema', inputSchema: { type: 'object', properties: {} } },
            { name: 'helix_init', description: 'Initialize helix project', inputSchema: { type: 'object', properties: { path: { type: 'string' } }, required: ['path'] } }
          ]
        }
      }));
    } else if (request.method === 'tools/call') {
      const { name, arguments: args } = request.params;
      let result;

      if (name === 'helix_query') {
        result = await execHelix(['query', args.query]);
      } else if (name === 'helix_check') {
        result = await execHelix(['check']);
      } else if (name === 'helix_init') {
        result = await execHelix(['init', args.path]);
      }

      console.log(JSON.stringify({
        jsonrpc: '2.0',
        id: request.id,
        result: { content: [{ type: 'text', text: result.stdout || result.stderr }] }
      }));
    }
  } catch (err) {
    console.error(JSON.stringify({
      jsonrpc: '2.0',
      id: 1,
      error: { code: -32603, message: err.message }
    }));
  }
});
