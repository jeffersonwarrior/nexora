#!/usr/bin/env node
/**
 * Antipattern Detector Hook
 * Catches duty-shirking patterns in code changes
 */

const fs = require('fs');
const path = require('path');

const ANTIPATTERNS = [
  // Stub implementations
  {
    pattern: /TODO.*\n\s*pass\s*$/m,
    name: 'stub-without-error',
    message: 'TODO with pass - use NotImplementedError instead',
    severity: 'error',
  },
  {
    pattern: /raise\s+NotImplementedError\s*\(\s*\)/,
    name: 'empty-not-implemented',
    message: 'NotImplementedError should include description',
    severity: 'warning',
  },

  // CI weakening
  {
    pattern: /continue-on-error:\s*true(?!\s*#\s*justified)/,
    name: 'ci-weakening',
    message: 'continue-on-error without justification comment',
    severity: 'error',
  },

  // Test weakening
  {
    pattern: /assert\s+True\s*(?:#.*)?$/m,
    name: 'assert-true',
    message: 'assert True - tests should verify actual conditions',
    severity: 'error',
  },
  {
    pattern: /pytest\.skip\(\s*\)(?!\s*#\s*\w)/,
    name: 'unconditional-skip',
    message: 'Unconditional pytest.skip without reason',
    severity: 'error',
  },
  {
    pattern: /\.skip\(['"]\s*['"]\)/,
    name: 'empty-skip-reason',
    message: 'Empty skip reason',
    severity: 'warning',
  },

  // Coverage reduction
  {
    pattern: /fail_under\s*[=:]\s*[0-4]\d(?:\.\d+)?/,
    name: 'low-coverage-threshold',
    message: 'Coverage threshold below 50% is suspicious',
    severity: 'warning',
  },

  // Go-specific antipatterns
  {
    pattern: /\/\/\s*TODO:?\s*implement/i,
    name: 'todo-implement',
    message: 'TODO implement marker - complete the implementation',
    severity: 'warning',
  },
  {
    pattern: /panic\s*\(\s*["']not implemented["']\s*\)/,
    name: 'panic-not-implemented',
    message: 'panic("not implemented") - complete the implementation',
    severity: 'error',
  },

  // Error swallowing
  {
    pattern: /catch\s*\([^)]*\)\s*\{\s*\}/,
    name: 'empty-catch',
    message: 'Empty catch block swallows errors',
    severity: 'error',
  },
  {
    pattern: /if\s+err\s*!=\s*nil\s*\{\s*\}/,
    name: 'empty-error-handler',
    message: 'Empty error handler in Go',
    severity: 'error',
  },
];

function checkContent(content, filename) {
  const issues = [];

  for (const { pattern, name, message, severity } of ANTIPATTERNS) {
    const matches = content.match(new RegExp(pattern, 'gm'));
    if (matches) {
      issues.push({
        name,
        message,
        severity,
        count: matches.length,
        filename,
      });
    }
  }

  return issues;
}

function formatIssues(issues) {
  if (issues.length === 0) return '';

  const errors = issues.filter(i => i.severity === 'error');
  const warnings = issues.filter(i => i.severity === 'warning');

  let output = '';

  if (errors.length > 0) {
    output += '\nERRORS:\n';
    for (const e of errors) {
      output += `  [${e.name}] ${e.message} (${e.count}x in ${e.filename})\n`;
    }
  }

  if (warnings.length > 0) {
    output += '\nWARNINGS:\n';
    for (const w of warnings) {
      output += `  [${w.name}] ${w.message} (${w.count}x in ${w.filename})\n`;
    }
  }

  return output;
}

// Main execution
if (require.main === module) {
  const files = process.argv.slice(2);

  if (files.length === 0) {
    console.log('Usage: antipattern-detector.cjs <file1> [file2] ...');
    process.exit(0);
  }

  let allIssues = [];

  for (const file of files) {
    if (!fs.existsSync(file)) continue;

    const content = fs.readFileSync(file, 'utf-8');
    const issues = checkContent(content, file);
    allIssues = allIssues.concat(issues);
  }

  const output = formatIssues(allIssues);

  if (output) {
    console.error(output);
    const hasErrors = allIssues.some(i => i.severity === 'error');
    process.exit(hasErrors ? 1 : 0);
  }

  process.exit(0);
}

module.exports = { checkContent, ANTIPATTERNS };
