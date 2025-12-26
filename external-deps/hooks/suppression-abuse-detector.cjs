#!/usr/bin/env node
/**
 * Suppression Abuse Detector Hook
 * Prevents hiding issues via mass suppressions
 */

const fs = require('fs');

const SUPPRESSION_PATTERNS = [
  // Python
  { pattern: /#\s*noqa(?::\s*\w+)?/gi, lang: 'python' },
  { pattern: /#\s*type:\s*ignore/gi, lang: 'python' },
  { pattern: /#\s*pylint:\s*disable/gi, lang: 'python' },

  // JavaScript/TypeScript
  { pattern: /\/\/\s*eslint-disable(?:-next)?-line/gi, lang: 'js' },
  { pattern: /\/\*\s*eslint-disable\s*\*\//gi, lang: 'js' },
  { pattern: /\/\/\s*@ts-ignore/gi, lang: 'ts' },
  { pattern: /\/\/\s*@ts-expect-error/gi, lang: 'ts' },
  { pattern: /\/\/\s*@ts-nocheck/gi, lang: 'ts' },

  // Go
  { pattern: /\/\/\s*nolint(?::\s*\w+)?/gi, lang: 'go' },
  { pattern: /\/\/\s*#nosec/gi, lang: 'go' },

  // General
  { pattern: /NOLINT/gi, lang: 'any' },
  { pattern: /NOSONAR/gi, lang: 'any' },
];

const MAX_SUPPRESSIONS_PER_COMMIT = 5;
const MAX_SUPPRESSIONS_PER_FILE = 3;

function countSuppressions(content) {
  let total = 0;
  const matches = [];

  for (const { pattern, lang } of SUPPRESSION_PATTERNS) {
    const found = content.match(pattern) || [];
    total += found.length;
    if (found.length > 0) {
      matches.push({ lang, count: found.length, examples: found.slice(0, 3) });
    }
  }

  return { total, matches };
}

function checkFile(filepath) {
  if (!fs.existsSync(filepath)) {
    return null;
  }

  const content = fs.readFileSync(filepath, 'utf-8');
  const result = countSuppressions(content);

  return {
    file: filepath,
    ...result,
    exceeds: result.total > MAX_SUPPRESSIONS_PER_FILE,
  };
}

function checkDiff(diffContent) {
  // Count only ADDED suppressions (lines starting with +)
  const addedLines = diffContent
    .split('\n')
    .filter(line => line.startsWith('+') && !line.startsWith('+++'));

  const addedContent = addedLines.join('\n');
  return countSuppressions(addedContent);
}

// Main execution
if (require.main === module) {
  const args = process.argv.slice(2);
  const mode = args[0];

  if (mode === '--diff') {
    // Read diff from stdin
    let diffContent = '';
    process.stdin.setEncoding('utf8');
    process.stdin.on('data', chunk => diffContent += chunk);
    process.stdin.on('end', () => {
      const result = checkDiff(diffContent);

      if (result.total > MAX_SUPPRESSIONS_PER_COMMIT) {
        console.error(`\nSUPPRESSION ABUSE DETECTED`);
        console.error(`Found ${result.total} new suppressions (max: ${MAX_SUPPRESSIONS_PER_COMMIT})`);
        console.error(`\nEach suppression requires justification. Fix the issues instead.\n`);

        for (const m of result.matches) {
          console.error(`  ${m.lang}: ${m.count}x`);
          m.examples.forEach(e => console.error(`    - ${e}`));
        }

        process.exit(1);
      }

      if (result.total > 0) {
        console.log(`Note: ${result.total} suppressions added (within limit)`);
      }

      process.exit(0);
    });
  } else {
    // Check individual files
    const files = args;
    let totalExceeds = 0;

    for (const file of files) {
      const result = checkFile(file);
      if (result && result.exceeds) {
        console.error(`${file}: ${result.total} suppressions (max: ${MAX_SUPPRESSIONS_PER_FILE})`);
        totalExceeds++;
      }
    }

    process.exit(totalExceeds > 0 ? 1 : 0);
  }
}

module.exports = { countSuppressions, checkFile, checkDiff };
