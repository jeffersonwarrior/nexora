#!/usr/bin/env node
/**
 * Bash Protection Hook
 * Blocks dangerous commands before execution
 */

const fs = require('fs');
const path = require('path');

const BLOCKED_PATTERNS = [
  // Destructive rm commands
  /rm\s+(-[rRfF]+\s+)*[~\/](?!\S*tmp)/,  // rm -rf ~/ or similar
  /rm\s+(-[rRfF]+\s+)*\$HOME/,
  /rm\s+(-[rRfF]+\s+)*\/(?:usr|etc|var|bin|sbin|lib|opt)/,

  // Dangerous git commands
  /git\s+push\s+.*--force.*(?:main|master)/,
  /git\s+push\s+-f\s+.*(?:main|master)/,
  /git\s+reset\s+--hard\s+(?:origin\/)?(?:main|master)/,

  // Database destruction
  /DROP\s+(?:DATABASE|TABLE|SCHEMA)/i,
  /TRUNCATE\s+TABLE/i,
  /DELETE\s+FROM\s+\w+\s*;?\s*$/i, // DELETE without WHERE

  // System-level dangers
  /chmod\s+777\s+\/|chmod\s+-R\s+777/,
  /chown\s+-R\s+.*\//,

  // Process killers (protected)
  /pkill\s+.*claude/i,
  /kill\s+-9\s+.*claude/i,
];

const BLOCKED_COMMANDS = [
  'rm -rf /',
  'rm -rf ~',
  'rm -rf $HOME',
  ':(){:|:&};:',  // Fork bomb
];

function checkCommand(command) {
  const cmd = command.trim();

  // Check exact blocked commands
  for (const blocked of BLOCKED_COMMANDS) {
    if (cmd.includes(blocked)) {
      return { blocked: true, reason: `Exact match: "${blocked}"` };
    }
  }

  // Check pattern matches
  for (const pattern of BLOCKED_PATTERNS) {
    if (pattern.test(cmd)) {
      return { blocked: true, reason: `Pattern match: ${pattern.source}` };
    }
  }

  return { blocked: false };
}

function logBlocked(command, reason) {
  const logDir = path.join(__dirname, '..', 'logs');
  const logFile = path.join(logDir, 'blocked_commands.log');

  try {
    if (!fs.existsSync(logDir)) {
      fs.mkdirSync(logDir, { recursive: true });
    }

    const entry = `[${new Date().toISOString()}] BLOCKED: ${command}\n  Reason: ${reason}\n\n`;
    fs.appendFileSync(logFile, entry);
  } catch (e) {
    console.error('Failed to log blocked command:', e.message);
  }
}

// Main execution
if (require.main === module) {
  const command = process.argv.slice(2).join(' ') || process.env.CLAUDE_COMMAND || '';

  if (!command) {
    process.exit(0);
  }

  const result = checkCommand(command);

  if (result.blocked) {
    logBlocked(command, result.reason);
    console.error(`BLOCKED: ${result.reason}`);
    process.exit(1);
  }

  process.exit(0);
}

module.exports = { checkCommand, BLOCKED_PATTERNS, BLOCKED_COMMANDS };
